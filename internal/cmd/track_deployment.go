package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/versioner-io/versioner-cli/internal/api"
	"github.com/versioner-io/versioner-cli/internal/cicd"
	"github.com/versioner-io/versioner-cli/internal/status"
)

var deploymentCmd = &cobra.Command{
	Use:   "deployment",
	Short: "Track a deployment event",
	Long: `Track a deployment lifecycle event with the Versioner API.
This command sends deployment information to track deployments to environments.`,
	Example: `  # Track a successful deployment
  versioner track deployment \
    --product=api-service \
    --environment=production \
    --version=1.2.3 \
    --status=success

  # Track a deployment with additional metadata
  versioner track deployment \
    --product=api-service \
    --environment=staging \
    --version=1.2.3 \
    --status=success \
    --scm-sha=abc123 \
    --build-number=456`,
	RunE: runDeploymentTrack,
}

func init() {
	trackCmd.AddCommand(deploymentCmd)

	// Required flags
	deploymentCmd.Flags().String("product", "", "Product/application name (required)")
	deploymentCmd.Flags().String("environment", "", "Environment name (required)")
	deploymentCmd.Flags().String("version", "", "Version string (required)")
	deploymentCmd.Flags().String("status", "success", "Deployment status (pending, started, completed, failed, aborted)")

	// Optional flags
	deploymentCmd.Flags().String("source-system", "", "Source system (github, jenkins, gitlab, etc.)")
	deploymentCmd.Flags().String("build-number", "", "Build number from CI system")
	deploymentCmd.Flags().String("scm-sha", "", "Git commit SHA (40-character hash)")
	deploymentCmd.Flags().String("scm-repository", "", "Source control repository (e.g., owner/repo)")
	deploymentCmd.Flags().String("build-url", "", "Link to CI/CD build run")
	deploymentCmd.Flags().String("invoke-id", "", "Invocation/run ID from CI system")
	deploymentCmd.Flags().String("deployed-by", "", "User identifier (username, email, or ID)")
	deploymentCmd.Flags().String("deployed-by-email", "", "User email")
	deploymentCmd.Flags().String("deployed-by-name", "", "User display name")
	deploymentCmd.Flags().String("completed-at", "", "Deployment completion timestamp (ISO 8601 format)")

	// Bind flags to viper
	viper.BindPFlag("product", deploymentCmd.Flags().Lookup("product"))
	viper.BindPFlag("environment", deploymentCmd.Flags().Lookup("environment"))
	viper.BindPFlag("version", deploymentCmd.Flags().Lookup("version"))
	viper.BindPFlag("status", deploymentCmd.Flags().Lookup("status"))
	viper.BindPFlag("source_system", deploymentCmd.Flags().Lookup("source-system"))
	viper.BindPFlag("build_number", deploymentCmd.Flags().Lookup("build-number"))
	viper.BindPFlag("scm_sha", deploymentCmd.Flags().Lookup("scm-sha"))
	viper.BindPFlag("scm_repository", deploymentCmd.Flags().Lookup("scm-repository"))
	viper.BindPFlag("build_url", deploymentCmd.Flags().Lookup("build-url"))
	viper.BindPFlag("invoke_id", deploymentCmd.Flags().Lookup("invoke-id"))
	viper.BindPFlag("deployed_by", deploymentCmd.Flags().Lookup("deployed-by"))
	viper.BindPFlag("deployed_by_email", deploymentCmd.Flags().Lookup("deployed-by-email"))
	viper.BindPFlag("deployed_by_name", deploymentCmd.Flags().Lookup("deployed-by-name"))
}

