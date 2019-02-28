package ranchhand

import "os"

const OutputDirectory = "rh-output"

type Config struct {
	SSHUser    string
	SSHPort    uint8
	SSHKeyPath string
	Nodes      []string
}

// required steps:
// 	- ensure the k8s cluster came up and is healthy
//  - deploy rancher into k8s via helm
//  - use rancher api to check server health
// desired steps:
// 	- add error ctx via wrapping
// 	- add logging
// 	- write tests
func Run(cfg *Config) error {
	if err := ensureDirectory(OutputDirectory); err != nil {
		return err
	}

	if err := os.Chdir(OutputDirectory); err != nil {
		return err
	}

	if err := installRequiredTools(); err != nil {
		return err
	}

	if err := processHosts(cfg); err != nil {
		return err
	}

	if err := installKubernetes(cfg); err != nil {
		return err
	}

	if err := installTiller(); err != nil {
		return err
	}

	return installRancher()
}

func ensureDirectory(dir string) error {
	if _, serr := os.Stat(dir); os.IsNotExist(serr) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
