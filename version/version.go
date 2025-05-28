package version

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

var (
	// Version is the current version of the standalone Google Drive client
	Version = "0.2.0"

	// BuildTime will be set at build time
	BuildTime = ""

	// GitCommit will be set at build time
	GitCommit = ""

	// IsRelease will be true for release builds
	IsRelease = false
)

// GetUserAgent returns the User-Agent string used in API requests
func GetUserAgent() string {
	ua := fmt.Sprintf("standalone-gdrive/%s", Version)

	// Add build information if available
	if GitCommit != "" {
		shortCommit := GitCommit
		if len(shortCommit) > 8 {
			shortCommit = shortCommit[:8]
		}
		ua += fmt.Sprintf(" git/%s", shortCommit)
	}

	// Add runtime information
	ua += fmt.Sprintf(" go/%s %s/%s", runtime.Version()[2:], runtime.GOOS, runtime.GOARCH)

	return ua
}

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("standalone-gdrive v%s\n", Version))

	if GitCommit != "" {
		sb.WriteString(fmt.Sprintf("Git commit: %s\n", GitCommit))
	}

	if BuildTime != "" {
		// Try to parse build time
		if t, err := time.Parse(time.RFC3339, BuildTime); err == nil {
			sb.WriteString(fmt.Sprintf("Build time: %s\n", t.Format(time.RFC1123)))
		} else {
			sb.WriteString(fmt.Sprintf("Build time: %s\n", BuildTime))
		}
	}

	sb.WriteString(fmt.Sprintf("Go version: %s\n", runtime.Version()))
	sb.WriteString(fmt.Sprintf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH))

	return sb.String()
}
