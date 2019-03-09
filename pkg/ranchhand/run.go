package ranchhand

import (
	"os"
	"time"

	"github.com/dominodatalab/ranchhand/pkg/helm"
	log "github.com/sirupsen/logrus"
)

const OutputDirectory = "ranchhand-output"

type Config struct {
	SSHUser          string
	SSHPort          uint
        SSHProxyHost     string
        SSHProxyUser     string
	SSHKeyPath       string
	Nodes           []string
	Timeout         time.Duration
}

// required steps:
// todo: ensure the k8s cluster came up and is healthy
func Run(cfg *Config) error {
	log.Info("ensuring output directory exists")
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

	log.Info("installing kubernetes")
	if err := installKubernetes(cfg); err != nil {
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

	log.Info("deploying rancher application")
	return installRancher(hClient, cfg.Nodes[0])
}
