package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/versioner-io/versioner-cli/internal/api"
	"github.com/versioner-io/versioner-cli/internal/cicd"
	"github.com/versioner-io/versioner-cli/internal/github"
	"github.com/versioner-io/versioner-cli/internal/status"
)

var deploymentCmd = &cobra.Command{
	Use:   "deployment",
	Short: "Track a deployment event",
	Long: `Track a deployment lifecycle event with the Versioner API.
This command sends deployment information to track deployments to environments.

When status=started, the API automatically runs preflight checks to validate:
- No concurrent deployments (409 Conflict)
- No-deploy windows/schedules (423 Locked)
- Flow requirements, soak time, approvals (428 Precondition Required)

Exit codes:
  0 - Success
  1 - General error (network, invalid arguments)
  4 - API error (validation, authentication)
  5 - Preflight check failure (deployment blocked)`,
	Example: `  # Track a deployment start (triggers preflight checks)
  versioner track deployment \
    --product=api-service \
    --environment=production \
    --version=1.2.3 \
    --status=started

  # Track deployment completion
  versioner track deployment \
    --product=api-service \
    --environment=production \
    --version=1.2.3 \
    --status=completed

  # Emergency deployment (skip preflight checks)
  versioner track deployment \
    --product=api-service \
    --environment=production \
    --version=1.2.3-hotfix \
    --status=started \
    --skip-preflight-checks`,
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
	deploymentCmd.Flags().String("deploy-url", "", "Link to deployment run/logs")
	deploymentCmd.Flags().String("invoke-id", "", "Invocation/run ID from CI system")
	deploymentCmd.Flags().String("deployed-by", "", "User identifier (username, email, or ID)")
	deploymentCmd.Flags().String("deployed-by-email", "", "User email")
	deploymentCmd.Flags().String("deployed-by-name", "", "User display name")
	deploymentCmd.Flags().String("completed-at", "", "Deployment completion timestamp (ISO 8601 format)")
	deploymentCmd.Flags().String("extra-metadata", "", "Additional metadata as JSON object (max 100KB)")
	deploymentCmd.Flags().Bool("fail-on-api-error", true, "Fail command if API is unreachable or returns auth/validation errors (default: true)")
	deploymentCmd.Flags().Bool("skip-preflight-checks", false, "Skip preflight checks (emergency use only)")

	// Bind flags to viper
	_ = viper.BindPFlag("product", deploymentCmd.Flags().Lookup("product"))
	_ = viper.BindPFlag("environment", deploymentCmd.Flags().Lookup("environment"))
	_ = viper.BindPFlag("version", deploymentCmd.Flags().Lookup("version"))
	_ = viper.BindPFlag("status", deploymentCmd.Flags().Lookup("status"))
	_ = viper.BindPFlag("source_system", deploymentCmd.Flags().Lookup("source-system"))
	_ = viper.BindPFlag("build_number", deploymentCmd.Flags().Lookup("build-number"))
	_ = viper.BindPFlag("scm_sha", deploymentCmd.Flags().Lookup("scm-sha"))
	_ = viper.BindPFlag("scm_repository", deploymentCmd.Flags().Lookup("scm-repository"))
	_ = viper.BindPFlag("deploy_url", deploymentCmd.Flags().Lookup("deploy-url"))
	_ = viper.BindPFlag("invoke_id", deploymentCmd.Flags().Lookup("invoke-id"))
	_ = viper.BindPFlag("deployed_by", deploymentCmd.Flags().Lookup("deployed-by"))
	_ = viper.BindPFlag("deployed_by_email", deploymentCmd.Flags().Lookup("deployed-by-email"))
	_ = viper.BindPFlag("deployed_by_name", deploymentCmd.Flags().Lookup("deployed-by-name"))
	_ = viper.BindPFlag("fail_on_api_error", deploymentCmd.Flags().Lookup("fail-on-api-error"))
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

	// Get fail-on-api-error flag (default: true)
	failOnApiError, _ := cmd.Flags().GetBool("fail-on-api-error")
	if !cmd.Flags().Changed("fail-on-api-error") {
		failOnApiError = viper.GetBool("fail_on_api_error")
		if !viper.IsSet("fail_on_api_error") {
			failOnApiError = true
		}
	}

	// Create API client
	client := api.NewClient(apiURL, apiKey, debug, failOnApiError)

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
		DeployURL:       getWithFallback("deploy-url", "deploy_url", detected.BuildURL),
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

	// Get auto-detected metadata from CI/CD system
	autoMetadata := detected.ExtraMetadata()

	// Parse user-provided extra metadata if provided
	var userMetadata map[string]interface{}
	if extraMetadataStr, _ := cmd.Flags().GetString("extra-metadata"); extraMetadataStr != "" {
		var err error
		userMetadata, err = ParseExtraMetadata(extraMetadataStr)
		if err != nil {
			return err
		}
	}

	// Merge metadata (user values take precedence)
	event.ExtraMetadata = MergeMetadata(autoMetadata, userMetadata)

	// Get skip-preflight-checks flag
	skipPreflightChecks, _ := cmd.Flags().GetBool("skip-preflight-checks")
	if skipPreflightChecks {
		fmt.Fprintf(os.Stderr, "⚠️  DEPRECATION WARNING: --skip-preflight-checks is deprecated\n")
		fmt.Fprintf(os.Stderr, "    Use server-side rule status control instead (disabled/report_only/enabled)\n")
		fmt.Fprintf(os.Stderr, "    Admins can change rule status in the Versioner UI without code changes\n")
		fmt.Fprintf(os.Stderr, "    This flag will be removed in a future version\n\n")
	}
	event.SkipPreflightChecks = skipPreflightChecks

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
			// Check if this is a preflight check failure
			if apiErr.IsPreflightError() {
				handlePreflightError(apiErr)
				os.Exit(5) // Exit code 5 for preflight failures
			}
			// Other API error - exit code 4
			github.WriteGenericErrorAnnotation("Deployment", "API Error", apiErr.Error())
			fmt.Fprintf(os.Stderr, "API error: %s\n", apiErr.Error())
			os.Exit(4)
		}
		// Network or other error - exit code 1
		github.WriteGenericErrorAnnotation("Deployment", "Network Error", err.Error())
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	// Success
	fmt.Printf("✓ Deployment event tracked successfully\n")
	fmt.Printf("  Event ID: %s\n", resp.ID)
	if verbose {
		fmt.Printf("  Product ID: %s\n", resp.ProductID)
		fmt.Printf("  Version ID: %s\n", resp.VersionID)
		fmt.Printf("  Environment ID: %s\n", resp.EnvironmentID)
	}

	// Write GitHub Actions job summary
	uiURL := viper.GetString("ui_url")
	github.WriteSuccessSummary("Deployment", environment, statusValue, version, event.SCMSha, uiURL, resp.ID)

	return nil
}

