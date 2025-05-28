// Package drive implements a Google Drive client for standalone usage
package drive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/standalone-gdrive/lib/oauthutil"
	"golang.org/x/oauth2"
	"golang.org/x/term"
)

// TokenManager handles loading and saving OAuth tokens
type TokenManager struct {
	tokenPath string
	password  string
}

// NewTokenManager creates a new token manager
func NewTokenManager(configDir string) *TokenManager {
	tokenPath := filepath.Join(configDir, "token.json")
	password := os.Getenv("GDRIVE_TOKEN_PASSWORD")

	return &TokenManager{
		tokenPath: tokenPath,
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
	isEncrypted, err := oauthutil.IsTokenEncrypted(tm.tokenPath)
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
		token, err := oauthutil.LoadEncryptedToken(tm.tokenPath, tm.password)
		if err != nil {
			return nil, fmt.Errorf("failed to load encrypted token: %w", err)
		}
		return token, nil
	}
	// Not encrypted, load normally
	token, err := oauthutil.LoadToken(filepath.Dir(tm.tokenPath), filepath.Base(strings.TrimSuffix(tm.tokenPath, ".token")))
	if err != nil {
		return nil, fmt.Errorf("failed to load token: %w", err)
	}

	return token, nil
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
		return oauthutil.SaveEncryptedToken(tm.tokenPath, token, tm.password)
	}

	return oauthutil.SaveToken(filepath.Dir(tm.tokenPath), filepath.Base(strings.TrimSuffix(tm.tokenPath, ".token")), token)
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
