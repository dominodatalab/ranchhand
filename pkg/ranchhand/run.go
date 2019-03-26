package ranchhand

import (
	"os"
	"time"

	"github.com/dominodatalab/ranchhand/pkg/helm"
	log "github.com/sirupsen/logrus"
)

const OutputDirectory = "ranchhand-output"

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
	SSH     *SSHConfig
	Nodes   []Node
	Timeout time.Duration
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
	return installRancher(hClient, cfg.Nodes[0].PublicIP)
}

func init() {
	log.SetOutput(os.Stdout)
}