// handlePreflightError formats and displays preflight check errors
func handlePreflightError(apiErr *api.APIError) {
	_, message, code, retryAfter, details, ok := apiErr.GetPreflightDetails()
	if !ok {
		// Fallback if we can't parse the error structure
		fmt.Fprintf(os.Stderr, "❌ Deployment Failed (HTTP %d)\n\n", apiErr.StatusCode)
		fmt.Fprintf(os.Stderr, "%s\n", apiErr.Error())
		return
	}

	// Get rule name from details if available
	ruleName := ""
	if details != nil {
		if name, exists := details["rule_name"].(string); exists {
			ruleName = name
		}
	}

	// Write GitHub Actions annotation if running in GitHub Actions
	github.WriteErrorAnnotation(apiErr.StatusCode, code, message, ruleName, retryAfter, details)

	// Format output based on status code and error code
	switch apiErr.StatusCode {
	case 409:
		// Deployment Conflict
		fmt.Fprintf(os.Stderr, "⚠️  Deployment Conflict\n\n")
		fmt.Fprintf(os.Stderr, "%s\n", message)
		fmt.Fprintf(os.Stderr, "Another deployment is in progress. Please wait and retry.\n")

	case 423:
		// Schedule Block
		fmt.Fprintf(os.Stderr, "🔒 Deployment Blocked by Schedule\n\n")
		if ruleName != "" {
			fmt.Fprintf(os.Stderr, "Rule: %s\n", ruleName)
		}
		fmt.Fprintf(os.Stderr, "%s\n", message)
		if retryAfter != "" {
			fmt.Fprintf(os.Stderr, "\nRetry after: %s\n", retryAfter)
		}
		fmt.Fprintf(os.Stderr, "\nTo skip checks (emergency only), add:\n")
		fmt.Fprintf(os.Stderr, "  --skip-preflight-checks\n")

	case 428:
		// Precondition Failed
		fmt.Fprintf(os.Stderr, "❌ Deployment Precondition Failed\n\n")
		fmt.Fprintf(os.Stderr, "Error: %s\n", code)
		if ruleName != "" {
			fmt.Fprintf(os.Stderr, "Rule: %s\n", ruleName)
		}
		fmt.Fprintf(os.Stderr, "%s\n", message)

		// Specific guidance based on error code
		switch code {
		case "FLOW_VIOLATION":
			fmt.Fprintf(os.Stderr, "\nDeploy to required environments first, then retry.\n")

		case "INSUFFICIENT_SOAK_TIME":
			if retryAfter != "" {
				fmt.Fprintf(os.Stderr, "\nRetry after: %s\n", retryAfter)
			}
			fmt.Fprintf(os.Stderr, "\nWait for soak time to complete, then retry.\n")
			fmt.Fprintf(os.Stderr, "\nTo skip checks (emergency only), add:\n")
			fmt.Fprintf(os.Stderr, "  --skip-preflight-checks\n")

		case "QUALITY_APPROVAL_REQUIRED", "APPROVAL_REQUIRED":
			fmt.Fprintf(os.Stderr, "\nApproval required before deployment can proceed.\n")
			fmt.Fprintf(os.Stderr, "Obtain approval via Versioner UI, then retry.\n")

		default:
			// Unknown error code - provide generic guidance
			if retryAfter != "" {
				fmt.Fprintf(os.Stderr, "\nRetry after: %s\n", retryAfter)
			}
			fmt.Fprintf(os.Stderr, "\nResolve the issue described above, then retry.\n")
			fmt.Fprintf(os.Stderr, "\nTo skip checks (emergency only), add:\n")
			fmt.Fprintf(os.Stderr, "  --skip-preflight-checks\n")
		}
	}

	// Always print full details for debugging
	if details != nil {
		fmt.Fprintf(os.Stderr, "\nDetails:\n")
		detailsJSON, err := json.MarshalIndent(details, "  ", "  ")
		if err == nil {
			fmt.Fprintf(os.Stderr, "  %s\n", string(detailsJSON))
		}
	}
}
