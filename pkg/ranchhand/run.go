package ranchhand

// required steps:
// 	- enforce node requirements
// 	- generate rke configuration
// 	- execute rke up
// 	- ensure the k8s cluster came up and is healthy
// 	- install tiller
//  - deploy rancher into k8s via helm
//  - use rancher api to check server health
func Run(ips []string, sshKey string) error {
	return nil
}
