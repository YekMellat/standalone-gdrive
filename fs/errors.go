// Package fs defines the core interfaces that the drive package depends on
package fs

import (
	"errors"
)

// IsDir returns a boolean indicating whether the error is known to report that
// the entry is a directory.
func IsDir(err error) bool {
	if err == ErrorIsDir || errors.Is(err, ErrorIsDir) {
		return true
	}
	return false
}
