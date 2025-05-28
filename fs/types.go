// Package fs defines the core interfaces that the drive package depends on
package fs

import (
	"context"
	"errors"
	"io"
	"math"
	"time"

	"github.com/standalone-gdrive/fs/hash"
)

// Constants
const (
	// ModTimeNotSupported is a very large precision value to show
	// mod time isn't supported on this Fs
	ModTimeNotSupported = 100 * 365 * 24 * time.Hour
	// MaxLevel is a sentinel representing an infinite depth for listings
	MaxLevel = math.MaxInt32
	// LinkSuffix is the suffix added to a translated symbolic link
	LinkSuffix = ".rclonelink"
)

// Common errors
var (
	ErrorDirNotFound          = errors.New("directory not found")
	ErrorObjectNotFound       = errors.New("object not found")
	ErrorIsDir                = errors.New("is a directory not a file")
	ErrorNotDir               = errors.New("not a directory")
	ErrorCantUploadEmptyFiles = errors.New("can't upload empty files")
	ErrorPermissionDenied     = errors.New("permission denied")
	ErrorNotImplemented       = errors.New("optional feature not implemented")
	ErrorLimitExceeded        = errors.New("limit exceeded")
	ErrorCantMove             = errors.New("can't move")
	ErrorCantCopy             = errors.New("can't copy")
	ErrorCantDirMove          = errors.New("can't move directory")
	ErrorDirExists            = errors.New("directory already exists")
	ErrorCantShareDirectories = errors.New("can't share directories")
	ErrorNotAnObject          = errors.New("not an object")
	ErrorNotDeleted           = errors.New("not deleted")
	ErrorCantUpdate           = errors.New("can't update")
	ErrorDirectoryNotEmpty    = errors.New("directory not empty")
)

// Fs is the interface a cloud storage system must provide
type Fs interface {
	Info

	// List the objects and directories in dir into entries
	List(ctx context.Context, dir string) (entries DirEntries, err error)

	// NewObject finds the Object at remote
	NewObject(ctx context.Context, remote string) (Object, error)

	// Put in to the remote path with the modTime given of the given size
	Put(ctx context.Context, in io.Reader, src ObjectInfo, options ...OpenOption) (Object, error)

	// Mkdir makes the directory (container, bucket)
	Mkdir(ctx context.Context, dir string) error

	// Rmdir removes the directory (container, bucket) if empty
	Rmdir(ctx context.Context, dir string) error
}

// Info provides a read only interface to information about a filesystem.
type Info interface {
	// Name of the remote (as passed into NewFs)
	Name() string

	// Root of the remote (as passed into NewFs)
	Root() string

	// String returns a description of the FS
	String() string

	// Precision of the ModTimes in this Fs
	Precision() time.Duration

	// Returns the supported hash types of the filesystem
	Hashes() hash.Set

	// Features returns the optional features of this Fs
	Features() *Features
}

// Object is a filesystem like object provided by an Fs
type Object interface {
	ObjectInfo

	// SetModTime sets the metadata on the object to set the modification date
	SetModTime(ctx context.Context, t time.Time) error

	// Open opens the file for read.  Call Close() on the returned io.ReadCloser
	Open(ctx context.Context, options ...OpenOption) (io.ReadCloser, error)

	// Update in to the object with the modTime given of the given size
	Update(ctx context.Context, in io.Reader, src ObjectInfo, options ...OpenOption) error

	// Removes this object
	Remove(ctx context.Context) error
}

// ObjectInfo provides read only information about an object
type ObjectInfo interface {
	DirEntry

	// Hash returns the selected checksum of the file
	Hash(ctx context.Context, ty hash.Type) (string, error)

	// Storable says whether this object can be stored
	Storable() bool
}

// DirEntry provides read only information about the common subset of
// a Dir or Object
type DirEntry interface {
	// Fs returns read only access to the Fs that this object is part of
	Fs() Info

	// String returns a description of the Object
	String() string

	// Remote returns the remote path
	Remote() string

	// ModTime returns the modification date of the file
	ModTime(context.Context) time.Time

	// Size returns the size of the file
	Size() int64
}

// Directory is a filesystem like directory provided by an Fs
type Directory interface {
	DirEntry

	// Items returns the count of items in this directory or this
	// directory and subdirectories if known, -1 for unknown
	Items() int64

	// ID returns the internal ID of this directory if known, or
	// "" otherwise
	ID() string
}
