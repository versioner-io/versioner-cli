package github

import (
	"encoding/json"
	"fmt"
	"os"
)

// WriteErrorAnnotation writes a GitHub Actions error annotation and job summary
// This makes errors visible in the GitHub Actions UI without digging through logs
func WriteErrorAnnotation(statusCode int, errorCode, message, ruleName string, retryAfter string, details map[string]interface{}) {
	// Only write annotations if running in GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return
	}

	// Write workflow command annotation
	writeWorkflowCommand(statusCode, errorCode, message, ruleName)

	// Write job summary
	writeJobSummary(statusCode, errorCode, message, ruleName, retryAfter, details)
}

// WriteSuccessSummary writes a GitHub Actions job summary for successful deployment tracking
func WriteSuccessSummary(action, environment, status, version, scmSha, deployURL string) {
	// Only write summaries if running in GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return
	}

	summaryPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryPath == "" {
		return
	}

	// Build the summary
	var summary string
	summary += "## 🚀 Versioner Summary\n\n"

	// Add key information
	summary += fmt.Sprintf("- **Action:** %s\n", action)
	summary += fmt.Sprintf("- **Environment:** %s\n", environment)
	summary += fmt.Sprintf("- **Status:** %s\n", formatStatus(status))
	summary += fmt.Sprintf("- **Version:** `%s`\n", version)

	if scmSha != "" {
		summary += fmt.Sprintf("- **Git SHA:** `%s`\n", scmSha)
	}

	if deployURL != "" {
		summary += fmt.Sprintf("\n[View Deployment Run →](%s)\n", deployURL)
	}

	// Write to file
	f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Silently fail - don't break the CLI if we can't write the summary
		return
	}
	defer f.Close()

	_, _ = f.WriteString(summary)
}

// formatStatus adds an emoji to the status for visual clarity
func formatStatus(status string) string {
	switch status {
	case "started", "in_progress":
		return "⏳ " + status
	case "completed", "success":
		return "✅ " + status
	case "failed":
		return "❌ " + status
	case "aborted", "cancelled":
		return "🚫 " + status
	case "pending":
		return "⏸️ " + status
	default:
		return status
	}
}

// writeWorkflowCommand outputs a GitHub Actions workflow command for error annotation
func writeWorkflowCommand(statusCode int, errorCode, message, ruleName string) {
	// Format: ::error title=<title>::<message>
	title := formatTitle(statusCode, errorCode, ruleName)

	// Escape special characters in message
	escapedMessage := escapeWorkflowCommand(message)

	fmt.Fprintf(os.Stdout, "::error title=%s::%s\n", title, escapedMessage)
}

// writeJobSummary writes a detailed error summary to GITHUB_STEP_SUMMARY
func writeJobSummary(statusCode int, errorCode, message, ruleName string, retryAfter string, details map[string]interface{}) {
	summaryPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryPath == "" {
		return
	}

	var summary string
	summary += "## ❌ Versioner Deployment Rejected\n\n"

	// Add status-specific emoji and title
	switch statusCode {
	case 409:
		summary += "### ⚠️ Deployment Conflict\n\n"
	case 423:
		summary += "### 🔒 Deployment Blocked by Schedule\n\n"
	case 428:
		summary += "### ❌ Deployment Precondition Failed\n\n"
	}

	// Add key information
	summary += fmt.Sprintf("- **Error Code:** `%s`\n", errorCode)
	if ruleName != "" {
		summary += fmt.Sprintf("- **Rule:** %s\n", ruleName)
	}
	summary += fmt.Sprintf("- **Message:** %s\n", message)

	if retryAfter != "" {
		summary += fmt.Sprintf("- **Retry After:** `%s`\n", retryAfter)
	}

	summary += "\n"

	// Add specific guidance based on status code and error code
	summary += "**Action Required:**\n"
	switch statusCode {
	case 409:
		summary += "- Wait for the current deployment to complete\n"
		summary += "- Retry this deployment\n"

	case 423:
		if retryAfter != "" {
			summary += fmt.Sprintf("- Wait until `%s`\n", retryAfter)
			summary += "- Retry automatically after the no-deploy window\n"
		}
		summary += "- Or use `--skip-preflight-checks` for emergencies\n"

	case 428:
		switch errorCode {
		case "FLOW_VIOLATION":
			summary += "- Deploy to required environments first\n"
			summary += "- Then retry this deployment\n"

		case "INSUFFICIENT_SOAK_TIME":
			summary += "- Wait for the soak time requirement to be met\n"
			if retryAfter != "" {
				summary += fmt.Sprintf("- Can deploy at: `%s`\n", retryAfter)
			}
			summary += "- Or use `--skip-preflight-checks` for emergencies\n"

		case "QUALITY_APPROVAL_REQUIRED", "APPROVAL_REQUIRED":
			summary += "- Obtain required approval via Versioner UI\n"
			summary += "- Then retry this deployment\n"

		default:
			summary += "- Resolve the issue described above\n"
			summary += "- Then retry this deployment\n"
			summary += "- Or use `--skip-preflight-checks` for emergencies\n"
		}
	}

	// Add details section if available
	if len(details) > 0 {
		summary += "\n**Details:**\n"
		summary += "```json\n"
		detailsJSON, err := json.MarshalIndent(details, "", "  ")
		if err == nil {
			summary += string(detailsJSON)
		}
		summary += "\n```\n"
	}

	// Write to file
	f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Silently fail - don't break the CLI if we can't write the summary
		return
	}
	defer f.Close()

	_, _ = f.WriteString(summary)
}

// formatTitle creates a concise title for the error annotation
func formatTitle(statusCode int, errorCode, ruleName string) string {
	switch statusCode {
	case 409:
		return "Deployment Conflict"
	case 423:
		if ruleName != "" {
			return fmt.Sprintf("Deployment Blocked: %s", ruleName)
		}
		return "Deployment Blocked by Schedule"
	case 428:
		if ruleName != "" {
			return fmt.Sprintf("%s: %s", errorCode, ruleName)
		}
		return errorCode
	default:
		return "Deployment Rejected"
	}
}

// escapeWorkflowCommand escapes special characters for GitHub Actions workflow commands
// See: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-error-message
func escapeWorkflowCommand(s string) string {
	s = replaceAll(s, "%", "%25")
	s = replaceAll(s, "\r", "%0D")
	s = replaceAll(s, "\n", "%0A")
	return s
}

// replaceAll is a simple string replacement helper
func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}
