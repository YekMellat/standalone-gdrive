// Package oauthutil provides OAuth utilities.
package oauthutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// TokenPath returns the path to the token file
func TokenPath(configDir, name string) string {
	return filepath.Join(configDir, name+".token")
}

// LoadToken loads the token from a file
func LoadToken(configDir, name string) (*oauth2.Token, error) {
	tokenPath := TokenPath(configDir, name)
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}
	token := &oauth2.Token{}
	err = json.Unmarshal(data, token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}
	return token, nil
}

// SaveToken saves the token to a file
func SaveToken(configDir, name string, token *oauth2.Token) error {
	// Make sure the directory exists
	err := os.MkdirAll(configDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal the token
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Write it out
	tokenPath := TokenPath(configDir, name)
	err = os.WriteFile(tokenPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to save token file: %w", err)
	}

	return nil
}

// persistentTokenSource is a token source that saves tokens to disk
type persistentTokenSource struct {
	wrapped   oauth2.TokenSource
	configDir string
	name      string
}

// Token returns a token from the wrapped source and saves it
func (s *persistentTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.wrapped.Token()
	if err != nil {
		return nil, err
	}

	// Save the token
	err = SaveToken(s.configDir, s.name, token)
	if err != nil {
		fmt.Printf("Warning: failed to save token: %v\n", err)
	}

	return token, nil
}

// NewPersistentTokenSource creates a token source that saves tokens to disk
func NewPersistentTokenSource(configDir, name string, wrapped oauth2.TokenSource) oauth2.TokenSource {
	return &persistentTokenSource{
		wrapped:   wrapped,
		configDir: configDir,
		name:      name,
	}
}
