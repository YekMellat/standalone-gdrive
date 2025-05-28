// Package fs provides the core interfaces for filesystem-like operations
package fs

import (
	"context"
	"fmt"
	"time"
)

// OpenOption is an option for Open
type OpenOption interface {
	Apply(string) error
}

// SizeSuffix represents a file size in bytes
type SizeSuffix int64

// String returns the size as a string
func (s SizeSuffix) String() string {
	return fmt.Sprintf("%d", s)
}

// Byte returns the size in bytes
func (s SizeSuffix) Byte() int64 {
	return int64(s)
}

// Instance of Size suffixes
const (
	Byte   SizeSuffix = 1
	KiByte            = Byte * 1024
	MiByte            = KiByte * 1024
	GiByte            = MiByte * 1024
	TiByte            = GiByte * 1024
	PiByte            = TiByte * 1024
	EiByte            = PiByte * 1024
)

// Counter is a 64-bit thread-safe counter
type Counter struct {
	count int64
}

// Inc adds n on to the counter
func (c *Counter) Inc(n int64) {
	if c == nil {
		return
	}
	c.count += n
}

// DirEntries is a slice of Object or *Dir
type DirEntries []DirEntry

// Len returns the length of entries
func (es DirEntries) Len() int {
	return len(es)
}

// Swap swaps entries
func (es DirEntries) Swap(i, j int) {
	es[i], es[j] = es[j], es[i]
}

// Less reports whether es[i] is less than es[j]
func (es DirEntries) Less(i, j int) bool {
	return es[i].Remote() < es[j].Remote()
}

// String returns a string representation of all the DirEntries
func (es DirEntries) String() string {
	items := make([]string, 0, len(es))
	for _, e := range es {
		items = append(items, e.String())
	}
	return fmt.Sprintf("%s", items)
}

// Duration is a time.Duration that serializes to JSON as human-readable
type Duration time.Duration

// String returns the duration as a string
func (d Duration) String() string {
	return time.Duration(d).String()
}

// RegInfo is info about a backend
type RegInfo struct {
	Name        string           // name of the backend
	Description string           // description of the backend
	NewFs       NewFsFunc        // create a new Fs object
	Options     OptionDefinition // options definition
}

// NewFsFunc is a function that creates a new filesystem
type NewFsFunc func(context.Context, string, string, Metadata) (Fs, error)

// Metadata is a simple key-value store
type Metadata map[string]string

// Pacer is a rate limiter for operations
type Pacer struct {
	calculateDelay func(state PacerState) time.Duration
	maxConnections int
	connTokens     chan struct{}
	retries        int
}

// NewPacer creates a Pacer with the given retries and max connections
func NewPacer(ctx context.Context, calculateDelay func(state PacerState) time.Duration) *Pacer {
	pacer := &Pacer{
		calculateDelay: calculateDelay,
		maxConnections: 1,
		retries:        3,
	}
	pacer.connTokens = make(chan struct{}, pacer.maxConnections)
	for i := 0; i < pacer.maxConnections; i++ {
		pacer.connTokens <- struct{}{}
	}
	return pacer
}

// PacerState represents the state of a Pacer
type PacerState struct {
	ConsecutiveRetries int
	LastError          error
}

// Call calls the supplied function, using rate limiting and retrying
func (p *Pacer) Call(ctx context.Context, fn func() error) error {
	var err error
	for try := 0; try <= p.retries; try++ {
		// Get a token
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.connTokens:
			// Got a token
		}
		// Do the operation
		err = fn()
		// Return the token
		p.connTokens <- struct{}{}
		if err == nil {
			break
		}
		if try >= p.retries {
			break
		}
		// Delay before retrying
		delay := p.calculateDelay(PacerState{
			ConsecutiveRetries: try,
			LastError:          err,
		})
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(delay)):
		}
	}
	return err
}

// ConfigInfo is a structure containing a remote name and a path
type ConfigInfo struct {
	NoRetries      bool // if set to true this will cause the pacer not to retry on error
	MaxConnections int  // Maximum number of concurrent connections
	Transfers      int  // Number of additional objects to transfer at the same time
}

// GetConfig gets the config from the context
func GetConfig(ctx context.Context) *ConfigInfo {
	// If no context, return defaults
	if ctx == nil {
		return &ConfigInfo{
			NoRetries:      false,
			MaxConnections: 4,
			Transfers:      4,
		}
	}
	// Check for config in context
	ci, ok := ctx.Value(configKey).(*ConfigInfo)
	if !ok {
		return &ConfigInfo{
			NoRetries:      false,
			MaxConnections: 4,
			Transfers:      4,
		}
	}
	return ci
}

type contextKey string

const configKey contextKey = "rclone.configInfo"
