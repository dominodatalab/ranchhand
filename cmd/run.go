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
  ranchhand -n "54.78.22.1:10.100.2.2, 77.13.122.9:10.100.2.5""
`

var (
	nodeIPs    []string
	sshUser    string
	sshPort    uint
	sshKeyPath string
	timeout    uint

	runCmd = &cobra.Command{
		Use:     "run",
		Short:   "Create a Rancher HA installation",
		Example: strings.TrimLeft(runExamples, "\n"),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := ranchhand.Config{
				Nodes:      ranchhand.BuildNodes(nodeIPs),
				SSHUser:    sshUser,
				SSHPort:    sshPort,
				SSHKeyPath: sshKeyPath,
				Timeout:    time.Duration(timeout) * time.Second,
			}

			if err := ranchhand.Run(&cfg); err != nil {
				log.Fatalln(err)
			}
		},
	}
)

func init() {
	runCmd.Flags().StringSliceVarP(&nodeIPs, "node-ips", "n", []string{}, "List of remote hosts (comma-delimited)")
	runCmd.Flags().StringVarP(&sshUser, "ssh-user", "u", "root", "User used to remote host")
	runCmd.Flags().UintVarP(&sshPort, "ssh-port", "p", 22, "Port to connect to on the remote host")
	runCmd.Flags().StringVarP(&sshKeyPath, "ssh-key-path", "i", "", "Path to private key")
	runCmd.Flags().UintVarP(&timeout, "timeout", "t", 300, "The duration (in seconds) RanchHand will wait process all hosts")

	runCmd.MarkFlagRequired("node-ips")
	runCmd.MarkFlagRequired("ssh-key-path")

	rootCmd.AddCommand(runCmd)
}
