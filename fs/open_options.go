// Package fs defines interfaces, types, and utilities for filesystem-like operations
package fs

import (
	"fmt"
)

// HTTPOption is an interface describing HTTP options for Open
// This is different from OpenOption in fs.go
type HTTPOption interface {
	fmt.Stringer

	// Header returns the option as an HTTP header
	Header() (key string, value string)

	// Mandatory returns whether this option can be ignored or not
	Mandatory() bool
}

// RangeOption defines an HTTP Range option with start and end.  If
// either start or end are < 0 then they will be omitted.
type RangeOption struct {
	Start int64
	End   int64
}

// Header formats the option as an http header
func (o *RangeOption) Header() (key string, value string) {
	key = "Range"
	value = "bytes="
	if o.Start >= 0 {
		value += fmt.Sprint(o.Start)
	}
	value += "-"
	if o.End >= 0 {
		value += fmt.Sprint(o.End)
	}
	return key, value
}

// Apply applies this option to the string
func (o *RangeOption) Apply(s string) error {
	return nil
}

// String formats the option into human-readable form
func (o *RangeOption) String() string {
	return fmt.Sprintf("RangeOption(%d,%d)", o.Start, o.End)
}

// Mandatory returns whether the option must be parsed or can be ignored
func (o *RangeOption) Mandatory() bool {
	return true
}

// Decode interprets the RangeOption into an offset and a limit
func (o *RangeOption) Decode(size int64) (offset, limit int64) {
	if o.Start >= 0 {
		offset = o.Start
		if o.End >= 0 {
			limit = o.End - o.Start + 1
		} else {
			limit = -1
		}
	} else {
		if o.End >= 0 {
			offset = size - o.End
		} else {
			offset = 0
		}
		limit = -1
	}
	return offset, limit
}

// SeekOption defines an HTTP Range option with start only
type SeekOption struct {
	Offset int64
}

// Header formats the option as an http header
func (o *SeekOption) Header() (key string, value string) {
	key = "Range"
	value = fmt.Sprintf("bytes=%d-", o.Offset)
	return key, value
}

// String formats the option into human-readable form
func (o *SeekOption) String() string {
	return fmt.Sprintf("SeekOption(%d)", o.Offset)
}

// Mandatory returns whether the option must be parsed or can be ignored
func (o *SeekOption) Mandatory() bool {
	return true
}

// Apply implements the OpenOption interface
func (o *SeekOption) Apply(string) error {
	return nil
}

// GenericHTTPOption defines a general purpose HTTP option
type GenericHTTPOption struct {
	Key   string
	Value string
}

// Header formats the option as an http header
func (o *GenericHTTPOption) Header() (key string, value string) {
	return o.Key, o.Value
}

// String formats the option into human-readable form
func (o *GenericHTTPOption) String() string {
	return fmt.Sprintf("GenericHTTPOption(%q,%q)", o.Key, o.Value)
}

// Mandatory returns whether the option must be parsed or can be ignored
func (o *GenericHTTPOption) Mandatory() bool {
	return false
}

// NullOption defines an Option which does nothing
type NullOption struct {
}

// Header formats the option as an http header
func (o NullOption) Header() (key string, value string) {
	return "", ""
}

// Apply doesn't do anything as this is a null option
func (o NullOption) Apply(string) error {
	return nil
}

// String formats the option into human-readable form
func (o NullOption) String() string {
	return "NullOption()"
}

// Mandatory returns whether the option must be parsed or can be ignored
func (o NullOption) Mandatory() bool {
	return false
}

// FixRangeOption adjusts RangeOptions that request a fetch from the end into an
// absolute fetch using the size passed in
func FixRangeOption(options []OpenOption, size int64) {
	if size < 0 {
		// Can't do anything for unknown length objects
		return
	} else if size == 0 {
		// if size 0 then remove RangeOptions
		for i := range options {
			if _, ok := options[i].(*RangeOption); ok {
				options[i] = NullOption{}
			}
		}
		return
	}
	for i, option := range options {
		switch x := option.(type) {
		case *RangeOption:
			// If start is < 0 then fetch from the end
			if x.Start < 0 {
				x = &RangeOption{Start: size - x.End, End: -1}
				options[i] = x
			}
			// If end is too big or undefined, fetch to the end
			if x.End > size || x.End < 0 {
				x = &RangeOption{Start: x.Start, End: size - 1}
				options[i] = x
			}
		case *SeekOption:
			options[i] = &RangeOption{Start: x.Offset, End: size - 1}
		}
	}
}
