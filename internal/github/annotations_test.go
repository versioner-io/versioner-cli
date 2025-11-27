package github

import (
	"os"
	"testing"
)

func TestFormatTitle(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errorCode  string
		ruleName   string
		want       string
	}{
		{
			name:       "409 conflict",
			statusCode: 409,
			errorCode:  "DEPLOYMENT_IN_PROGRESS",
			ruleName:   "",
			want:       "Deployment Conflict",
		},
		{
			name:       "423 with rule name",
			statusCode: 423,
			errorCode:  "NO_DEPLOY_WINDOW",
			ruleName:   "No Deploy Fridays",
			want:       "Deployment Blocked: No Deploy Fridays",
		},
		{
			name:       "423 without rule name",
			statusCode: 423,
			errorCode:  "NO_DEPLOY_WINDOW",
			ruleName:   "",
			want:       "Deployment Blocked by Schedule",
		},
		{
			name:       "428 with rule name",
			statusCode: 428,
			errorCode:  "FLOW_VIOLATION",
			ruleName:   "Staging First",
			want:       "FLOW_VIOLATION: Staging First",
		},
		{
			name:       "428 without rule name",
			statusCode: 428,
			errorCode:  "INSUFFICIENT_SOAK_TIME",
			ruleName:   "",
			want:       "INSUFFICIENT_SOAK_TIME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTitle(tt.statusCode, tt.errorCode, tt.ruleName)
			if got != tt.want {
				t.Errorf("formatTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEscapeWorkflowCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special chars",
			input: "simple message",
			want:  "simple message",
		},
		{
			name:  "with percent",
			input: "100% complete",
			want:  "100%25 complete",
		},
		{
			name:  "with newline",
			input: "line1\nline2",
			want:  "line1%0Aline2",
		},
		{
			name:  "with carriage return",
			input: "line1\rline2",
			want:  "line1%0Dline2",
		},
		{
			name:  "multiple special chars",
			input: "50%\ndone\r",
			want:  "50%25%0Adone%0D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeWorkflowCommand(tt.input)
			if got != tt.want {
				t.Errorf("escapeWorkflowCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteErrorAnnotation_NotInGitHub(t *testing.T) {
	// Ensure GITHUB_ACTIONS is not set
	os.Unsetenv("GITHUB_ACTIONS")

	// Should not panic or error when not in GitHub Actions
	WriteErrorAnnotation(423, "NO_DEPLOY_WINDOW", "Test message", "Test Rule", "", nil)
}

func TestWriteErrorAnnotation_InGitHub(t *testing.T) {
	// Set GitHub Actions environment
	os.Setenv("GITHUB_ACTIONS", "true")
	defer os.Unsetenv("GITHUB_ACTIONS")

	// Create a temporary file for GITHUB_STEP_SUMMARY
	tmpFile, err := os.CreateTemp("", "github-summary-*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	os.Setenv("GITHUB_STEP_SUMMARY", tmpFile.Name())
	defer os.Unsetenv("GITHUB_STEP_SUMMARY")

	// Call the function
	details := map[string]interface{}{
		"rule_name":    "Test Rule",
		"window_start": "2025-11-27T00:00:00Z",
	}
	WriteErrorAnnotation(423, "NO_DEPLOY_WINDOW", "No deployments allowed", "Test Rule", "2025-11-27T23:59:59Z", details)

	// Read the summary file
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read summary file: %v", err)
	}

	summary := string(content)

	// Verify key content is present
	if !contains(summary, "Versioner Deployment Rejected") {
		t.Error("Summary should contain 'Versioner Deployment Rejected'")
	}
	if !contains(summary, "NO_DEPLOY_WINDOW") {
		t.Error("Summary should contain error code")
	}
	if !contains(summary, "Test Rule") {
		t.Error("Summary should contain rule name")
	}
	if !contains(summary, "No deployments allowed") {
		t.Error("Summary should contain message")
	}
	if !contains(summary, "2025-11-27T23:59:59Z") {
		t.Error("Summary should contain retry_after timestamp")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
