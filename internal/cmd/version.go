package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/versioner-io/versioner-cli/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Display the version, git commit, and build date of the Versioner CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("versioner version %s\n", version.Version)
		fmt.Printf("  git commit: %s\n", version.Commit)
		fmt.Printf("  build date: %s\n", version.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
