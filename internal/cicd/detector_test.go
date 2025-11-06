package cicd

import (
	"os"
	"testing"
)

func TestDetectGitHub(t *testing.T) {
	// Save original env
	originalEnv := make(map[string]string)
	envVars := []string{
		"GITHUB_ACTIONS", "GITHUB_REPOSITORY", "GITHUB_SHA",
		"GITHUB_REF_NAME", "GITHUB_RUN_ID", "GITHUB_RUN_NUMBER",
		"GITHUB_ACTOR", "GITHUB_SERVER_URL",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, val := range originalEnv {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Set GitHub Actions environment
	os.Setenv("GITHUB_ACTIONS", "true")
	os.Setenv("GITHUB_REPOSITORY", "versioner-io/versioner-cli")
	os.Setenv("GITHUB_SHA", "abc123def456789012345678901234567890abcd")
	os.Setenv("GITHUB_REF_NAME", "main")
	os.Setenv("GITHUB_RUN_ID", "123456")
	os.Setenv("GITHUB_RUN_NUMBER", "42")
	os.Setenv("GITHUB_ACTOR", "testuser")
	os.Setenv("GITHUB_SERVER_URL", "https://github.com")

	detected := Detect()

	if detected.System != SystemGitHub {
		t.Errorf("Expected system %s, got %s", SystemGitHub, detected.System)
	}

	if detected.SCMRepository != "versioner-io/versioner-cli" {
		t.Errorf("Expected repository versioner-io/versioner-cli, got %s", detected.SCMRepository)
	}

	if detected.Product != "versioner-cli" {
		t.Errorf("Expected product versioner-cli, got %s", detected.Product)
	}

	if detected.SCMSha != "abc123def456789012345678901234567890abcd" {
		t.Errorf("Expected SHA abc123def456789012345678901234567890abcd, got %s", detected.SCMSha)
	}

	if detected.Version != "abc123de" {
		t.Errorf("Expected version abc123de, got %s", detected.Version)
	}

	expectedURL := "https://github.com/versioner-io/versioner-cli/actions/runs/123456"
	if detected.BuildURL != expectedURL {
		t.Errorf("Expected build URL %s, got %s", expectedURL, detected.BuildURL)
	}
}

func TestDetectGitLab(t *testing.T) {
	// Save and clear environment (including GitHub Actions vars that may be set in CI)
	originalEnv := make(map[string]string)
	envVars := []string{
		"GITLAB_CI", "CI_PROJECT_PATH", "CI_COMMIT_SHA",
		"CI_COMMIT_REF_NAME", "CI_PIPELINE_ID", "CI_PIPELINE_IID",
		"CI_PIPELINE_URL", "GITLAB_USER_LOGIN",
		"GITHUB_ACTIONS", "GITHUB_REPOSITORY", "GITHUB_SHA",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, val := range originalEnv {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Set GitLab CI environment
	os.Setenv("GITLAB_CI", "true")
	os.Setenv("CI_PROJECT_PATH", "myorg/my-project")
	os.Setenv("CI_COMMIT_SHA", "def456abc789012345678901234567890abcdef1")
	os.Setenv("CI_COMMIT_REF_NAME", "develop")
	os.Setenv("CI_PIPELINE_ID", "789")
	os.Setenv("CI_PIPELINE_IID", "123")
	os.Setenv("CI_PIPELINE_URL", "https://gitlab.com/myorg/my-project/-/pipelines/789")
	os.Setenv("GITLAB_USER_LOGIN", "testuser")

	detected := Detect()

	if detected.System != SystemGitLab {
		t.Errorf("Expected system %s, got %s", SystemGitLab, detected.System)
	}

	if detected.SCMRepository != "myorg/my-project" {
		t.Errorf("Expected repository myorg/my-project, got %s", detected.SCMRepository)
	}

	if detected.Product != "my-project" {
		t.Errorf("Expected product my-project, got %s", detected.Product)
	}
}

func TestDetectRundeck(t *testing.T) {
	// Save and clear environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"RD_JOB_ID", "RD_JOB_EXECID", "RD_JOB_NAME",
		"RD_JOB_USERNAME", "RD_JOB_USER_NAME", "RD_JOB_PROJECT",
		"RD_JOB_SERVERURL",
		"GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, val := range originalEnv {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Set Rundeck environment
	os.Setenv("RD_JOB_ID", "abc-123-def-456")
	os.Setenv("RD_JOB_EXECID", "789")
	os.Setenv("RD_JOB_NAME", "deploy-api-service")
	os.Setenv("RD_JOB_USERNAME", "testuser")
	os.Setenv("RD_JOB_PROJECT", "production")
	os.Setenv("RD_JOB_SERVERURL", "https://rundeck.example.com")

	detected := Detect()

	if detected.System != SystemRundeck {
		t.Errorf("Expected system %s, got %s", SystemRundeck, detected.System)
	}

	if detected.Product != "deploy-api-service" {
		t.Errorf("Expected product deploy-api-service, got %s", detected.Product)
	}

	if detected.BuildNumber != "789" {
		t.Errorf("Expected build number 789, got %s", detected.BuildNumber)
	}

	if detected.InvokeID != "789" {
		t.Errorf("Expected invoke ID 789, got %s", detected.InvokeID)
	}

	if detected.BuiltBy != "testuser" {
		t.Errorf("Expected built by testuser, got %s", detected.BuiltBy)
	}

	if detected.Version != "789" {
		t.Errorf("Expected version 789, got %s", detected.Version)
	}

	expectedURL := "https://rundeck.example.com/project/production/execution/show/789"
	if detected.BuildURL != expectedURL {
		t.Errorf("Expected build URL %s, got %s", expectedURL, detected.BuildURL)
	}
}

func TestDetectUnknown(t *testing.T) {
	// Clear all CI environment variables
	ciEnvVars := []string{
		"GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "CIRCLECI",
		"BITBUCKET_BUILD_NUMBER", "TF_BUILD", "TRAVIS", "RD_JOB_ID",
	}
	originalEnv := make(map[string]string)
	for _, key := range ciEnvVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, val := range originalEnv {
			if val != "" {
				os.Setenv(key, val)
			}
		}
	}()

	detected := Detect()

	if detected.System != SystemUnknown {
		t.Errorf("Expected system %s, got %s", SystemUnknown, detected.System)
	}
}

func TestNormalizeGitURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com/owner/repo.git", "github.com/owner/repo"},
		{"http://github.com/owner/repo", "github.com/owner/repo"},
		{"git@github.com:owner/repo.git", "github.com/owner/repo"},
		{"github.com/owner/repo", "github.com/owner/repo"},
	}

	for _, test := range tests {
		result := normalizeGitURL(test.input)
		if result != test.expected {
			t.Errorf("normalizeGitURL(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}
