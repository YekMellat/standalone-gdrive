package drive

import (
	"context"

	"github.com/standalone-gdrive/fs"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// system metadata keys which this backend owns
var systemMetadataInfo = map[string]fs.MetadataHelp{
	"content-type": {
		Help:    "The MIME type of the file.",
		Type:    "string",
		Example: "text/plain",
	},
	"mtime": {
		Help:    "Time of last modification with mS accuracy.",
		Type:    "RFC 3339",
		Example: "2006-01-02T15:04:05.999Z07:00",
	},
	"btime": {
		Help:    "Time of file birth (creation) with mS accuracy. Note that this is only writable on fresh uploads - it can't be written for updates.",
		Type:    "RFC 3339",
		Example: "2006-01-02T15:04:05.999Z07:00",
	},
	"copy-requires-writer-permission": {
		Help:    "Whether the options to copy, print, or download this file, should be disabled for readers and commenters.",
		Type:    "boolean",
		Example: "true",
	},
	"writers-can-share": {
		Help:    "Whether users with only writer permission can modify the file's permissions. Not populated and ignored when setting for items in shared drives.",
		Type:    "boolean",
		Example: "false",
	},
}

// Metadata returns metadata for an object
func (o *Object) Metadata(ctx context.Context) (metadata fs.Metadata, err error) {
	var info *drive.File
	err = o.fs.pacer.Call(ctx, func() error {
		info, err = o.fs.svc.Files.Get(o.id).Fields("*").SupportsAllDrives(o.fs.isTeamDrive).Do()
		return err
	})
	if err != nil {
		return nil, err
	}

	// Construct metadata
	metadata = fs.Metadata{
		"content-type": info.MimeType,
		"mtime":        info.ModifiedTime,
	}

	// Add creation time if available
	if info.CreatedTime != "" {
		metadata["btime"] = info.CreatedTime
	}

	// Add permissions info if available
	if info.CopyRequiresWriterPermission {
		metadata["copy-requires-writer-permission"] = "true"
	}
	if info.WritersCanShare {
		metadata["writers-can-share"] = "true"
	}

	// Add user metadata from properties
	if info.Properties != nil {
		for k, v := range info.Properties {
			metadata["user."+k] = v
		}
	}

	return metadata, nil
}

// SetMetadata sets metadata for an object
func (o *Object) SetMetadata(ctx context.Context, metadata fs.Metadata) error {
	// Create update info
	updateInfo := &drive.File{}

	// Process system metadata
	for k, v := range metadata {
		switch k {
		case "content-type":
			updateInfo.MimeType = v
		case "mtime":
			updateInfo.ModifiedTime = v
		case "btime":
			updateInfo.CreatedTime = v
		case "copy-requires-writer-permission":
			updateInfo.CopyRequiresWriterPermission = v == "true"
		case "writers-can-share":
			updateInfo.WritersCanShare = v == "true"
		}
	}

	// Process user metadata
	properties := map[string]string{}
	for k, v := range metadata {
		if len(k) > 5 && k[:5] == "user." {
			properties[k[5:]] = v
		}
	}
	if len(properties) > 0 {
		updateInfo.Properties = properties
	}
	// Send update
	err := o.fs.pacer.Call(ctx, func() error {
		_, err := o.fs.svc.Files.Update(o.id, updateInfo).
			Fields(googleapi.Field(partialFields)).
			SupportsAllDrives(o.fs.isTeamDrive).
			Do()
		return err
	})

	return err
}

// Metadata for directories
func (d *Directory) Metadata(ctx context.Context) (fs.Metadata, error) {
	var info *drive.File
	err := d.fs.pacer.Call(ctx, func() error {
		var err error
		info, err = d.fs.svc.Files.Get(d.id).Fields("*").SupportsAllDrives(d.fs.isTeamDrive).Do()
		return err
	})
	if err != nil {
		return nil, err
	}

	// Construct metadata
	metadata := fs.Metadata{
		"mtime": info.ModifiedTime,
	}

	// Add creation time if available
	if info.CreatedTime != "" {
		metadata["btime"] = info.CreatedTime
	}

	// Add user metadata from properties
	if info.Properties != nil {
		for k, v := range info.Properties {
			metadata["user."+k] = v
		}
	}

	return metadata, nil
}

// SetMetadata sets metadata for a Directory
func (d *Directory) SetMetadata(ctx context.Context, metadata fs.Metadata) error {
	// Create update info
	updateInfo := &drive.File{}

	// Process system metadata
	for k, v := range metadata {
		switch k {
		case "mtime":
			updateInfo.ModifiedTime = v
		case "btime":
			updateInfo.CreatedTime = v
		}
	}

	// Process user metadata
	properties := map[string]string{}
	for k, v := range metadata {
		if len(k) > 5 && k[:5] == "user." {
			properties[k[5:]] = v
		}
	}
	if len(properties) > 0 {
		updateInfo.Properties = properties
	}
	// Send update
	err := d.fs.pacer.Call(ctx, func() error {
		_, err := d.fs.svc.Files.Update(d.id, updateInfo).
			Fields(googleapi.Field(partialFields)).
			SupportsAllDrives(d.fs.isTeamDrive).
			Do()
		return err
	})

	return err
}
