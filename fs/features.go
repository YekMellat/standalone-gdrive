package fs

import (
	"context"
	"io"
)

// Features describe the optional features of the Fs
type Features struct {
	// Feature flags, whether Fs supports this feature
	CaseInsensitive         bool // Not really optional but here for convenience
	DuplicateFiles          bool // Allows duplicate files
	ReadMimeType            bool // Can read the mime type of objects
	WriteMimeType           bool // Can set the mime type of objects
	CanHaveEmptyDirectories bool // Can have empty directories
	ServerSideAcrossConfigs bool // Can server-side copy between different remotes
	FilterAware             bool // Filesystems that can apply filter rules on listing
	ReadMetadata            bool // Can read metadata
	WriteMetadata           bool // Can write metadata
	UserMetadata            bool // Can read/write user metadata
	ReadDirMetadata         bool // Can read directory metadata
	WriteDirMetadata        bool // Can write directory metadata
	UserDirMetadata         bool // Can read/write user metadata on directories
	WriteDirSetModTime      bool // Can set mod time of directory

	// Purge all files in the directory specified
	Purge func(ctx context.Context, dir string) error

	// Copy src to this remote using server-side copy if possible
	Copy func(ctx context.Context, src Object, remote string) (Object, error)

	// Move src to this remote using server-side move if possible
	Move func(ctx context.Context, src Object, remote string) (Object, error)

	// DirMove moves src, srcRemote to this remote at dstRemote
	// using server-side move if possible
	DirMove func(ctx context.Context, src Fs, srcRemote, dstRemote string) error

	// ChangeNotify calls the passed function with a path
	// that has had changes
	ChangeNotify func(ctx context.Context, notifyFunc func(string, EntryType)) chan bool

	// UnWrap returns the Fs that this Fs is wrapping
	UnWrap func() Fs

	// PutUnchecked uploads the object
	//
	// This will create a duplicate if we upload a new file without
	// checking to see if there is one already
	PutUnchecked func(ctx context.Context, in io.Reader, src ObjectInfo, options ...OpenOption) (Object, error)

	// PutStream uploads the object with a stream from a reader
	PutStream func(ctx context.Context, in io.Reader, src ObjectInfo, options ...OpenOption) (Object, error)

	// MergeDirs merges the contents of all the directories passed
	// in into the first one and rmdirs the other directories.
	MergeDirs func(ctx context.Context, dirs []Directory) error
}

// Fill fills in the function pointers in the Features struct from the
// optional interfaces.  It returns the original updated Features
// struct passed in.
func (ftrs *Features) Fill(ctx context.Context, f Fs) *Features {
	if do, ok := f.(Purger); ok {
		ftrs.Purge = do.Purge
	}
	if do, ok := f.(Copier); ok {
		ftrs.Copy = do.Copy
	}
	if do, ok := f.(Mover); ok {
		ftrs.Move = do.Move
	}
	if do, ok := f.(DirMover); ok {
		ftrs.DirMove = do.DirMove
	}
	if do, ok := f.(ChangeNotifier); ok {
		ftrs.ChangeNotify = do.ChangeNotify
	}
	if do, ok := f.(UnWrapper); ok {
		ftrs.UnWrap = do.UnWrap
	}
	if do, ok := f.(PutUncheckeder); ok {
		ftrs.PutUnchecked = do.PutUnchecked
	}
	if do, ok := f.(PutStreamer); ok {
		ftrs.PutStream = do.PutStream
	}
	if do, ok := f.(MergeDirser); ok {
		ftrs.MergeDirs = do.MergeDirs
	}
	return ftrs
}

// EntryType can be associated with remote paths to identify their type
type EntryType int

// Constants
const (
	// EntryDirectory should be used to classify remote paths in directories
	EntryDirectory EntryType = iota // 0
	// EntryObject should be used to classify remote paths in objects
	EntryObject // 1
)

// Purger is an optional interfaces for Fs
type Purger interface {
	// Purge all files in the directory
	Purge(ctx context.Context, dir string) error
}

// Copier is an optional interface for Fs
type Copier interface {
	// Copy src to this remote using server-side copy if possible
	Copy(ctx context.Context, src Object, remote string) (Object, error)
}

// Mover is an optional interface for Fs
type Mover interface {
	// Move src to this remote using server-side move if possible
	Move(ctx context.Context, src Object, remote string) (Object, error)
}

// DirMover is an optional interface for Fs
type DirMover interface {
	// DirMove moves src, srcRemote to this remote at dstRemote
	// using server-side move if possible
	DirMove(ctx context.Context, src Fs, srcRemote, dstRemote string) error
}

// ChangeNotifier is an optional interface for Fs
type ChangeNotifier interface {
	// ChangeNotify calls the passed function with a path
	// that has had changes
	ChangeNotify(ctx context.Context, notifyFunc func(string, EntryType)) chan bool
}

// UnWrapper is an optional interfaces for Fs
type UnWrapper interface {
	// UnWrap returns the Fs that this Fs is wrapping
	UnWrap() Fs
}

// PutUncheckeder is an optional interface for Fs
type PutUncheckeder interface {
	// PutUnchecked uploads the object
	//
	// This will create a duplicate if we upload a new file without
	// checking to see if there is one already
	PutUnchecked(ctx context.Context, in io.Reader, src ObjectInfo, options ...OpenOption) (Object, error)
}

// PutStreamer is an optional interface for Fs
type PutStreamer interface {
	// PutStream uploads to the remote path with the modTime given of indeterminate size
	PutStream(ctx context.Context, in io.Reader, src ObjectInfo, options ...OpenOption) (Object, error)
}

// MergeDirser is an optional interface for Fs
type MergeDirser interface {
	// MergeDirs merges the contents of all the directories passed
	// in into the first one and rmdirs the other directories.
	MergeDirs(ctx context.Context, dirs []Directory) error
}
