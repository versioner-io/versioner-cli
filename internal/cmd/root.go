package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	debug   bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "versioner",
	Short: "Track build and deployment events with Versioner",
	Long: `Versioner CLI is a command-line tool for tracking build and deployment events
in your CI/CD pipelines. It sends events to the Versioner API for deployment
tracking, visibility, and audit purposes.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.versioner/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug output (includes HTTP requests/responses)")

	// API configuration flags
	rootCmd.PersistentFlags().String("api-url", "", "Versioner API URL (default: https://api.versioner.io)")
	rootCmd.PersistentFlags().String("api-key", "", "Versioner API key (prefer VERSIONER_API_KEY env var)")

	// Bind flags to viper
	viper.BindPFlag("api_url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
}

// initConfig reads in config file and ENV variables
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
			os.Exit(1)
		}

		// Search for config in home directory and current directory
		viper.AddConfigPath(home + "/.versioner")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Environment variables
	viper.SetEnvPrefix("VERSIONER")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("api_url", "https://api.versioner.io")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}

	// Warn if API key is passed via flag (security concern)
	if rootCmd.PersistentFlags().Changed("api-key") {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Passing API key via --api-key flag is visible in process lists.\n")
		fmt.Fprintf(os.Stderr, "   Prefer using VERSIONER_API_KEY environment variable or config file.\n\n")
	}
}
