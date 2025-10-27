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
	// Save and clear environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"GITLAB_CI", "CI_PROJECT_PATH", "CI_COMMIT_SHA",
		"CI_COMMIT_REF_NAME", "CI_PIPELINE_ID", "CI_PIPELINE_IID",
		"CI_PIPELINE_URL", "GITLAB_USER_LOGIN",
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

func TestDetectUnknown(t *testing.T) {
	// Clear all CI environment variables
	ciEnvVars := []string{
		"GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "CIRCLECI",
		"BITBUCKET_BUILD_NUMBER", "TF_BUILD", "TRAVIS",
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
