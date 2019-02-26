package ranchhand

// required steps:
// 	- generate rke configuration
// 	- execute rke up
// 	- ensure the k8s cluster came up and is healthy
// 	- install tiller
//  - deploy rancher into k8s via helm
//  - use rancher api to check server health
func Run(hosts []string, sshKeyPath string) error {
	if err := enforceRequirements(hosts, sshKeyPath); err != nil {
		return err
	}

	return nil
}

