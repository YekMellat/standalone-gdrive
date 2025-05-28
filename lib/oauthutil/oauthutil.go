// Package oauthutil provides OAuth utilities.
package oauthutil

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/oauth2"
)

// TokenSource stores the OAuth2 token source
type TokenSource struct {
	tokenSource oauth2.TokenSource
	name        string
}

// Token returns a token or an error
func (ts *TokenSource) Token() (*oauth2.Token, error) {
	return ts.tokenSource.Token()
}

// Config contains the data for the oauth config
type Config struct {
	OAuth2Config *oauth2.Config
	Scopes       []string
	AuthURL      string
	TokenURL     string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Options contains the options for OAuthClient
type Options struct {
	OAuth2Config *oauth2.Config
}

// ConfigOutData holds the config response
type ConfigOutData struct {
	State string
}

// ConfigOut returns the Response for configuring the oauth
func ConfigOut(nextState string, options *Options) (*ConfigOutData, error) {
	return &ConfigOutData{
		State: nextState,
	}, nil
}

// RedirectURL is redirect to local webserver when active
const RedirectURL = "http://127.0.0.1:53682/"

// TitleBarRedirectURL is the OAuth2 redirect URL to use when the authorization
// code should be returned in the title bar of the browser, with the page text
// prompting the user to copy the code and paste it in the application.
const TitleBarRedirectURL = "urn:ietf:wg:oauth:2.0:oob"

var (
	// tokenMu protects token
	tokenMu sync.Mutex

	// tokenCache is the cache of tokens
	tokenCache = map[string]*oauth2.Token{}
)

// PutToken stores the token in the cache
func PutToken(name string, token *oauth2.Token) {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	tokenCache[name] = token
}

// GetToken retrieves the token from the cache
func GetToken(name string) *oauth2.Token {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	return tokenCache[name]
}

// Context returns a context with an HTTP Client which
// authenticates with OAuth for the Config passed in.
//
// If optClient is nil, a new http.Client will be created with default timeouts.
func Context(ctx context.Context, optClient *http.Client) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, optClient)
}

// NewClient gets a token from the config file and configures
// a Client with it
func NewClient(ctx context.Context, name string, m map[string]string, config *Config) (*http.Client, *TokenSource, error) {
	return NewClientWithBaseClient(ctx, name, m, config, nil)
}

// NewClientWithBaseClient gets a token from the config file and configures
// a Client with it
func NewClientWithBaseClient(ctx context.Context, name string, m map[string]string, config *Config, baseClient *http.Client) (*http.Client, *TokenSource, error) {
	// Get config directory from map if available
	configDir := "~/.config/standalone-gdrive"
	if m != nil {
		if dir, ok := m["config_dir"]; ok && dir != "" {
			configDir = dir
		}
	}

	// Expand user home directory if needed
	if strings.HasPrefix(configDir, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			configDir = filepath.Join(home, configDir[2:])
		}
	}

	// Create a TokenManager to handle encrypted tokens
	tokenManager := NewTokenManager(configDir, name)

	// Try to load token using the TokenManager
	token, err := tokenManager.LoadToken(ctx)
	if err == nil && token != nil {
		// Token loaded successfully
		config.OAuth2Config.RedirectURL = RedirectURL
		tokenSource := config.OAuth2Config.TokenSource(ctx, token)
		persistentSource := NewPersistentTokenSourceWithManager(tokenManager, tokenSource)
		ts := &TokenSource{
			tokenSource: persistentSource,
			name:        name,
		}
		return oauth2.NewClient(ctx, persistentSource), ts, nil
	}

	// Try to load token from cache as fallback
	token = GetToken(name)
	if token != nil {
		config.OAuth2Config.RedirectURL = RedirectURL
		tokenSource := config.OAuth2Config.TokenSource(ctx, token)
		persistentSource := NewPersistentTokenSourceWithManager(tokenManager, tokenSource)
		ts := &TokenSource{
			tokenSource: persistentSource,
			name:        name,
		}
		return oauth2.NewClient(ctx, persistentSource), ts, nil
	}

	// Simulate a manual authorization flow
	fmt.Printf("No token found. Please authorize this app by visiting:\n")
	config.OAuth2Config.RedirectURL = TitleBarRedirectURL
	authURL := config.OAuth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("%s\n", authURL)
	fmt.Printf("Enter the authorization code: ")
	var code string
	fmt.Scanln(&code)

	token, err = config.OAuth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// Store the token for next time
	PutToken(name, token)

	// Save token using the TokenManager
	err = tokenManager.SaveToken(token)
	if err != nil {
		fmt.Printf("Warning: failed to save token: %v\n", err)
	}

	// Complete the token exchange
	tokenSource := config.OAuth2Config.TokenSource(ctx, token)
	persistentSource := NewPersistentTokenSourceWithManager(tokenManager, tokenSource)
	ts := &TokenSource{
		tokenSource: persistentSource,
		name:        name,
	}
	return oauth2.NewClient(ctx, persistentSource), ts, nil
}
