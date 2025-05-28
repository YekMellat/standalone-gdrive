package drive

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/standalone-gdrive/fs"
	"github.com/standalone-gdrive/fs/hash"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// ------------------------------------------------------------
// Object specific methods

// Fs returns read only access to the Fs that this object is part of
func (o *Object) Fs() fs.Info {
	return o.fs
}

// Return a string version
func (o *Object) String() string {
	if o == nil {
		return "<nil>"
	}
	return o.remote
}

// Remote returns the remote path
func (o *Object) Remote() string {
	return o.remote
}

// Hash returns the requested hash of a file
func (o *Object) Hash(ctx context.Context, t hash.Type) (string, error) {
	switch t {
	case hash.MD5:
		return o.md5sum, nil
	case hash.SHA1:
		return o.sha1sum, nil
	case hash.SHA256:
		return o.sha256sum, nil
	}
	return "", hash.ErrUnsupported
}

// Size returns the size of an object in bytes
func (o *Object) Size() int64 {
	return o.bytes
}

// MimeType returns the content type of the Object if known
func (o *Object) MimeType(ctx context.Context) string {
	return o.mimeType
}

// SetModTime sets the modification time of the drive fs object
func (o *Object) SetModTime(ctx context.Context, modTime time.Time) error {
	// New metadata
	updateInfo := &drive.File{
		ModifiedTime: modTime.Format(timeFormatOut),
	}
	// Set options
	err := o.fs.pacer.Call(ctx, func() error {
		_, err := o.fs.svc.Files.Update(o.id, updateInfo).
			Fields(googleapi.Field(partialFields)).
			SupportsAllDrives(o.fs.isTeamDrive).
			Do()
		return err
	})
	if err != nil {
		return err
	}

	// Update info
	o.modifiedDate = modTime.Format(timeFormatOut)
	return nil
}

// ModTime returns the modification time of the object
func (o *Object) ModTime(ctx context.Context) time.Time {
	modTime, err := time.Parse(timeFormatIn, o.modifiedDate)
	if err != nil {
		return time.Now()
	}
	return modTime
}

// Open an object for read
func (o *Object) Open(ctx context.Context, options ...fs.OpenOption) (io.ReadCloser, error) {
	var url string
	var resp *http.Response
	var err error
	if o.v2Download {
		// Use Drive API v2 to download the file to get a more reliable
		// download experience for large files
		err = o.fs.pacer.Call(ctx, func() (err error) {
			resp, err = o.fs.v2Svc.Files.Get(o.id).Download()
			return err
		})
		if err != nil {
			return nil, err
		}
		err = o.fs.pacer.Call(ctx, func() (err error) {
			resp, err = o.fs.client.Get(url)
			return err
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Use Drive API v3
		var downloadURL string
		if o.fs.opt.AcknowledgeAbuse {
			downloadURL = "https://www.googleapis.com/drive/v3/files/" + o.id + "?alt=media&acknowledgeAbuse=true"
		} else {
			downloadURL = "https://www.googleapis.com/drive/v3/files/" + o.id + "?alt=media"
		}
		err = o.fs.pacer.Call(ctx, func() (err error) {
			resp, err = o.fs.client.Get(downloadURL)
			return err
		})
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("bad response: %d: %s", resp.StatusCode, resp.Status)
	}

	return resp.Body, nil
}

// Update in to the object
func (o *Object) Update(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) error {
	size := src.Size()
	if size < 0 {
		size = 0
	}

	// Create a new file info
	updateInfo := &drive.File{}

	// Set modified date if set
	modTime := src.ModTime(ctx)
	updateInfo.ModifiedTime = modTime.Format(timeFormatOut)

	var info *drive.File
	var err error
	if size > int64(o.fs.opt.UploadCutoff) {
		// Upload in chunks
		info, err = o.fs.uploadChunked(ctx, in, size, updateInfo)
	} else { // Simple upload
		err = o.fs.pacer.Call(ctx, func() (err error) {
			info, err = o.fs.svc.Files.Update(o.id, updateInfo).
				Media(in, googleapi.ContentType("")).
				Fields(googleapi.Field(partialFields)).
				SupportsAllDrives(o.fs.isTeamDrive).
				KeepRevisionForever(o.fs.opt.KeepRevisionForever).
				Do()
			return err
		})
	}

	if err != nil {
		return err
	}

	// Update object
	o.id = info.Id
	o.modifiedDate = info.ModifiedTime
	o.mimeType = info.MimeType
	o.bytes = info.Size
	o.md5sum = info.Md5Checksum
	o.sha1sum = info.Sha1Checksum
	o.sha256sum = info.Sha256Checksum
	o.v2Download = o.fs.opt.V2DownloadMinSize >= 0 && info.Size >= int64(o.fs.opt.V2DownloadMinSize)

	return nil
}

// Remove an object
func (o *Object) Remove(ctx context.Context) error {
	err := o.fs.pacer.Call(ctx, func() error {
		if o.fs.opt.UseTrash {
			// Put file in the trash instead of deleting it permanently
			updateInfo := &drive.File{
				Trashed: true,
			}
			_, err := o.fs.svc.Files.Update(o.id, updateInfo).
				Fields("").
				SupportsAllDrives(o.fs.isTeamDrive).
				Do()
			return err
		}
		// Delete file permanently
		return o.fs.svc.Files.Delete(o.id).
			SupportsAllDrives(o.fs.isTeamDrive).
			Do()
	})
	return err
}

// ID gets the ID of the Object
func (o *Object) ID() string {
	return o.id
}

// ParentID gets the ID of the Object parent
func (o *Object) ParentID() string {
	if len(o.parents) > 0 {
		return o.parents[0]
	}
	return ""
}

// Storable returns whether this object is storable
func (o *Object) Storable() bool {
	return true
}

// ------------------------------------------------------------
// Directory specific methods

// Fs returns read only access to the Fs that this object is part of
func (d *Directory) Fs() fs.Info {
	return d.fs
}

// String returns a string version
func (d *Directory) String() string {
	if d == nil {
		return "<nil>"
	}
	return d.remote
}

// Remote returns the remote path
func (d *Directory) Remote() string {
	return d.remote
}

// ModTime returns the modification time
func (d *Directory) ModTime(ctx context.Context) time.Time {
	modTime, err := time.Parse(timeFormatIn, d.modifiedDate)
	if err != nil {
		return time.Now()
	}
	return modTime
}

// Size returns the size of the directory
func (d *Directory) Size() int64 {
	return 0
}

// ID gets the ID of the directory
func (d *Directory) ID() string {
	return d.id
}

// ParentID gets the ID of the directory parent
func (d *Directory) ParentID() string {
	if len(d.parents) > 0 {
		return d.parents[0]
	}
	return ""
}

// Items returns the count of items in this directory
func (d *Directory) Items() int64 {
	return -1
}
