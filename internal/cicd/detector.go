package cicd

import (
	"fmt"
	"os"
	"strings"
)

// System represents a detected CI/CD system
type System string

const (
	SystemGitHub    System = "github"
	SystemGitLab    System = "gitlab"
	SystemJenkins   System = "jenkins"
	SystemCircleCI  System = "circleci"
	SystemBitbucket System = "bitbucket"
	SystemAzure     System = "azure-devops"
	SystemTravis    System = "travis"
	SystemUnknown   System = "unknown"
)

// DetectedValues holds auto-detected values from CI/CD environment
type DetectedValues struct {
	System        System
	Product       string
	Version       string
	SCMRepository string
	SCMSha        string
	SCMBranch     string
	BuildNumber   string
	BuildURL      string
	InvokeID      string
	BuiltBy       string
	BuiltByEmail  string
	BuiltByName   string
}

// Detect identifies the CI/CD system and extracts relevant values
func Detect() *DetectedValues {
	detected := &DetectedValues{
		System: detectSystem(),
	}

	switch detected.System {
	case SystemGitHub:
		detectGitHub(detected)
	case SystemGitLab:
		detectGitLab(detected)
	case SystemJenkins:
		detectJenkins(detected)
	case SystemCircleCI:
		detectCircleCI(detected)
	case SystemBitbucket:
		detectBitbucket(detected)
	case SystemAzure:
		detectAzure(detected)
	case SystemTravis:
		detectTravis(detected)
	}

	return detected
}

// detectSystem identifies which CI/CD system is running
func detectSystem() System {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return SystemGitHub
	}
	if os.Getenv("GITLAB_CI") == "true" {
		return SystemGitLab
	}
	if os.Getenv("JENKINS_URL") != "" {
		return SystemJenkins
	}
	if os.Getenv("CIRCLECI") == "true" {
		return SystemCircleCI
	}
	if os.Getenv("BITBUCKET_BUILD_NUMBER") != "" {
		return SystemBitbucket
	}
	if os.Getenv("TF_BUILD") == "True" {
		return SystemAzure
	}
	if os.Getenv("TRAVIS") == "true" {
		return SystemTravis
	}
	return SystemUnknown
}

// detectGitHub extracts values from GitHub Actions environment
func detectGitHub(d *DetectedValues) {
	d.SCMRepository = os.Getenv("GITHUB_REPOSITORY")
	d.SCMSha = os.Getenv("GITHUB_SHA")
	d.SCMBranch = os.Getenv("GITHUB_REF_NAME")
	d.InvokeID = os.Getenv("GITHUB_RUN_ID")
	d.BuildNumber = os.Getenv("GITHUB_RUN_NUMBER")
	d.BuiltBy = os.Getenv("GITHUB_ACTOR")

	// Build URL
	serverURL := os.Getenv("GITHUB_SERVER_URL")
	repo := os.Getenv("GITHUB_REPOSITORY")
	runID := os.Getenv("GITHUB_RUN_ID")
	if serverURL != "" && repo != "" && runID != "" {
		d.BuildURL = fmt.Sprintf("%s/%s/actions/runs/%s", serverURL, repo, runID)
	}

	// Use repository name as product if not set
	if d.Product == "" && d.SCMRepository != "" {
		parts := strings.Split(d.SCMRepository, "/")
		if len(parts) == 2 {
			d.Product = parts[1]
		}
	}

	// Use SHA as version fallback
	if d.Version == "" && d.SCMSha != "" {
		d.Version = d.SCMSha[:8] // Use short SHA
	}
}

// detectGitLab extracts values from GitLab CI environment
func detectGitLab(d *DetectedValues) {
	d.SCMRepository = os.Getenv("CI_PROJECT_PATH")
	d.SCMSha = os.Getenv("CI_COMMIT_SHA")
	d.SCMBranch = os.Getenv("CI_COMMIT_REF_NAME")
	d.InvokeID = os.Getenv("CI_PIPELINE_ID")
	d.BuildNumber = os.Getenv("CI_PIPELINE_IID")
	d.BuildURL = os.Getenv("CI_PIPELINE_URL")
	d.BuiltBy = os.Getenv("GITLAB_USER_LOGIN")
	d.BuiltByEmail = os.Getenv("GITLAB_USER_EMAIL")
	d.BuiltByName = os.Getenv("GITLAB_USER_NAME")

	// Use project path as product if not set
	if d.Product == "" && d.SCMRepository != "" {
		parts := strings.Split(d.SCMRepository, "/")
		if len(parts) > 0 {
			d.Product = parts[len(parts)-1]
		}
	}

	// Use SHA as version fallback
	if d.Version == "" && d.SCMSha != "" {
		d.Version = d.SCMSha[:8]
	}
}

// detectJenkins extracts values from Jenkins environment
func detectJenkins(d *DetectedValues) {
	d.SCMRepository = normalizeGitURL(os.Getenv("GIT_URL"))
	d.SCMSha = os.Getenv("GIT_COMMIT")
	d.SCMBranch = os.Getenv("GIT_BRANCH")
	d.BuildNumber = os.Getenv("BUILD_NUMBER")
	d.InvokeID = os.Getenv("BUILD_ID")
	d.BuildURL = os.Getenv("BUILD_URL")
	d.BuiltBy = os.Getenv("BUILD_USER")
	d.BuiltByEmail = os.Getenv("BUILD_USER_EMAIL")

	// Extract product from repository URL
	if d.Product == "" && d.SCMRepository != "" {
		parts := strings.Split(d.SCMRepository, "/")
		if len(parts) > 0 {
			d.Product = strings.TrimSuffix(parts[len(parts)-1], ".git")
		}
	}

	// Use build number as version fallback
	if d.Version == "" && d.BuildNumber != "" {
		d.Version = d.BuildNumber
	}
}

