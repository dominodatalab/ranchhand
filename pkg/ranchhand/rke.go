package ranchhand

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	RKEKubeConfig = "kube_config_rancher-cluster.yml"

	RKEConfigFile = "rancher-cluster.yml"

	RKETemplate = `# DO NOT EDIT THIS FILE - GENERATED BY RANCHHAND
ssh_key_path: {{ .SSH.KeyPath }}
ignore_docker_version: false

nodes:
{{- range .Nodes }}
  - address: {{ .PublicIP }}
  {{- with .PrivateIP }}
    internal_address: {{ . }}
  {{- end }}
    user: {{ $.SSH.User }}
    port: {{ $.SSH.Port }}
    role: [controlplane,worker,etcd]
{{- end }}

services:
  etcd:
    snapshot: true
    creation: 6h
    retention: 24h
  kubelet:
    extra_args:
      streaming-connection-idle-timeout: "30m"
      protect-kernel-defaults: "true"
      make-iptables-util-chains: "true"
      event-qps: "0"

ingress:
  provider: nginx
  extra_args:
    default-ssl-certificate: ingress-nginx/ingress-default-cert

addons: |-
  ---
  apiVersion: v1
  kind: Secret
  metadata:
    name: ingress-default-cert
    namespace: ingress-nginx
  type: kubernetes.io/tls
  data:
    tls.crt: {{ .CertPEM | base64Encode }}
    tls.key: {{ .KeyPEM | base64Encode }}
`
)

var tpl *template.Template

type tmplData struct {
	*Config
	CertPEM, KeyPEM []byte
}

func launchRKE(cfg *Config, certPEM, keyPEM []byte) error {
	var buf bytes.Buffer
	tplData := tmplData{
		Config:  cfg,
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	}
	if err := tpl.Execute(&buf, tplData); err != nil {
		return errors.Wrap(err, "rke template render failed")
	}
	tplContents := buf.Bytes()

	if _, err := os.Stat(RKEConfigFile); os.IsNotExist(err) {
		return configAndRunCLI(tplContents)
	}

	fileContents, err := ioutil.ReadFile(RKEConfigFile)
	if err != nil {
		return errors.Wrap(err, "rke config read failed")
	}
	if !bytes.Equal(fileContents, tplContents) {
		log.Info("rewriting rke config because of template change")
		return configAndRunCLI(tplContents)
	}

	return nil
}

func configAndRunCLI(contents []byte) error {
	if err := ioutil.WriteFile(RKEConfigFile, contents, 0644); err != nil {
		return errors.Wrap(err, "rke config write failed")
	}

	cmd := exec.Command("rke", "up", "--config", RKEConfigFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Info("launching rke up")
	return errors.Wrap(cmd.Run(), "cannot install kubernetes")
}

func init() {
	tpl = template.New("rke-tmpl")
	tpl.Funcs(template.FuncMap{
		"base64Encode": func(bs []byte) string {
			return base64.StdEncoding.EncodeToString(bs)
		},
	})
	template.Must(tpl.Parse(RKETemplate))
}
