// Package drive implements a Google Drive client for standalone usage
//
// This file contains path utilities
package drive

import (
	"path"
	"strings"
)

// parseDrivePathDetailed converts a user-visible path to a cleaned path without a
// leading or trailing '/' with additional validation
func parseDrivePathDetailed(p string) (string, error) {
	// Clean the path
	p = strings.Trim(path.Clean(p), "/")

	// Check if path contains invalid characters like '*'
	// Google Drive API doesn't allow certain special characters
	if strings.ContainsAny(p, "*?|<>:") {
		return "", InvalidCharacters
	}

	return p, nil
}

// splitPath splits a remote path into directory and leaf components
func splitPath(p string) (dir, leaf string) {
	p = strings.Trim(p, "/")
	i := strings.LastIndex(p, "/")
	if i >= 0 {
		return p[:i], p[i+1:]
	}
	return "", p
}

// isRootDirectory returns true if this is the root directory
func isRootDirectory(p string) bool {
	return p == "" || p == "/" || p == "."
}

// getRelativePath returns a path relative to the root
// It returns the empty string if the path equals the root
func getRelativePath(root, p string) string {
	if strings.HasPrefix(p, root) {
		p = p[len(root):]
	}
	return strings.TrimPrefix(p, "/")
}

// joinPath joins directory and leaf into a path
func joinPath(dir, leaf string) string {
	// If directory is empty, just return the leaf
	if dir == "" {
		return leaf
	}
	// Join with a slash
	return dir + "/" + leaf
}