func runDeploymentTrack(cmd *cobra.Command, args []string) error {
	// Auto-detect CI/CD environment
	detected := cicd.Detect()

	// Get required fields (with auto-detection fallback)
	product, _ := cmd.Flags().GetString("product")
	if product == "" {
		product = viper.GetString("product")
	}
	if product == "" {
		product = detected.Product
	}

	environment, _ := cmd.Flags().GetString("environment")
	if environment == "" {
		environment = viper.GetString("environment")
	}

	version, _ := cmd.Flags().GetString("version")
	if version == "" {
		version = viper.GetString("version")
	}
	if version == "" {
		version = detected.Version
	}

	statusValue, _ := cmd.Flags().GetString("status")

	// Normalize and validate status
	canonicalStatus, wasNormalized := status.Normalize(statusValue)
	if verbose && wasNormalized {
		fmt.Fprintf(os.Stderr, "ℹ Status '%s' will be normalized to '%s' by the API\n", statusValue, canonicalStatus)
	}

	// Validate required fields
	if product == "" {
		return fmt.Errorf("--product is required")
	}
	if environment == "" {
		return fmt.Errorf("--environment is required")
	}
	if version == "" {
		return fmt.Errorf("--version is required")
	}

	// Get API configuration
	apiURL := viper.GetString("api_url")
	apiKey := viper.GetString("api_key")

	if apiKey == "" {
		return fmt.Errorf("API key is required. Set VERSIONER_API_KEY environment variable or use --api-key flag")
	}

	// Create API client
	client := api.NewClient(apiURL, apiKey, debug)

	// Helper function to get value with fallback (cmd flags -> viper -> auto-detected)
	getWithFallback := func(flagName string, viperKey string, fallback string) string {
		// Try command flag first
		if val, _ := cmd.Flags().GetString(flagName); val != "" {
			return val
		}
		// Try viper (env vars, config file)
		if val := viper.GetString(viperKey); val != "" {
			return val
		}
		// Fall back to auto-detected value
		return fallback
	}

	// Build the event with auto-detected fallbacks
	event := &api.DeploymentEventCreate{
		ProductName:     product,
		Version:         version,
		EnvironmentName: environment,
		Status:          statusValue,
		SourceSystem:    getWithFallback("source-system", "source_system", string(detected.System)),
		BuildNumber:     getWithFallback("build-number", "build_number", detected.BuildNumber),
		SCMSha:          getWithFallback("scm-sha", "scm_sha", detected.SCMSha),
		SCMRepository:   getWithFallback("scm-repository", "scm_repository", detected.SCMRepository),
		BuildURL:        getWithFallback("build-url", "build_url", detected.BuildURL),
		InvokeID:        getWithFallback("invoke-id", "invoke_id", detected.InvokeID),
		DeployedBy:      getWithFallback("deployed-by", "deployed_by", detected.BuiltBy),
		DeployedByEmail: getWithFallback("deployed-by-email", "deployed_by_email", detected.BuiltByEmail),
		DeployedByName:  getWithFallback("deployed-by-name", "deployed_by_name", detected.BuiltByName),
	}

	// Parse timestamp if provided
	if completedAtStr := cmd.Flags().Lookup("completed-at").Value.String(); completedAtStr != "" {
		completedAt, err := time.Parse(time.RFC3339, completedAtStr)
		if err != nil {
			return fmt.Errorf("invalid completed-at timestamp: %w", err)
		}
		event.CompletedAt = &completedAt
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Tracking deployment event:\n")
		if detected.System != cicd.SystemUnknown {
			fmt.Fprintf(os.Stderr, "  ℹ Auto-detected CI system: %s\n", detected.System)
		}
		fmt.Fprintf(os.Stderr, "  Product: %s\n", product)
		fmt.Fprintf(os.Stderr, "  Environment: %s\n", environment)
		fmt.Fprintf(os.Stderr, "  Version: %s\n", version)
		fmt.Fprintf(os.Stderr, "  Status: %s\n", statusValue)
		if event.SourceSystem != "" {
			fmt.Fprintf(os.Stderr, "  Source System: %s\n", event.SourceSystem)
		}
		if event.SCMRepository != "" {
			fmt.Fprintf(os.Stderr, "  Repository: %s\n", event.SCMRepository)
		}
		if event.SCMSha != "" {
			fmt.Fprintf(os.Stderr, "  Commit SHA: %s\n", event.SCMSha)
		}
		fmt.Fprintf(os.Stderr, "  API URL: %s\n", apiURL)
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Send the event
	resp, err := client.CreateDeploymentEvent(event)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			// API error - exit code 2
			fmt.Fprintf(os.Stderr, "API error: %s\n", apiErr.Error())
			os.Exit(2)
		}
		// Network or other error - exit code 2
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(2)
	}

	// Success
	fmt.Printf("✓ Deployment event tracked successfully\n")
	if verbose {
		fmt.Printf("  Event ID: %s\n", resp.ID)
		fmt.Printf("  Product ID: %s\n", resp.ProductID)
		fmt.Printf("  Version ID: %s\n", resp.VersionID)
		fmt.Printf("  Environment ID: %s\n", resp.EnvironmentID)
	}

	return nil
}
