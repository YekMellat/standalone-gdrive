// Package oauthutil provides OAuth utilities.
package oauthutil

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestTokenSource(t *testing.T) {
	// Create a mock token source
	token := &oauth2.Token{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}
	staticTokenSource := oauth2.StaticTokenSource(token)
	
	ts := &TokenSource{
		tokenSource: staticTokenSource,
		name:        "test",
	}
	
	// Test the Token method
	retrievedToken, err := ts.Token()
	if err != nil {
		t.Errorf("TokenSource.Token() error = %v", err)
	}
	
	if retrievedToken.AccessToken != token.AccessToken {
		t.Errorf("TokenSource.Token() AccessToken = %v, want %v", retrievedToken.AccessToken, token.AccessToken)
	}
}

func TestTokenPersistence(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "oauth_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create a test token
	token := &oauth2.Token{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}
	
	// Test saving token
	tokenPath := filepath.Join(tmpDir, "token.json")
	err = SaveToken(tokenPath, token)
	if err != nil {
		t.Errorf("SaveToken() error = %v", err)
	}
	
	// Test loading token
	loadedToken, err := LoadToken(tokenPath)
	if err != nil {
		t.Errorf("LoadToken() error = %v", err)
	}
	
	if loadedToken.AccessToken != token.AccessToken {
		t.Errorf("Loaded token AccessToken = %v, want %v", loadedToken.AccessToken, token.AccessToken)
	}
	
	if loadedToken.RefreshToken != token.RefreshToken {
		t.Errorf("Loaded token RefreshToken = %v, want %v", loadedToken.RefreshToken, token.RefreshToken)
	}
}

// Mock HTTP handler for OAuth callback testing
type mockHTTPHandler struct {
	code  string
	state string
	err   error
}

func (h *mockHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/auth" {
		// Simulate the OAuth provider redirect
		http.Redirect(w, r, "/callback?code="+h.code+"&state="+h.state, http.StatusFound)
		return
	}
	
	// Handle callback
	if r.URL.Path == "/callback" {
		w.Write([]byte("Authentication successful"))
	}
}

// This test would require a more sophisticated setup in a real environment
func TestOAuthFlow(t *testing.T) {
	// This is a simplified test - in a real environment you'd need to mock
	// the OAuth provider's behavior or use a real provider with test credentials
	t.Skip("Skipping OAuth flow test - requires manual testing with real OAuth provider")
}
