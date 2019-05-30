package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/dominodatalab/ranchhand/pkg/ranchhand"
	"github.com/spf13/cobra"
)

const runExamples = `
  # Single node cluster
  ranchhand -n 54.78.22.1

  # Multi-node cluster
  ranchhand -n "54.78.22.1, 77.13.122.9"

  # Cluster with nodes that need to use private IPs for internal communication
  ranchhand -n "54.78.22.1:10.100.2.2, 77.13.122.9:10.100.2.5"
`

var (
	nodeIPs       []string
	sshUser       string
	sshPort       uint
	sshKeyPath    string
	sshTimeout    uint
	timeout       uint
	certIPs       []string
	certDNSNames  []string
	adminPassword string

	runCmd = &cobra.Command{
		Use:     "run",
		Short:   "Create a Rancher HA installation",
		Example: strings.TrimLeft(runExamples, "\n"),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := ranchhand.Config{
				SSH: &ranchhand.SSHConfig{
					User:              sshUser,
					Port:              sshPort,
					KeyPath:           sshKeyPath,
					ConnectionTimeout: sshTimeout,
				},
				Nodes:         ranchhand.BuildNodes(nodeIPs),
				Timeout:       time.Duration(timeout) * time.Second,
				CertIPs:       certIPs,
				CertDNSNames:  certDNSNames,
				AdminPassword: adminPassword,
			}

			if err := ranchhand.Run(&cfg); err != nil {
				log.Fatalln(err)
			}
		},
	}
)

func init() {
	runCmd.Flags().SortFlags = false

	runCmd.Flags().StringSliceVarP(&nodeIPs, "node-ips", "n", []string{}, "list of hosts (comma-delimited)")
	runCmd.Flags().StringVarP(&sshUser, "ssh-user", "u", "root", "host ssh username")
	runCmd.Flags().UintVarP(&sshPort, "ssh-port", "p", 22, "host ssh port")
	runCmd.Flags().StringVarP(&sshKeyPath, "ssh-key-path", "i", "", "path to private ssh key")
	runCmd.Flags().UintVarP(&sshTimeout, "ssh-connect-timeout", "c", 30, "time to wait (in secs) for hosts to accept connection")
	runCmd.Flags().StringSliceVarP(&certIPs, "cert-ips", "a", []string{"127.0.0.1"}, "list of ip addresses in ca cert (comma-delimited)")
	runCmd.Flags().StringSliceVarP(&certDNSNames, "cert-dns-names", "d", []string{"rancher.example.org"}, "list of dns names in ca cert (comma-delimited, first is CN)")
	runCmd.Flags().StringVarP(&adminPassword, "admin-password", "r", "", "change default admin password")
	runCmd.Flags().UintVarP(&timeout, "timeout", "t", 300, "total time to wait (in secs) for host processing to complete")

	runCmd.MarkFlagRequired("node-ips")
	runCmd.MarkFlagRequired("ssh-key-path")

	rootCmd.AddCommand(runCmd)
}
