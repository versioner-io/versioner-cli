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

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Track a build event",
	Long: `Track a CI/CD build lifecycle event with the Versioner API.
This command sends build information to track the build process.`,
	Example: `  # Track a completed build
  versioner track build --product=api-service --version=1.2.3 --status=completed

  # Track a build with additional metadata
  versioner track build \
    --product=api-service \
    --version=1.2.3 \
    --status=completed \
    --scm-sha=abc123 \
    --build-number=456`,
	RunE: runBuildTrack,
}

func init() {
	trackCmd.AddCommand(buildCmd)

	// Required flags
	buildCmd.Flags().String("product", "", "Product/application name (required)")
	buildCmd.Flags().String("version", "", "Version string (required)")
	buildCmd.Flags().String("status", "completed", "Build status (pending, started, completed, failed, aborted)")

	// Optional flags
	buildCmd.Flags().String("source-system", "", "Source system (github, jenkins, gitlab, etc.)")
	buildCmd.Flags().String("build-number", "", "Build number from CI system")
	buildCmd.Flags().String("scm-sha", "", "Git commit SHA (40-character hash)")
	buildCmd.Flags().String("scm-branch", "", "Git branch name")
	buildCmd.Flags().String("scm-repository", "", "Source control repository (e.g., owner/repo)")
	buildCmd.Flags().String("build-url", "", "Link to CI/CD build run")
	buildCmd.Flags().String("invoke-id", "", "Invocation/run ID from CI system")
	buildCmd.Flags().String("built-by", "", "User identifier (username, email, or ID)")
	buildCmd.Flags().String("built-by-email", "", "User email")
	buildCmd.Flags().String("built-by-name", "", "User display name")
	buildCmd.Flags().String("started-at", "", "Build start timestamp (ISO 8601 format)")
	buildCmd.Flags().String("completed-at", "", "Build completion timestamp (ISO 8601 format)")
	buildCmd.Flags().String("extra-metadata", "", "Additional metadata as JSON object (max 100KB)")

	// Bind flags to viper
	_ = viper.BindPFlag("product", buildCmd.Flags().Lookup("product"))
	_ = viper.BindPFlag("version", buildCmd.Flags().Lookup("version"))
	_ = viper.BindPFlag("status", buildCmd.Flags().Lookup("status"))
	_ = viper.BindPFlag("source_system", buildCmd.Flags().Lookup("source-system"))
	_ = viper.BindPFlag("build_number", buildCmd.Flags().Lookup("build-number"))
	_ = viper.BindPFlag("scm_sha", buildCmd.Flags().Lookup("scm-sha"))
	_ = viper.BindPFlag("scm_branch", buildCmd.Flags().Lookup("scm-branch"))
	_ = viper.BindPFlag("scm_repository", buildCmd.Flags().Lookup("scm-repository"))
	_ = viper.BindPFlag("build_url", buildCmd.Flags().Lookup("build-url"))
	_ = viper.BindPFlag("invoke_id", buildCmd.Flags().Lookup("invoke-id"))
	_ = viper.BindPFlag("built_by", buildCmd.Flags().Lookup("built-by"))
	_ = viper.BindPFlag("built_by_email", buildCmd.Flags().Lookup("built-by-email"))
	_ = viper.BindPFlag("built_by_name", buildCmd.Flags().Lookup("built-by-name"))
}

func runBuildTrack(cmd *cobra.Command, args []string) error {
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
	event := &api.BuildEventCreate{
		ProductName:   product,
		Version:       version,
		Status:        statusValue,
		SourceSystem:  getWithFallback("source-system", "source_system", string(detected.System)),
		BuildNumber:   getWithFallback("build-number", "build_number", detected.BuildNumber),
		SCMSha:        getWithFallback("scm-sha", "scm_sha", detected.SCMSha),
		SCMBranch:     getWithFallback("scm-branch", "scm_branch", detected.SCMBranch),
		SCMRepository: getWithFallback("scm-repository", "scm_repository", detected.SCMRepository),
		BuildURL:      getWithFallback("build-url", "build_url", detected.BuildURL),
		InvokeID:      getWithFallback("invoke-id", "invoke_id", detected.InvokeID),
		BuiltBy:       getWithFallback("built-by", "built_by", detected.BuiltBy),
		BuiltByEmail:  getWithFallback("built-by-email", "built_by_email", detected.BuiltByEmail),
		BuiltByName:   getWithFallback("built-by-name", "built_by_name", detected.BuiltByName),
	}

	// Parse timestamps if provided
	if startedAtStr := cmd.Flags().Lookup("started-at").Value.String(); startedAtStr != "" {
		startedAt, err := time.Parse(time.RFC3339, startedAtStr)
		if err != nil {
			return fmt.Errorf("invalid started-at timestamp: %w", err)
		}
		event.StartedAt = &startedAt
	}

	if completedAtStr := cmd.Flags().Lookup("completed-at").Value.String(); completedAtStr != "" {
		completedAt, err := time.Parse(time.RFC3339, completedAtStr)
		if err != nil {
			return fmt.Errorf("invalid completed-at timestamp: %w", err)
		}
		event.CompletedAt = &completedAt
	}

	// Parse extra metadata if provided
	if extraMetadataStr, _ := cmd.Flags().GetString("extra-metadata"); extraMetadataStr != "" {
		metadata, err := ParseExtraMetadata(extraMetadataStr)
		if err != nil {
			return err
		}
		event.ExtraMetadata = metadata
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Tracking build event:\n")
		if detected.System != cicd.SystemUnknown {
			fmt.Fprintf(os.Stderr, "  ℹ Auto-detected CI system: %s\n", detected.System)
		}
		fmt.Fprintf(os.Stderr, "  Product: %s\n", product)
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
	resp, err := client.CreateBuildEvent(event)
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
	fmt.Printf("✓ Build event tracked successfully\n")
	fmt.Printf("  Event ID: %s\n", resp.ID)
	if verbose {
		fmt.Printf("  Product ID: %s\n", resp.ProductID)
		fmt.Printf("  Version ID: %s\n", resp.VersionID)
	}

	return nil
}
