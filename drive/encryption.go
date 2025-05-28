// Package drive implements a Google Drive client for standalone usage
//
// This file contains cryptographic utilities for file encryption/decryption
package drive

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Errors for encryption operations
var (
	ErrInvalidKey          = errors.New("invalid encryption key")
	ErrEncryptionFailed    = errors.New("encryption failed")
	ErrDecryptionFailed    = errors.New("decryption failed")
	ErrInvalidCiphertext   = errors.New("invalid ciphertext")
)

// EncryptFile encrypts the source file and writes it to the destination
func EncryptFile(sourcePath, destPath, password string) error {
	// Open the source file
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("cannot open source file: %w", err)
	}
	defer source.Close()

	// Create the destination file
	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("cannot create destination file: %w", err)
	}
	defer dest.Close()

	// Get source file info
	sourceInfo, err := source.Stat()
	if err != nil {
		return fmt.Errorf("cannot get source file info: %w", err)
	}

	// Derive key from password
	key := deriveKey(password)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// Generate random IV (initialization vector)
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// Write IV to the beginning of the output file
	if _, err := dest.Write(iv); err != nil {
		return fmt.Errorf("failed to write IV: %w", err)
	}

	// Create stream cipher
	stream := cipher.NewCTR(block, iv)

	// Create an encrypt writer
	encryptWriter := &cipher.StreamWriter{S: stream, W: dest}

	// Copy the source file to the destination while encrypting
	useProgress := isTerminalFunc() && sourceInfo.Size() > 0

	var reader io.Reader = source
	
	if useProgress {
		// Create a progress tracking wrapper around the reader
		progressReader := NewProgressReader(source, sourceInfo.Size(), 
			DefaultProgressPrinter(fmt.Sprintf("Encrypting %s", sourcePath)))
		reader = progressReader
	}

	if _, err = io.Copy(encryptWriter, reader); err != nil {
		return fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	if useProgress {
		FinishProgress() // Print newline after progress bar
	}

	return nil
}

// DecryptFile decrypts the source file and writes it to the destination
func DecryptFile(sourcePath, destPath, password string) error {
	// Open the source file
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("cannot open encrypted file: %w", err)
	}
	defer source.Close()

	// Get source file info
	sourceInfo, err := source.Stat()
	if err != nil {
		return fmt.Errorf("cannot get source file info: %w", err)
	}

	// Create the destination file
	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("cannot create destination file: %w", err)
	}
	defer dest.Close()

	// Read the IV from the first 16 bytes
	iv := make([]byte, aes.BlockSize)
	if _, err := source.Read(iv); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCiphertext, err)
	}

	// Derive key from password
	key := deriveKey(password)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	// Create stream cipher
	stream := cipher.NewCTR(block, iv)

	// Create a decrypt reader
	decryptReader := &cipher.StreamReader{S: stream, R: source}

	// Use progress tracking for decryption
	useProgress := isTerminalFunc() && sourceInfo.Size() > aes.BlockSize
	
	var reader io.Reader = decryptReader
	
	if useProgress {
		// Create a progress tracking wrapper around the reader
		progressReader := NewProgressReader(decryptReader, sourceInfo.Size()-aes.BlockSize, 
			DefaultProgressPrinter(fmt.Sprintf("Decrypting %s", sourcePath)))
		reader = progressReader
	}

	// Copy the decrypted data to the destination
	if _, err = io.Copy(dest, reader); err != nil {
		return fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	if useProgress {
		FinishProgress() // Print newline after progress bar
	}

	return nil
}

// deriveKey creates a 32-byte key from a password
func deriveKey(password string) []byte {
	// Simple key derivation with SHA-256
	// In a production system, use a proper KDF like Argon2id or PBKDF2
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

// isTerminalFunc is a placeholder for terminal detection function
// In a real implementation, this would detect if stdout is a terminal
func isTerminalFunc() bool {
	// For simplicity, always return true
	// In a real implementation, this would check if stdout is a terminal
	return true
}

// IsEncrypted attempts to determine if a file is encrypted
func IsEncrypted(filename string) bool {
	// Check if the file has an encrypted extension
	return strings.HasSuffix(strings.ToLower(filename), ".enc") ||
		strings.HasSuffix(strings.ToLower(filename), ".encrypted")
}

// GenerateEncryptionKey generates a random encryption key and returns it as a hex string
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}
