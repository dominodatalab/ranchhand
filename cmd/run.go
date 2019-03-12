package cmd

import (
	"log"
	"time"

	"github.com/dominodatalab/ranchhand/pkg/ranchhand"
	"github.com/spf13/cobra"
)

var (
	nodeIPs     []string
	internalIPs []string
	sshUser     string
	sshPort     uint
	sshKeyPath  string
	timeout     uint

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Create a Rancher HA installation",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := ranchhand.Config{
				Nodes:           nodeIPs,
				NodeInternalIPs: internalIPs,
				SSHUser:         sshUser,
				SSHPort:         sshPort,
				SSHKeyPath:      sshKeyPath,
				Timeout:         time.Duration(timeout) * time.Second,
			}
			if err := ranchhand.Run(&cfg); err != nil {
				log.Fatalln(err)
			}
		},
	}
)

func init() {
	runCmd.Flags().StringSliceVarP(&nodeIPs, "node-ips", "n", []string{}, "Comma-delimited list of remote hosts")
	runCmd.Flags().StringSliceVarP(&internalIPs, "internal-ips", "", []string{}, "Comma-delimited list of hosts' private ips (optional)")
	runCmd.Flags().StringVarP(&sshUser, "ssh-user", "u", "root", "User used to remote host")
	runCmd.Flags().UintVarP(&sshPort, "ssh-port", "p", 22, "Port to connect to on the remote host")
	runCmd.Flags().StringVarP(&sshKeyPath, "ssh-key-path", "i", "", "Path to private key")
	runCmd.Flags().UintVarP(&timeout, "timeout", "t", 300, "The duration (in seconds) RanchHand will wait process all hosts")

	runCmd.MarkFlagRequired("node-ips")
	runCmd.MarkFlagRequired("ssh-key-path")

	rootCmd.AddCommand(runCmd)
}