// detectCircleCI extracts values from CircleCI environment
func detectCircleCI(d *DetectedValues) {
	username := os.Getenv("CIRCLE_PROJECT_USERNAME")
	reponame := os.Getenv("CIRCLE_PROJECT_REPONAME")
	if username != "" && reponame != "" {
		d.SCMRepository = fmt.Sprintf("%s/%s", username, reponame)
	}

	d.SCMSha = os.Getenv("CIRCLE_SHA1")
	d.SCMBranch = os.Getenv("CIRCLE_BRANCH")
	if d.SCMBranch == "" {
		d.SCMBranch = os.Getenv("CIRCLE_TAG")
	}
	d.BuildNumber = os.Getenv("CIRCLE_BUILD_NUM")
	d.InvokeID = os.Getenv("CIRCLE_WORKFLOW_ID")
	d.BuildURL = os.Getenv("CIRCLE_BUILD_URL")
	d.BuiltBy = os.Getenv("CIRCLE_USERNAME")

	// Use repo name as product
	if d.Product == "" && reponame != "" {
		d.Product = reponame
	}

	// Use SHA as version fallback
	if d.Version == "" && d.SCMSha != "" {
		d.Version = d.SCMSha[:8]
	}
}

// detectBitbucket extracts values from Bitbucket Pipelines environment
func detectBitbucket(d *DetectedValues) {
	d.SCMRepository = os.Getenv("BITBUCKET_REPO_FULL_NAME")
	d.SCMSha = os.Getenv("BITBUCKET_COMMIT")
	d.SCMBranch = os.Getenv("BITBUCKET_BRANCH")
	if d.SCMBranch == "" {
		d.SCMBranch = os.Getenv("BITBUCKET_TAG")
	}
	d.BuildNumber = os.Getenv("BITBUCKET_BUILD_NUMBER")
	d.InvokeID = os.Getenv("BITBUCKET_PIPELINE_UUID")

	// Build URL
	repoFullName := os.Getenv("BITBUCKET_REPO_FULL_NAME")
	buildNum := os.Getenv("BITBUCKET_BUILD_NUMBER")
	if repoFullName != "" && buildNum != "" {
		d.BuildURL = fmt.Sprintf("https://bitbucket.org/%s/pipelines/results/%s", repoFullName, buildNum)
	}

	// Use repo slug as product
	repoSlug := os.Getenv("BITBUCKET_REPO_SLUG")
	if d.Product == "" && repoSlug != "" {
		d.Product = repoSlug
	}

	// Use SHA as version fallback
	if d.Version == "" && d.SCMSha != "" {
		d.Version = d.SCMSha[:8]
	}
}

// detectAzure extracts values from Azure DevOps environment
func detectAzure(d *DetectedValues) {
	d.SCMRepository = os.Getenv("BUILD_REPOSITORY_NAME")
	d.SCMSha = os.Getenv("BUILD_SOURCEVERSION")
	d.SCMBranch = os.Getenv("BUILD_SOURCEBRANCHNAME")
	d.BuildNumber = os.Getenv("BUILD_BUILDNUMBER")
	d.InvokeID = os.Getenv("BUILD_BUILDID")
	d.BuildURL = os.Getenv("BUILD_BUILDURI")
	d.BuiltBy = os.Getenv("BUILD_REQUESTEDFOR")
	d.BuiltByEmail = os.Getenv("BUILD_REQUESTEDFOREMAIL")

	// Use repository name as product
	if d.Product == "" && d.SCMRepository != "" {
		parts := strings.Split(d.SCMRepository, "/")
		if len(parts) > 0 {
			d.Product = parts[len(parts)-1]
		}
	}

	// Use build number as version fallback
	if d.Version == "" && d.BuildNumber != "" {
		d.Version = d.BuildNumber
	}
}

// detectTravis extracts values from Travis CI environment
func detectTravis(d *DetectedValues) {
	d.SCMRepository = os.Getenv("TRAVIS_REPO_SLUG")
	d.SCMSha = os.Getenv("TRAVIS_COMMIT")
	d.SCMBranch = os.Getenv("TRAVIS_BRANCH")
	if d.SCMBranch == "" {
		d.SCMBranch = os.Getenv("TRAVIS_TAG")
	}
	d.BuildNumber = os.Getenv("TRAVIS_BUILD_NUMBER")
	d.InvokeID = os.Getenv("TRAVIS_BUILD_ID")
	d.BuildURL = os.Getenv("TRAVIS_BUILD_WEB_URL")

	// Use repo name as product
	if d.Product == "" && d.SCMRepository != "" {
		parts := strings.Split(d.SCMRepository, "/")
		if len(parts) == 2 {
			d.Product = parts[1]
		}
	}

	// Use SHA as version fallback
	if d.Version == "" && d.SCMSha != "" {
		d.Version = d.SCMSha[:8]
	}
}

// normalizeGitURL removes https:// and .git suffix from Git URLs
func normalizeGitURL(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git@")
	url = strings.TrimSuffix(url, ".git")

	// Convert git@github.com:owner/repo to github.com/owner/repo
	url = strings.Replace(url, ":", "/", 1)

	return url
}
