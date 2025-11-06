package version

// Version information - injected at build time via ldflags
var (
	// Version is the semantic version (e.g., "1.0.0")
	Version = "dev"

	// Commit is the git commit SHA
	Commit = "unknown"

	// BuildDate is the build timestamp
	BuildDate = "unknown"
)

// GetVersion returns the full version string
func GetVersion() string {
	if Version == "dev" {
		return "dev"
	}
	return Version
}

// GetUserAgent returns the User-Agent string for API requests
func GetUserAgent() string {
	return "versioner-cli/" + GetVersion()
}
