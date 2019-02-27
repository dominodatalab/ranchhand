package ranchhand

type Config struct {
	SSHUser    string
	SSHPort    uint8
	SSHKeyPath string
	Nodes      []string
}

// required steps:
// 	- execute rke up
// 	- ensure the k8s cluster came up and is healthy
// 	- install tiller
//  - deploy rancher into k8s via helm
//  - use rancher api to check server health
// 	- check error output
// 	- add logging
// 	- write tests
func Run(cfg *Config) error {
	if err := installRequiredTools(); err != nil {
		return err
	}

	if err := processHosts(cfg); err != nil {
		return err
	}

	if err := installKubernetes(cfg); err != nil {
		return err
	}

	return nil
}
