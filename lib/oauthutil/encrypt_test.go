package oauthutil

import (
	"testing"
	
	"golang.org/x/oauth2"
)

func TestTokenEncryptionDecryption(t *testing.T) {
	// Create a test token
	token := &oauth2.Token{
		AccessToken: "test-access-token",
		TokenType: "Bearer",
		RefreshToken: "test-refresh-token",
	}
	
	// Test password
	password := "test-password"
	
	// Encrypt the token
	encrypted, err := EncryptToken(token, password)
	if err != nil {
		t.Fatalf("Failed to encrypt token: %v", err)
	}
	
	// Decrypt the token
	decrypted, err := DecryptToken(encrypted, password)
	if err != nil {
		t.Fatalf("Failed to decrypt token: %v", err)
	}
	
	// Verify the decrypted token matches the original
	if token.AccessToken != decrypted.AccessToken {
		t.Errorf("AccessToken mismatch: expected %s, got %s", token.AccessToken, decrypted.AccessToken)
	}
	if token.TokenType != decrypted.TokenType {
		t.Errorf("TokenType mismatch: expected %s, got %s", token.TokenType, decrypted.TokenType)
	}
	if token.RefreshToken != decrypted.RefreshToken {
		t.Errorf("RefreshToken mismatch: expected %s, got %s", token.RefreshToken, decrypted.RefreshToken)
	}
}

func TestIsTokenEncrypted(t *testing.T) {
	// Test with an encrypted token
	token := &oauth2.Token{
		AccessToken: "test-access-token",
		TokenType: "Bearer",
		RefreshToken: "test-refresh-token",
	}
	
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "token-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	
	// Test 1: Save a plain token
	data, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("Failed to marshal token: %v", err)
	}
	if err := os.WriteFile(tmpfile.Name(), data, 0600); err != nil {
		t.Fatalf("Failed to write token file: %v", err)
	}
	
	// Check that it's detected as not encrypted
	encrypted, err := IsTokenEncrypted(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to check if token is encrypted: %v", err)
	}
	if encrypted {
		t.Error("Plain JSON token incorrectly detected as encrypted")
	}
	
	// Test 2: Save an encrypted token
	encryptedData, err := EncryptToken(token, "test-password")
	if err != nil {
		t.Fatalf("Failed to encrypt token: %v", err)
	}
	if err := os.WriteFile(tmpfile.Name(), []byte(encryptedData), 0600); err != nil {
		t.Fatalf("Failed to write encrypted token file: %v", err)
	}
	
	// Check that it's detected as encrypted
	encrypted, err = IsTokenEncrypted(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to check if token is encrypted: %v", err)
	}
	if !encrypted {
		t.Error("Encrypted token incorrectly detected as not encrypted")
	}
}

func TestTokenManager(t *testing.T) {
	// Create a test directory
	tempDir, err := os.MkdirTemp("", "token-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a test token
	token := &oauth2.Token{
		AccessToken: "test-access-token",
		TokenType: "Bearer", 
		RefreshToken: "test-refresh-token",
	}
	
	// Create a token manager with a password
	manager := NewTokenManager(tempDir, "test")
	manager.password = "test-password"
	
	// Save the token
	err = manager.SaveToken(token)
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}
	
	// Load the token
	loaded, err := manager.LoadToken(nil)
	if err != nil {
		t.Fatalf("Failed to load token: %v", err)
	}
	
	// Verify the loaded token matches the original
	if token.AccessToken != loaded.AccessToken {
		t.Errorf("AccessToken mismatch: expected %s, got %s", token.AccessToken, loaded.AccessToken)
	}
	if token.TokenType != loaded.TokenType {
		t.Errorf("TokenType mismatch: expected %s, got %s", token.TokenType, loaded.TokenType)
	}
	if token.RefreshToken != loaded.RefreshToken {
		t.Errorf("RefreshToken mismatch: expected %s, got %s", token.RefreshToken, loaded.RefreshToken)
	}
}
