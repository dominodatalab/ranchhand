package cmd

import (
	"log"

	"github.com/cerebrotech/ranchhand/pkg/ranchhand"
	"github.com/spf13/cobra"
)

var (
	nodeIPs    []string
	sshUser    string
	sshPort    uint8
	sshKeyPath string

	rootCmd = &cobra.Command{
		Use:   "ranchhand",
		Short: "Create a Rancher HA installation",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := ranchhand.Config{
				Nodes:      nodeIPs,
				SSHUser:    sshUser,
				SSHPort:    sshPort,
				SSHKeyPath: sshKeyPath,
			}
			if err := ranchhand.Run(&cfg); err != nil {
				log.Fatalln(err)
			}
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func init() {
	rootCmd.Flags().StringSliceVarP(&nodeIPs, "node-ips", "n", []string{}, "List of remote hosts")
	rootCmd.Flags().StringVarP(&sshUser, "ssh-user", "u", "root", "User used to remote host")
	rootCmd.Flags().Uint8VarP(&sshPort, "ssh-port", "p", 22, "Port to connect to on the remote host")
	rootCmd.Flags().StringVarP(&sshKeyPath, "ssh-key-path", "k", "", "Path to private key")

	rootCmd.MarkFlagRequired("node-ips")
	rootCmd.MarkFlagRequired("ssh-key-path")
}
