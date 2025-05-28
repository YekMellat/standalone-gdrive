// Package drive implements a Google Drive client for standalone usage
//
// This file contains the upload implementation
package drive

import (
	"context"
	"io"

	"google.golang.org/api/drive/v3"
)

// uploadChunkedDetailed uploads a file using the Google Drive API
func (f *Fs) uploadChunkedDetailed(ctx context.Context, in io.Reader, size int64, createInfo *drive.File) (*drive.File, error) {
	var fileInfo *drive.File

	err := f.pacer.Call(ctx, func() error {
		// Create the file using the Drive API
		call := f.svc.Files.Create(createInfo)
		call.SupportsAllDrives(f.isTeamDrive)
		call.Media(in)

		var err error
		fileInfo, err = call.Do()
		return err
	})

	if err != nil {
		return nil, err
	}

	return fileInfo, nil
}
