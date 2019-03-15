package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   string
	gitCommit string
	buildDate string

	shortVersion bool

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			if shortVersion {
				fmt.Println(version)
			} else {
				fmt.Printf("Version: %s\nGitCommit: %s\nBuildDate: %s\n", version, gitCommit, buildDate)
			}
		},
	}
)

func init() {
	versionCmd.Flags().BoolVarP(&shortVersion, "short", "s", false, "only print the version number")

	rootCmd.AddCommand(versionCmd)
}
