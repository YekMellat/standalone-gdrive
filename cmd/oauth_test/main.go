package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/standalone-gdrive/drive"
	"github.com/standalone-gdrive/fs"
)

func main() {
	// Create context
	ctx := context.Background()

	// Parse command-line arguments
	var encrypt, reset, decrypt bool
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--encrypt":
			encrypt = true
		case "--reset":
			reset = true
		case "--decrypt":
			decrypt = true
		}
	}

	// Configuration options
	configDir := os.Getenv("HOME") + "/.config/standalone-gdrive-test"
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		configDir = dir
	}

	// Ensure config directory exists
	os.MkdirAll(configDir, 0700)

	// Delete existing token file to force re-authentication
	tokenFile := filepath.Join(configDir, "token.json")
	if _, err := os.Stat(tokenFile); err == nil {
		if reset {
			fmt.Println("Removing existing token file to force re-authentication")
			os.Remove(tokenFile)
		} else {
			fmt.Println("Using existing token file. Use --reset to force re-authentication.")

			// Handle token encryption/decryption if requested
			if encrypt && os.Getenv("GDRIVE_TOKEN_PASSWORD") == "" {
				fmt.Println("To encrypt the token, please set the GDRIVE_TOKEN_PASSWORD environment variable")
				os.Exit(1)
			}

			if decrypt && os.Getenv("GDRIVE_TOKEN_PASSWORD") == "" {
				fmt.Println("To decrypt the token, please set the GDRIVE_TOKEN_PASSWORD environment variable")
				os.Exit(1)
			}
		}
	} // Configuration for the drive client (not used directly)
	// options := &drive.Options{
	//     ConfigDir: configDir,
	// }

	if encrypt || os.Getenv("GDRIVE_TOKEN_PASSWORD") != "" {
		fmt.Println("Token will be encrypted using the password from GDRIVE_TOKEN_PASSWORD")
	}

	fmt.Println("Starting OAuth authentication flow...")
	fmt.Println("You should be redirected to a browser for authentication.")
	// Initialize Drive filesystem - this will trigger OAuth flow if token doesn't exist
	startTime := time.Now()
	config := map[string]string{
		"config_dir": configDir,
		"type":       "drive",
	}
	driveFs, err := drive.NewFs(ctx, "gdrive", "/", config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing Google Drive: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Authentication successful! Took %v\n", time.Since(startTime))

	// Test that we can actually access the drive
	fmt.Println("Testing access by listing root directory...")
	entries, err := driveFs.List(ctx, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Access confirmed! Found %d entries in root directory.\n", len(entries))

	// Create a test file to verify write access
	fmt.Println("Testing write access by creating a test file...")

	// Create test content
	testContent := []byte(fmt.Sprintf("Test file created at %s", time.Now().Format(time.RFC3339)))
	// Upload test file
	info := &fs.ObjectInfoImpl{
		RemoteName:  "oauth-test-file.txt",
		FileSize:    int64(len(testContent)),
		FileModTime: time.Now(),
	}

	obj, err := driveFs.Put(ctx, bytes.NewReader(testContent), info, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating test file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Write access confirmed! Created file: %s (ID: %s)\n", obj.Remote(), obj.(fs.IDer).ID())

	// Clean up test file
	fmt.Println("Cleaning up test file...")
	err = obj.Remove(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to remove test file: %v\n", err)
	} else {
		fmt.Println("Test file removed successfully")
	}

	fmt.Println("OAuth flow test completed successfully!")
}
