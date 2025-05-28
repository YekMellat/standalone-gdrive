// Package dircache provides a simple cache for caching directory ID
// to path lookups and the inverse.
package dircache

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// DirCacher describes an interface for doing the low level directory work
//
// This should be implemented by the backend and will be called by the
// dircache package when appropriate.
type DirCacher interface {
	FindLeaf(ctx context.Context, pathID, leaf string) (pathIDOut string, found bool, err error)
	CreateDir(ctx context.Context, pathID, leaf string) (newID string, err error)
}

// DirCache caches paths to directory IDs and vice versa
type DirCache struct {
	cacheMu  sync.RWMutex // protects cache and invCache
	cache    map[string]string
	invCache map[string]string

	mu           sync.Mutex // protects the below
	fs           DirCacher  // Interface to find and make directories
	trueRootID   string     // ID of the absolute root
	root         string     // the path the cache is rooted on
	rootID       string     // ID of the root directory
	rootParentID string     // ID of the root's parent directory
	foundRoot    bool       // Whether we have found the root or not
}

// New makes a DirCache
//
// This is created with the true root ID and the root path.
//
// In order to use the cache FindRoot() must be called on it without
// error. This isn't done at initialization as it isn't known whether
// the root and intermediate directories need to be created or not.
func New(root, rootID string, fs DirCacher) *DirCache {
	d := &DirCache{
		fs:         fs,
		root:       root,
		rootID:     rootID,
		trueRootID: rootID,
		cache:      make(map[string]string),
		invCache:   make(map[string]string),
	}
	return d
}

// Get an ID given a path
func (dc *DirCache) Get(path string) (id string, ok bool) {
	dc.cacheMu.RLock()
	id, ok = dc.cache[path]
	dc.cacheMu.RUnlock()
	return id, ok
}

// GetInv gets a path given an ID
func (dc *DirCache) GetInv(id string) (path string, ok bool) {
	dc.cacheMu.RLock()
	path, ok = dc.invCache[id]
	dc.cacheMu.RUnlock()
	return path, ok
}

// Put a path, id in the cache
func (dc *DirCache) Put(path, id string) {
	dc.cacheMu.Lock()
	dc.cache[path] = id
	dc.invCache[id] = path
	dc.cacheMu.Unlock()
}

// Finds the actual root from the root path and rootID
//
// # Call this first and call the functions given back
//
// Returns the directoryID for the root
func (dc *DirCache) FindRoot(ctx context.Context) (string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if dc.foundRoot {
		return dc.rootID, nil
	}
	rootID, err := dc._findRoot(ctx, dc.root)
	if err != nil {
		return "", err
	}
	dc.foundRoot = true
	dc.rootID = rootID
	return rootID, nil
}

// _findRoot finds the root directory if it isn't already found
//
// Call with the root directory you wish to find.
//
// It returns the ID of the root directory.
//
// Note that you can choose a sub-directory as the root of your remote,
// so this may or may not check the actual root of the drive.
//
// This should be called with the lock held
func (dc *DirCache) _findRoot(ctx context.Context, root string) (string, error) {
	if root == "" {
		return dc.rootID, nil
	}

	// Split the root path into directories
	directories := strings.Split(root, "/")
	// Find the first parent directory that exists
	parentID := dc.rootID
	foundAll := true
	for i := range directories {
		if directories[i] == "" {
			continue
		}
		dirPath := filepath.Join(directories[:i+1]...)
		if dirID, ok := dc.Get(dirPath); ok {
			// Found in the cache
			parentID = dirID
		} else {
			foundAll = false
			break
		}
	}
	if foundAll {
		return parentID, nil
	}

	// Find all the directories that don't exist
	var lastPath string
	for i := range directories {
		if directories[i] == "" {
			continue
		}
		dirPath := filepath.Join(directories[:i+1]...)
		if _, ok := dc.Get(dirPath); ok {
			continue
		} // This directory needs to be created
		leaf := directories[i]
		parentPath := filepath.Join(directories[:i]...)
		parentID, ok := dc.Get(parentPath)
		if !ok {
			return "", fmt.Errorf("couldn't find parent directory: %s", parentPath)
		}
		dirID, err := dc.fs.CreateDir(ctx, parentID, leaf)
		if err != nil {
			return "", fmt.Errorf("failed to make directory %q: %w", dirPath, err)
		}
		dc.Put(dirPath, dirID)
		lastPath = dirPath
	}
	id, ok := dc.Get(lastPath)
	if !ok {
		return "", errors.New("internal error: couldn't find lastPath in the cache")
	}
	return id, nil
}

// FindDir finds the directory passed in returning the directory ID
// starting from pathID
//
// path should be a directory path either "" or "dir" or "dir/dir2"
func (dc *DirCache) FindDir(ctx context.Context, path string) (string, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if !dc.foundRoot {
		return "", errors.New("internal error: FindRoot not called")
	}
	if path == "" {
		return dc.rootID, nil
	}
	parts := strings.Split(path, "/")
	var partsToFind []string
	parentID := dc.rootID
	dirPath := ""
	for _, part := range parts {
		dirPath = filepath.Join(dirPath, part)
		if dirID, ok := dc.Get(dirPath); ok {
			// Found in the cache
			parentID = dirID
		} else {
			partsToFind = append(partsToFind, part)
		}
	}

	for _, part := range partsToFind {
		dirID, found, err := dc.fs.FindLeaf(ctx, parentID, part)
		if err != nil {
			return "", err
		}
		if !found {
			return "", fmt.Errorf("couldn't find directory %q", path)
		}
		parentID = dirID
		dirPath = filepath.Join(dirPath, part)
		dc.Put(dirPath, dirID)
	}
	return parentID, nil
}

// FindPath finds the path for the directory ID passed in
func (dc *DirCache) FindPath(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", errors.New("can't find path for empty ID")
	}
	if id == dc.rootID || id == dc.trueRootID {
		return dc.root, nil
	}

	path, ok := dc.GetInv(id)
	if ok {
		return path, nil
	}

	// Need to check the existence
	return "", fmt.Errorf("couldn't find path for ID %q", id)
}
