// Package oauthutil provides OAuth utilities.
package oauthutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
)

// The key used for encryption
var cryptKey = []byte{
	0x9c, 0x93, 0x5b, 0x48, 0x73, 0x0a, 0x55, 0x4d,
	0x6b, 0xfd, 0x7c, 0x63, 0xc8, 0x86, 0xa9, 0x2b,
	0xd3, 0x90, 0x19, 0x8e, 0xb8, 0x12, 0x8a, 0xfb,
	0x7d, 0x59, 0x45, 0x36, 0x69, 0x25, 0x77, 0xc1,
}

// crypt encrypts or decrypts data using the provided key and iv
func crypt(ciphertext, plaintext, iv []byte) error {
	aesBlock, err := aes.NewCipher(cryptKey)
	if err != nil {
		return err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	stream.XORKeyStream(ciphertext, plaintext)
	return nil
}

// Obscure obscures a string using AES-CTR
func Obscure(plaintext string) (string, error) {
	plaintextBytes := []byte(plaintext)
	ciphertext := make([]byte, aes.BlockSize+len(plaintextBytes))
	iv := ciphertext[:aes.BlockSize]

	// Generate random IV
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate random IV: %w", err)
	}

	// Encrypt the data
	if err := crypt(ciphertext[aes.BlockSize:], plaintextBytes, iv); err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// Reveal decrypts an obscured string
func Reveal(obscured string) (string, error) {
	ciphertext, err := base64.RawURLEncoding.DecodeString(obscured)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("input too short")
	}

	iv := ciphertext[:aes.BlockSize]
	plaintext := make([]byte, len(ciphertext)-aes.BlockSize)

	if err := crypt(plaintext, ciphertext[aes.BlockSize:], iv); err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// encryptTokenObj encrypts an OAuth2 token object with a password
// This uses the EncryptToken function from crypto.go
func encryptTokenObj(token *oauth2.Token, password string) (string, error) {
	// Marshal the token to JSON
	data, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %w", err)
	}
	// Use the existing EncryptToken function
	return EncryptToken(string(data), password)
}

// DecryptTokenObj decrypts an OAuth2 token string into a token object
// This is different from the DecryptToken in crypto.go
func DecryptTokenObj(encryptedData, password string) (*oauth2.Token, error) {
	// Use existing DecryptToken from crypto.go
	plaintext, err := DecryptToken(encryptedData, password)
	if err != nil {
		return nil, err
	}

	// Unmarshal the token
	var token oauth2.Token
	err = json.Unmarshal([]byte(plaintext), &token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

// SaveEncryptedToken saves an OAuth token to a file with encryption
func SaveEncryptedToken(path string, token *oauth2.Token, password string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// If no password provided, save as plain JSON
	if password == "" {
		data, err := json.Marshal(token)
		if err != nil {
			return fmt.Errorf("failed to marshal token: %w", err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			return fmt.Errorf("failed to write token file: %w", err)
		}
		return nil
	}
	// Encrypt the token
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	encryptedData, err := EncryptToken(string(data), password)
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, []byte(encryptedData), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadEncryptedToken loads an OAuth token from a file with encryption
func LoadEncryptedToken(path string, password string) (*oauth2.Token, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// If no password provided, assume plain JSON
	if password == "" {
		var token oauth2.Token
		if err := json.Unmarshal(data, &token); err != nil {
			return nil, fmt.Errorf("failed to unmarshal token: %w", err)
		}
		return &token, nil
	}
	// Decrypt the token
	plaintext, err := DecryptToken(string(data), password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(plaintext), &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted token: %w", err)
	}

	return &token, nil
}

// IsTokenEncrypted checks if a token file is encrypted
func IsTokenEncrypted(path string) (bool, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	content := strings.TrimSpace(string(data))

	// Check if it has the encrypted token prefix
	if strings.HasPrefix(content, "ENCRYPTED:") {
		return true, nil
	}

	// Try to unmarshal as plain JSON
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err == nil {
		// Successfully unmarshaled as plain JSON
		return false, nil
	}

	// If it's not JSON and doesn't have the prefix, it might be an old format
	// Try to decode as base64 (legacy format)
	_, err = base64.StdEncoding.DecodeString(content)
	if err != nil {
		return false, errors.New("file is neither plain JSON nor encrypted format")
	}

	return true, nil
}

// GenerateRandomPassword generates a secure random password
func GenerateRandomPassword(length int) (string, error) {
	if length < 16 {
		length = 16 // Minimum reasonable password length
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}
