// Package drive implements the Google Drive backend for standalone-gdrive
package drive

import (
	"errors"
)

// Error codes specific to drive
var (
	// ErrFileNotFound is returned when a file can't be found in Google Drive
	ErrFileNotFound = errors.New("file not found")

	// ErrDirectoryNotFound is returned when a directory can't be found
	ErrDirectoryNotFound = errors.New("directory not found")

	// ErrPermissionDenied is returned when the operation is not permitted
	ErrPermissionDenied = errors.New("permission denied")

	// ErrUploadFailed is returned when an upload operation fails
	ErrUploadFailed = errors.New("upload failed")

	// ErrDownloadFailed is returned when a download operation fails
	ErrDownloadFailed = errors.New("download failed")

	// ErrInvalidAuthToken is returned when authentication fails
	ErrInvalidAuthToken = errors.New("invalid authentication token")
)
