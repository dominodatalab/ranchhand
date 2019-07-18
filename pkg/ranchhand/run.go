package ranchhand

import (
	"os"
	"time"

	"github.com/dominodatalab/ranchhand/pkg/helm"
	"github.com/sirupsen/logrus"
)

const OutputDirectory = "ranchhand-output"

var log = logrus.StandardLogger()

type Node struct {
	PublicIP  string
	PrivateIP string
}

type SSHConfig struct {
	User              string
	Port              uint
	KeyPath           string
	ConnectionTimeout uint
}

type Config struct {
	SSH           *SSHConfig
	Nodes         []Node
	Timeout       time.Duration
	CertIPs       []string
	CertDNSNames  []string
	AdminPassword string

	UpgradeRancher    bool
	UpgradeKubernetes bool
}

func Run(cfg *Config) error {
	log.Infof("ensuring output directory [%s] exists", OutputDirectory)
	if err := ensureDirectory(OutputDirectory); err != nil {
		return err
	}
	if err := os.Chdir(OutputDirectory); err != nil {
		return err
	}

	log.Info("installing required cli tools")
	if err := installRequiredTools(); err != nil {
		return err
	}

	log.Info("processing remote hosts")
	if err := processHosts(cfg); err != nil {
		return err
	}

	log.Info("generating ingress certificate")
	certPEM, keyPEM, err := generateCertificate(cfg)
	if err != nil {
		return err
	}

	log.Info("installing kubernetes via rke")
	if err := launchRKE(cfg, certPEM, keyPEM); err != nil {
		return err
	}

	log.Info("initializing helm/tiller")
	hClient, err := helm.New(".helm", RKEKubeConfig)
	if err != nil {
		return err
	}
	if err := hClient.Init(); err != nil {
		return err
	}

	log.Info("creating rancher ca cert secret")
	if err := createRancherSecret(certPEM, RKEKubeConfig); err != nil {
		return err
	}

	nodeIP := cfg.Nodes[0].PublicIP

	log.Info("deploying rancher application")
	if err := installRancher(hClient, nodeIP, cfg.UpgradeRancher); err != nil {
		return err
	}

	log.Info("checking rancher admin password")
	return modifyRancherAdminPassword(nodeIP, cfg.AdminPassword)

}

func init() {
	log.SetOutput(os.Stdout)
}
