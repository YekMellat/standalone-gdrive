// Package hash provides hash utilities for the filesystem
package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"
)

// ErrUnsupported indicates requested hash type is not supported
var ErrUnsupported = errors.New("hash type not supported")

// Type indicates a standard hashing algorithm
type Type int

// Constants for supported hash types
const (
	None Type = 0
	MD5  Type = 1 << iota
	SHA1
	SHA256
	TypeUnset Type = 0xFFFFFFFF
)

type hashDefinition struct {
	width    int
	name     string
	newFunc  func() hash.Hash
	hashType Type
}

var (
	type2hash = map[Type]*hashDefinition{
		MD5:    {width: 32, name: "md5", newFunc: md5.New, hashType: MD5},
		SHA1:   {width: 40, name: "sha1", newFunc: sha1.New, hashType: SHA1},
		SHA256: {width: 64, name: "sha256", newFunc: sha256.New, hashType: SHA256},
	}
	name2hash = map[string]*hashDefinition{
		"md5":    type2hash[MD5],
		"sha1":   type2hash[SHA1],
		"sha256": type2hash[SHA256],
	}
	supported = []Type{MD5, SHA1, SHA256}
)

// Set contains all the hashes currently in use
// The value of the map is the result of the hash for a file
type Set map[Type]string

// NewHashSet returns a hash.Set of all the requested hash types
func NewHashSet(types []Type) Set {
	if len(types) == 0 {
		return nil
	}
	set := make(Set, len(types))
	for _, hashType := range types {
		set[hashType] = ""
	}
	return set
}

// String returns a string representation of the hash type
func (t Type) String() string {
	var types []string
	for _, hashType := range supported {
		if t&hashType != 0 {
			types = append(types, type2hash[hashType].name)
		}
	}
	if len(types) == 0 {
		return "none"
	}
	return strings.Join(types, ",")
}

// FromString parses a string representation of hash types
func FromString(s string) (Type, error) {
	if s == "" || s == "none" {
		return None, nil
	}
	parts := strings.Split(s, ",")
	hashType := None
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if definition, ok := name2hash[strings.ToLower(part)]; ok {
			hashType |= definition.hashType
		} else {
			return Type(0), fmt.Errorf("unknown hash type %q", part)
		}
	}
	return hashType, nil
}

// Width returns the width in characters for the hash
func (t Type) Width() int {
	if t == None {
		return 0
	}
	definition, ok := type2hash[t]
	if !ok {
		return 0
	}
	return definition.width
}

// New returns a new hash of the given type
func (t Type) New() hash.Hash {
	if t == None {
		return nil
	}
	definition, ok := type2hash[t]
	if !ok {
		return nil
	}
	return definition.newFunc()
}

// Sum returns the hash of the data passed in according to the hash type
func (t Type) Sum(data []byte) (string, error) {
	if t == None {
		return "", errors.New("can't compute hash for None type")
	}
	definition, ok := type2hash[t]
	if !ok {
		return "", fmt.Errorf("unknown hash type %v", t)
	}
	h := definition.newFunc()
	_, err := h.Write(data)
	if err != nil {
		return "", fmt.Errorf("failed to hash data: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// Stream returns the hash of the io.Reader passed in
func (t Type) Stream(r io.Reader) (string, error) {
	if t == None {
		return "", errors.New("can't compute hash for None type")
	}
	definition, ok := type2hash[t]
	if !ok {
		return "", fmt.Errorf("unknown hash type %v", t)
	}
	h := definition.newFunc()
	_, err := io.Copy(h, r)
	if err != nil {
		return "", fmt.Errorf("failed to hash data: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
