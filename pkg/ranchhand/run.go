package ranchhand

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

const OutputDirectory = "rh-output"

type Config struct {
	SSHUser    string
	SSHPort    uint
	SSHKeyPath string
	Nodes      []string
	Timeout    time.Duration
}

// required steps:
// todo: ensure the k8s cluster came up and is healthy
//
// desired steps:
// 	- add error ctx via wrapping
// 	- add logging
// 	- write tests
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
	if err := installTiller(); err != nil {
		return err
	}

	log.Info("installing rancher")
	return installRancher(cfg.Nodes[0])
}
