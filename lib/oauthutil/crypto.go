// Package oauthutil provides OAuth utilities.
package oauthutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

// ErrWrongPassword is returned when the password is incorrect
var ErrWrongPassword = errors.New("incorrect password")

// EncryptedTokenPrefix is the prefix for encrypted tokens
const EncryptedTokenPrefix = "ENCRYPTED:"

// IsEncryptedToken checks if a token is encrypted
func IsEncryptedToken(token string) bool {
	return strings.HasPrefix(token, EncryptedTokenPrefix)
}

// EncryptToken encrypts a token with a password
func EncryptToken(token, password string) (string, error) {
	// Check if the token is already encrypted
	if IsEncryptedToken(token) {
		return "", errors.New("token is already encrypted")
	}

	// Generate a key from the password using SHA-256
	key := sha256.Sum256([]byte(password))

	// Create AES cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the token
	ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)

	// Base64 encode the ciphertext
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	// Return the encrypted token with the prefix
	return EncryptedTokenPrefix + encoded, nil
}

// DecryptToken decrypts a token with a password
func DecryptToken(encryptedToken, password string) (string, error) {
	// Check if the token has the prefix
	if !IsEncryptedToken(encryptedToken) {
		return "", errors.New("token is not encrypted")
	}

	// Remove the prefix
	encryptedToken = strings.TrimPrefix(encryptedToken, EncryptedTokenPrefix)

	// Base64 decode the ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", err
	}

	// Generate a key from the password using SHA-256
	key := sha256.Sum256([]byte(password))

	// Create AES cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Check if ciphertext is long enough
	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce from ciphertext
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	// Decrypt the token
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrWrongPassword
	}

	return string(plaintext), nil
}

// GeneratePassword generates a random password
func GeneratePassword(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("password length must be positive")
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b), nil
}
