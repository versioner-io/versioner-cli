package cmd

import (
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Track build and deployment events",
	Long: `Track build and deployment events with the Versioner API.
Use 'track build' to track CI/CD build lifecycle events.
Use 'track deployment' to track deployment lifecycle events.`,
}

func init() {
	rootCmd.AddCommand(trackCmd)
}
