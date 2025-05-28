package fs

import (
	"context"
)

// MetadataHelp contains specifics on metadata keys
type MetadataHelp struct {
	Help     string // Help text, markdown formatted
	Type     string // Type of data in the key
	Example  string // Example value
	ReadOnly bool   // True if this can't be set
}

// MetadataInfo contains info about the metadata for this backend
type MetadataInfo struct {
	System map[string]MetadataHelp // System metadata
	Help   string                  // Help text, markdown formatted
}

// Metadataer is an optional interface for Object
type Metadataer interface {
	// Metadata returns metadata for an object
	Metadata(ctx context.Context) (Metadata, error)
}

// SetMetadataer is an optional interface for Object
type SetMetadataer interface {
	// SetMetadata sets metadata for an object
	SetMetadata(ctx context.Context, metadata Metadata) error
}

// Keys are case-insensitive in Metadata

// Get gets a metadata key
//
// It returns the value or "" if not found
func (m Metadata) Get(key string) string {
	if m == nil {
		return ""
	}
	return m[key]
}

// Set sets a metadata key
func (m Metadata) Set(key, value string) {
	if m == nil {
		return
	}
	m[key] = value
}

// DeleteKey removes a key from the metadata
func (m Metadata) DeleteKey(key string) {
	if m == nil {
		return
	}
	delete(m, key)
}

// Equal returns true if metadata are equal
func (m Metadata) Equal(other Metadata) bool {
	if len(m) != len(other) {
		return false
	}
	for k, v := range m {
		if other[k] != v {
			return false
		}
	}
	return true
}

// Copy returns a deep copy of the metadata
func (m Metadata) Copy() Metadata {
	newM := make(Metadata, len(m))
	for k, v := range m {
		newM[k] = v
	}
	return newM
}
