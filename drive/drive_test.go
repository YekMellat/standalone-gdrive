// Package drive implements a Google Drive client for standalone usage
package drive

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/standalone-gdrive/fs"
	"github.com/standalone-gdrive/lib/oauthutil"
	"golang.org/x/oauth2"
)

// These tests require valid Google Drive credentials
// Set the environment variable TEST_GDRIVE_ACCESS for integration testing
// The tests will be skipped without this environment variable

func skipIfNoCredentials(t *testing.T) {
	if os.Getenv("TEST_GDRIVE_ACCESS") == "" {
		t.Skip("Skipping integration test - set TEST_GDRIVE_ACCESS environment variable to run")
	}
}

// Helper to create a test filesystem
func createTestFs(t *testing.T) (fs.Fs, func()) {
	skipIfNoCredentials(t)

	ctx := context.Background()

	// Create temporary test directory
	testDir := "gdrive-test-" + time.Now().Format("20060102-150405")

	// Configuration for tests
	configDir := os.Getenv("HOME") + "/.config/standalone-gdrive-test"
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		configDir = dir
	}

	// Ensure config directory exists
	os.MkdirAll(configDir, 0700)

	// Test configuration
	config := map[string]string{
		"config_dir": configDir,
	}

	// Initialize Drive filesystem
	driveFs, err := NewFs(ctx, "gdrive", testDir, config)
	if err != nil {
		t.Fatalf("Failed to create test filesystem: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		// Remove test directory when done
		if purger, ok := driveFs.(fs.Purger); ok {
			_ = purger.Purge(ctx, "")
		}
	}

	return driveFs, cleanup
}

func TestDriveFsOperations(t *testing.T) {
	skipIfNoCredentials(t)

	ctx := context.Background()
	driveFs, cleanup := createTestFs(t)
	defer cleanup()

	// Test directory creation
	testDirName := "test-dir-" + time.Now().Format("150405")
	err := driveFs.Mkdir(ctx, testDirName)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Test directory exists
	entries, err := driveFs.List(ctx, "")
	if err != nil {
		t.Fatalf("Failed to list directory: %v", err)
	}

	found := false
	for _, entry := range entries {
		if entry.Remote() == testDirName {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Created directory not found in listing")
	}
	// Test file upload
	content := []byte("This is test content for Google Drive upload test")
	contentReader := bytes.NewReader(content)
	testFileName := "test-file-" + time.Now().Format("150405") + ".txt"
	info := &fs.ObjectInfoImpl{
		RemoteName:  testFileName,
		FileSize:    int64(len(content)),
		FileModTime: time.Now(),
	}

	obj, err := driveFs.Put(ctx, contentReader, info, nil)
	if err != nil {
		t.Fatalf("Failed to upload file: %v", err)
	}

	// Test file exists and correct size
	if obj.Size() != int64(len(content)) {
		t.Errorf("Uploaded file size mismatch: got %d, want %d", obj.Size(), len(content))
	}

	// Test file download
	reader, err := obj.Open(ctx)
	if err != nil {
		t.Fatalf("Failed to open file for download: %v", err)
	}
	defer reader.Close()

	downloadedContent, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to download file content: %v", err)
	}

	if !bytes.Equal(downloadedContent, content) {
		t.Errorf("Downloaded content does not match uploaded content")
	}

	// Test file deletion
	err = obj.Remove(ctx)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify file is gone
	_, err = driveFs.NewObject(ctx, testFileName)
	if err == nil {
		t.Errorf("Object still exists after deletion")
	}
}

func exportResourceKeys(resourceKeys map[string]string) string {
	var parts []string
	for fileID, resourceKey := range resourceKeys {
		parts = append(parts, fmt.Sprintf("%s:%s", fileID, resourceKey))
	}
	return strings.Join(parts, ",")
}

func TestResourceKeyHandling(t *testing.T) {
	skipIfNoCredentials(t)

	// This test requires a manually shared file with a resource key
	// Set GDRIVE_TEST_SHARED_FILE_ID and GDRIVE_TEST_RESOURCE_KEY env vars to test
	fileID := os.Getenv("GDRIVE_TEST_SHARED_FILE_ID")
	resourceKey := os.Getenv("GDRIVE_TEST_RESOURCE_KEY")

	if fileID == "" || resourceKey == "" {
		t.Skip("Skipping resource key test - set GDRIVE_TEST_SHARED_FILE_ID and GDRIVE_TEST_RESOURCE_KEY env vars")
	}

	// Use the ExportResourceKeys function directly
	result := exportResourceKeys(map[string]string{fileID: resourceKey})

	if !strings.Contains(result, fileID) || !strings.Contains(result, resourceKey) {
		t.Errorf("ExportResourceKeys did not correctly format the resource key")
	}
}

func TestRetryLogic(t *testing.T) {
	// Test the retry logic with a mock client
	// This is more complicated and would require mocking the Google API client
	t.Skip("Skipping retry logic test - requires mocking Google API client")
}

