// Package fs defines the core interfaces that the drive package depends on.
// This package provides the foundational interfaces and types for Google Drive operations.
package fs

import (
	"context"
	"time"
)

// IDer is an interface for objects that have a unique ID.
// Google Drive resources typically have a unique ID separate from their name.
type IDer interface {
	// ID returns the unique identifier of the object.
	// For Google Drive, this will typically be the file ID used in the API.
	ID() string
}

// MimeTyper is an interface for objects that have a MIME type.
// MIME types in Google Drive are essential for file type identification
// and handling of native Google Workspace documents.
type MimeTyper interface {
	// MimeType returns the MIME type of the object.
	// This is important for Google Workspace documents which have special MIME types.
	// For example, Google Docs have the MIME type "application/vnd.google-apps.document".
	MimeType(context.Context) string
}

// ModTimeSetter is an interface for objects that can have their
// modification time set. This allows syncing tools to preserve
// file timestamps when transferring files.
type ModTimeSetter interface {
	// SetModTime sets the modification time of the object.
	// Google Drive allows modification times to be set on files
	// although with some limitations compared to local filesystems.
	SetModTime(context.Context, time.Time) error
}

// Note: The following interfaces have been moved to features.go to avoid duplication:
// - Purger
// - Copier
// - Mover
