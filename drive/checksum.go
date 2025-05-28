// Package drive implements a Google Drive client for standalone usage
//
// This file contains checksum verification utilities
package drive

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
)

var (
	// ErrChecksumMismatch indicates that the calculated checksum doesn't match the expected one
	ErrChecksumMismatch = errors.New("checksum mismatch")

	// ErrNoChecksum indicates that no checksum was available for verification
	ErrNoChecksum = errors.New("no checksum available")
)

// ChecksumType represents the type of checksum
type ChecksumType string

const (
	// ChecksumMD5 is the MD5 checksum type
	ChecksumMD5 ChecksumType = "md5"

	// ChecksumSHA1 is the SHA-1 checksum type
	ChecksumSHA1 ChecksumType = "sha1"

	// ChecksumSHA256 is the SHA-256 checksum type
	ChecksumSHA256 ChecksumType = "sha256"
)

// VerifyFileChecksum verifies that the checksum of the given file matches the expected value
func VerifyFileChecksum(filePath string, checksumType ChecksumType, expectedSum string) error {
	if expectedSum == "" {
		return ErrNoChecksum
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum verification: %w", err)
	}
	defer file.Close()

	var h hash.Hash

	switch checksumType {
	case ChecksumMD5:
		h = md5.New()
	case ChecksumSHA1:
		h = sha1.New()
	case ChecksumSHA256:
		h = sha256.New()
	default:
		return fmt.Errorf("unsupported checksum type: %s", checksumType)
	}

	if _, err := io.Copy(h, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	calculatedSum := hex.EncodeToString(h.Sum(nil))

	if calculatedSum != expectedSum {
		return fmt.Errorf("%w: expected %s but got %s (type: %s)",
			ErrChecksumMismatch, expectedSum, calculatedSum, checksumType)
	}

	return nil
}

// CalculateFileChecksum calculates the checksum of the given file
func CalculateFileChecksum(filePath string, checksumType ChecksumType) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum calculation: %w", err)
	}
	defer file.Close()

	var h hash.Hash

	switch checksumType {
	case ChecksumMD5:
		h = md5.New()
	case ChecksumSHA1:
		h = sha1.New()
	case ChecksumSHA256:
		h = sha256.New()
	default:
		return "", fmt.Errorf("unsupported checksum type: %s", checksumType)
	}

	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