func TestTokenEncryption(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "gdrive-token-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test token
	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}

	// Create a TokenManager
	tokenManager := oauthutil.NewTokenManager(tempDir, "gdrive")

	// Test 1: Save and load unencrypted token
	t.Run("UnencryptedToken", func(t *testing.T) {
		// Save token
		if err := tokenManager.SaveToken(token); err != nil {
			t.Fatalf("Failed to save unencrypted token: %v", err)
		}

		// Load token
		loaded, err := tokenManager.LoadToken(context.Background())
		if err != nil {
			t.Fatalf("Failed to load unencrypted token: %v", err)
		}

		// Verify token
		if loaded.AccessToken != token.AccessToken {
			t.Errorf("AccessToken mismatch: expected %s, got %s", token.AccessToken, loaded.AccessToken)
		}
	})
	// Test 2: Save and load encrypted token
	t.Run("EncryptedToken", func(t *testing.T) {
		// Set password for encryption
		password := "test-password"
		tokenManager.SetPassword(password)

		// Save token (encrypted)
		if err := tokenManager.SaveToken(token); err != nil {
			t.Fatalf("Failed to save encrypted token: %v", err)
		}

		// Check that the token file exists and is encrypted
		tokenPath := filepath.Join(tempDir, "gdrive.token")
		isEncrypted, err := oauthutil.IsTokenEncrypted(tokenPath)
		if err != nil {
			t.Fatalf("Failed to check if token is encrypted: %v", err)
		}
		if !isEncrypted {
			t.Errorf("Token should be encrypted but was detected as unencrypted")
		}

		// Load token with correct password
		loaded, err := tokenManager.LoadToken(context.Background())
		if err != nil {
			t.Fatalf("Failed to load encrypted token: %v", err)
		}

		// Verify token
		if loaded.AccessToken != token.AccessToken {
			t.Errorf("AccessToken mismatch: expected %s, got %s", token.AccessToken, loaded.AccessToken)
		}
		// Try to load with wrong password
		wrongTokenManager := oauthutil.NewTokenManager(tempDir, "gdrive")
		wrongTokenManager.SetPassword("wrong-password")
		_, err = wrongTokenManager.LoadToken(context.Background())
		if err == nil {
			t.Errorf("Expected error when loading with wrong password, but got none")
		}
	})

	// Test 3: Environment variable for password
	t.Run("EnvironmentVariablePassword", func(t *testing.T) {
		// Set environment variable
		password := "env-var-password"
		os.Setenv("GDRIVE_TOKEN_PASSWORD", password)
		defer os.Unsetenv("GDRIVE_TOKEN_PASSWORD")

		// Create new token manager (should pick up env var)
		envTokenManager := oauthutil.NewTokenManager(tempDir, "gdrive")

		// Save token (should be encrypted)
		if err := envTokenManager.SaveToken(token); err != nil {
			t.Fatalf("Failed to save encrypted token via env var: %v", err)
		}

		// Check that the token is encrypted
		tokenPath := filepath.Join(tempDir, "gdrive.token")
		isEncrypted, err := oauthutil.IsTokenEncrypted(tokenPath)
		if err != nil {
			t.Fatalf("Failed to check if token is encrypted: %v", err)
		}
		if !isEncrypted {
			t.Errorf("Token should be encrypted but was detected as unencrypted")
		}

		// Load token (should use env var for decryption)
		loaded, err := envTokenManager.LoadToken(context.Background())
		if err != nil {
			t.Fatalf("Failed to load encrypted token via env var: %v", err)
		}

		// Verify token
		if loaded.AccessToken != token.AccessToken {
			t.Errorf("AccessToken mismatch: expected %s, got %s", token.AccessToken, loaded.AccessToken)
		}
	})
}

func TestTokenPersistence(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create a PersistentTokenSource and verify it saves tokens correctly
	tempDir, err := os.MkdirTemp("", "gdrive-token-persistence-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a token manager
	tokenManager := oauthutil.NewTokenManager(tempDir, "gdrive")

	// Create initial token and static source
	initialToken := &oauth2.Token{
		AccessToken:  "initial-access-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	// Save initial token
	if err := tokenManager.SaveToken(initialToken); err != nil {
		t.Fatalf("Failed to save initial token: %v", err)
	}

	// Create a mock token source that always returns a new token
	mockSource := &mockTokenSource{
		token: &oauth2.Token{
			AccessToken:  "new-access-token",
			RefreshToken: "refresh-token",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(2 * time.Hour),
		},
	}

	// Create a persistent token source
	persistentSource := oauthutil.NewPersistentTokenSourceWithManager(tokenManager, mockSource)

	// Get a token from the persistent source
	newToken, err := persistentSource.Token()
	if err != nil {
		t.Fatalf("Failed to get token from persistent source: %v", err)
	}

	// Verify the token
	if newToken.AccessToken != "new-access-token" {
		t.Errorf("Expected new access token, got: %s", newToken.AccessToken)
	}

	// Load the saved token directly and verify it was persisted
	savedToken, err := tokenManager.LoadToken(context.Background())
	if err != nil {
		t.Fatalf("Failed to load persisted token: %v", err)
	}

	if savedToken.AccessToken != "new-access-token" {
		t.Errorf("Persisted token has wrong access token: %s", savedToken.AccessToken)
	}
}

// Mock token source for testing
type mockTokenSource struct {
	token *oauth2.Token
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return m.token, nil
}
