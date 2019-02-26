package cmd

import (
	"log"

	"github.com/cerebrotech/ranchhand/pkg/ranchhand"
	"github.com/spf13/cobra"
)

var (
	nodeIPs    []string
	sshKeyPath string

	rootCmd = &cobra.Command{
		Use: "ranchhand",
		Short: "Create a Rancher HA installation",
		Run: func(cmd *cobra.Command, args []string) {
			ranchhand.Run(nodeIPs, sshKeyPath)
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func init() {
	rootCmd.Flags().StringSliceVarP(&nodeIPs, "node-ips", "n", []string{}, "")
	rootCmd.MarkFlagRequired("node-ips")

	rootCmd.Flags().StringVarP(&sshKeyPath, "ssh-key-path", "k", "", "derpa derpa dee")
	rootCmd.MarkFlagRequired("ssh-key-path")
}
