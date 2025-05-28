// Package oauthutil provides OAuth utilities.
package oauthutil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/oauth2"
	"golang.org/x/term"
)

// TokenManager handles loading and saving OAuth tokens with encryption support
type TokenManager struct {
	tokenPath string
	name      string
	password  string
}

// NewTokenManager creates a new token manager
func NewTokenManager(configDir, name string) *TokenManager {
	tokenPath := TokenPath(configDir, name)
	password := os.Getenv("GDRIVE_TOKEN_PASSWORD")

	return &TokenManager{
		tokenPath: tokenPath,
		name:      name,
		password:  password,
	}
}

// LoadToken loads a token from the configured path
func (tm *TokenManager) LoadToken(ctx context.Context) (*oauth2.Token, error) {
	// Check if token file exists
	if _, err := os.Stat(tm.tokenPath); os.IsNotExist(err) {
		return nil, nil
	}

	// Check if token is encrypted
	isEncrypted, err := IsTokenEncrypted(tm.tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if token is encrypted: %w", err)
	}

	// If token is encrypted but no password provided, prompt for password
	if isEncrypted && tm.password == "" {
		if !isInteractive() {
			return nil, fmt.Errorf("token is encrypted but no password provided (set GDRIVE_TOKEN_PASSWORD environment variable)")
		}

		fmt.Fprintln(os.Stderr, "Token is encrypted. Please enter password to decrypt.")
		password, err := promptPassword()
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}
		tm.password = password
	}

	// Load token
	if isEncrypted {
		token, err := LoadEncryptedToken(tm.tokenPath, tm.password)
		if err != nil {
			return nil, fmt.Errorf("failed to load encrypted token: %w", err)
		}
		return token, nil
	}

	// Not encrypted, load normally
	data, err := os.ReadFile(tm.tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token oauth2.Token
	err = json.Unmarshal(data, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &token, nil
}

// SaveToken saves a token to the configured path
func (tm *TokenManager) SaveToken(token *oauth2.Token) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(tm.tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if we should encrypt
	if tm.password != "" {
		return SaveEncryptedToken(tm.tokenPath, token, tm.password)
	}

	// Marshal token to JSON
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Write token to file
	err = os.WriteFile(tm.tokenPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// isInteractive returns true if running in an interactive terminal
func isInteractive() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// promptPassword prompts for a password with no echo
func promptPassword() (string, error) {
	fmt.Print("Enter password: ")

	// Import required for terminal password input
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}

	return string(password), nil
}

// persistentTokenSourceWithManager is a token source that uses TokenManager to save tokens
type persistentTokenSourceWithManager struct {
	wrapped oauth2.TokenSource
	manager *TokenManager
}

// Token returns a token from the wrapped source and saves it using TokenManager
func (s *persistentTokenSourceWithManager) Token() (*oauth2.Token, error) {
	token, err := s.wrapped.Token()
	if err != nil {
		return nil, err
	}

	// Save the token using TokenManager
	err = s.manager.SaveToken(token)
	if err != nil {
		fmt.Printf("Warning: failed to save token: %v\n", err)
	}

	return token, nil
}

// NewPersistentTokenSourceWithManager creates a token source that saves tokens using TokenManager
func NewPersistentTokenSourceWithManager(manager *TokenManager, wrapped oauth2.TokenSource) oauth2.TokenSource {
	return &persistentTokenSourceWithManager{
		wrapped: wrapped,
		manager: manager,
	}
}
