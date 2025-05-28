package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/standalone-gdrive/lib/oauthutil"
	"golang.org/x/oauth2"
)

func main() {
	// Get home directory for config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}
	configDir := filepath.Join(homeDir, ".config", "standalone-gdrive")

	// Example 1: Basic token management with no encryption
	fmt.Println("=== Example 1: Basic token management ===")
	exampleBasicTokenManagement(configDir)

	// Example 2: Encrypted token management
	fmt.Println("\n=== Example 2: Encrypted token management ===")
	exampleEncryptedTokenManagement(configDir)

	// Example 3: Using the drive client with encrypted tokens
	fmt.Println("\n=== Example 3: Using drive client with encrypted tokens ===")
	exampleDriveWithEncryptedTokens()
}

func exampleBasicTokenManagement(configDir string) {
	// Create a token manager
	manager := oauthutil.NewTokenManager(configDir, "example")

	// Create a sample token (in real usage, this would come from OAuth flow)
	token := &oauth2.Token{
		AccessToken:  "example-access-token",
		TokenType:    "Bearer",
		RefreshToken: "example-refresh-token",
	}

	// Save the token (unencrypted)
	err := manager.SaveToken(token)
	if err != nil {
		log.Fatalf("Failed to save token: %v", err)
	}
	fmt.Println("Token saved successfully (unencrypted)")

	// Load the token
	loadedToken, err := manager.LoadToken(context.Background())
	if err != nil {
		log.Fatalf("Failed to load token: %v", err)
	}
	fmt.Printf("Token loaded successfully: %s\n", loadedToken.AccessToken)
}

func exampleEncryptedTokenManagement(configDir string) {
	// Create a token manager with password
	manager := oauthutil.NewTokenManager(configDir, "encrypted-example")

	// Set password (in real usage, you would use environment variable or prompt)
	password := "example-password"
	fmt.Println("Using password:", password)

	// You can set the password for the token manager
	// In production, you would typically use the GDRIVE_TOKEN_PASSWORD env var
	os.Setenv("GDRIVE_TOKEN_PASSWORD", password)

	// Create a sample token
	token := &oauth2.Token{
		AccessToken:  "example-encrypted-access-token",
		TokenType:    "Bearer",
		RefreshToken: "example-encrypted-refresh-token",
	}

	// Save the token (encrypted)
	manager = oauthutil.NewTokenManager(configDir, "encrypted-example") // Recreate to pick up env var
	err := manager.SaveToken(token)
	if err != nil {
		log.Fatalf("Failed to save encrypted token: %v", err)
	}
	fmt.Println("Token saved successfully (encrypted)")

	// Load the encrypted token
	loadedToken, err := manager.LoadToken(context.Background())
	if err != nil {
		log.Fatalf("Failed to load encrypted token: %v", err)
	}
	fmt.Printf("Encrypted token loaded successfully: %s\n", loadedToken.AccessToken)
}

func exampleDriveWithEncryptedTokens() {
	// In a real application, you would:
	// 1. Set GDRIVE_TOKEN_PASSWORD environment variable, or
	// 2. Let the system prompt for password when needed
	// Example: Create drive options with config directory
	// opts := &drive.Options{
	//     ConfigDir: filepath.Join(os.Getenv("HOME"), ".config", "standalone-gdrive"),
	//     // Other options...
	// }

	fmt.Println("To use drive client with encrypted tokens:")
	fmt.Println("1. Set GDRIVE_TOKEN_PASSWORD environment variable")
	fmt.Println("2. Run your application normally")
	fmt.Println("3. If no password is provided and terminal is interactive,")
	fmt.Println("   the system will prompt for password")
}
