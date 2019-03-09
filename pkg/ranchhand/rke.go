package ranchhand

import (
	"context"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

var tpl *template.Template

const (
	RKEKubeConfig = "kube_config_rancher-cluster.yml"
	RKEConfigFile = "rancher-cluster.yml"
	RKETemplate   = `# DO NOT EDIT THIS FILE - GENERATED BY RANCHHAND
ssh_key_path: {{ .SSHKeyPath }}
ignore_docker_version: false

nodes:
{{- range $index, $element := .Nodes }}
  - address: {{ . }}
    user: {{ $.SSHUser }}
    port: {{ $.SSHPort }}
    role: [controlplane,worker,etcd]
    {{- with (index $.NodeInternalIPs $index) }}
    internal_address: {{ . }}
    {{- end }}
{{- end }}

{{- if .SSHProxyHost }}
bastion_host:
    address: {{ .SSHProxyHost }}
    user: {{ .SSHProxyUser }}
    port: {{ .SSHPort }}
    ssh_key_path: {{ .SSHKeyPath }}
{{- end }}

services:
  etcd:
    snapshot: true
    creation: 6h
    retention: 24h
`
)

func installKubernetes(cfg *Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := exec.CommandContext(ctx, "rke", "version", "--config", RKEConfigFile).Run(); err == nil {
		return nil
	}

	// generate rke config
	file, err := os.Create(RKEConfigFile)
	if err != nil {
		return errors.Wrapf(err, "cannot create %s", RKEConfigFile)
	}
	defer file.Close()

	if err := tpl.Execute(file, cfg); err != nil {
		return errors.Wrap(err, "rke template render failed")
	}

	// execute rke up
	cmd := exec.Command("rke", "up", "--config", RKEConfigFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "cannot install kubernetes")
	}

	// todo: add cluster-ready check

	return nil
}

func init() {
	tpl = template.Must(template.New("rke-tmpl").Parse(RKETemplate))
}
