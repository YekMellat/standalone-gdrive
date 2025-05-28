// Package fs defines the core interfaces that the drive package depends on
package fs

import (
	"context"
	"time"

	"github.com/standalone-gdrive/fs/hash"
)

// ObjectInfoImpl is a simple implementation of the ObjectInfo interface
type ObjectInfoImpl struct {
	RemoteName  string
	FileSize    int64
	FileModTime time.Time
	Hashes      map[hash.Type]string
}

// Fs returns nil as this is not part of an Fs
func (o *ObjectInfoImpl) Fs() Info {
	return nil
}

// String returns a description of the ObjectInfo
func (o *ObjectInfoImpl) String() string {
	return o.RemoteName
}

// Remote returns the remote path
func (o *ObjectInfoImpl) Remote() string {
	return o.RemoteName
}

// Hash returns the requested hash
func (o *ObjectInfoImpl) Hash(_ context.Context, ht hash.Type) (string, error) {
	if o.Hashes == nil {
		return "", nil
	}
	hashVal, ok := o.Hashes[ht]
	if !ok {
		return "", hash.ErrUnsupported
	}
	return hashVal, nil
}

// ModTime returns the modification date of the file
func (o *ObjectInfoImpl) ModTime(_ context.Context) time.Time {
	return o.FileModTime
}

// Size returns the size of the file
func (o *ObjectInfoImpl) Size() int64 {
	return o.FileSize
}

// Storable returns whether this object can be stored
func (o *ObjectInfoImpl) Storable() bool {
	return true
}
