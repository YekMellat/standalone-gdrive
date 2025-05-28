// Package drive implements a Google Drive client for standalone usage
//
// This file contains shared resource functions
package drive

import (
	"context"
	"fmt"

	"google.golang.org/api/drive/v3"
)

// getResourceKey retrieves the resource key for a file or directory if available
func (f *Fs) getResourceKey(ctx context.Context, id string) string {
	// Check if we already have the resource key cached
	if value, ok := f.dirResourceKeys.Load(id); ok {
		if resourceKey, ok := value.(string); ok {
			return resourceKey
		}
	}
	// Query for the file to get the resource key
	var response *drive.File
	err := f.pacer.Call(ctx, func() error {
		var err error
		response, err = f.svc.Files.Get(id).
			Fields("resourceKey").
			SupportsAllDrives(f.isTeamDrive).
			Context(ctx).
			Do()
		return err
	})

	if err != nil {
		// If there's an error, return empty string
		return ""
	}

	// If we got a resource key, cache it
	if response.ResourceKey != "" {
		f.dirResourceKeys.Store(id, response.ResourceKey)
	}

	return response.ResourceKey
}

// addResourceKey adds a resource key to a file/folder ID if necessary
func (f *Fs) addResourceKey(id, resourceKey string) string {
	if resourceKey != "" {
		return fmt.Sprintf("%s:%s", id, resourceKey)
	}
	return id
}

// applyResourceKey applies a resource key to a file/folder ID if necessary
func (f *Fs) applyResourceKey(ctx context.Context, id string) string {
	resourceKey := f.getResourceKey(ctx, id)
	return f.addResourceKey(id, resourceKey)
}
