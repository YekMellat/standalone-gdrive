// Package drive implements a Google Drive client for standalone usage
//
// This file contains version helper functions
package drive

import (
	"github.com/standalone-gdrive/version"
)

// getVersionUserAgent returns the user agent string from the version package
func getVersionUserAgent() string {
	return version.GetUserAgent()
}
