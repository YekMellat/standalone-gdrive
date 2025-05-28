package main

import (
	"context"
	"fmt"
	"os"

	"github.com/standalone-gdrive/drive"
	"github.com/standalone-gdrive/fs"
)

func main() {
	// Create context
	ctx := context.Background()

	// Configuration options
	configDir := os.Getenv("HOME") + "/.config/standalone-gdrive"
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		configDir = dir
	}

	// Ensure config directory exists
	os.MkdirAll(configDir, 0700)

	// Configuration for the drive client
	config := map[string]string{
		"config_dir": configDir,
	}

	// Path in Google Drive to access
	path := "/"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	fmt.Printf("Connecting to Google Drive...\n")

	// Initialize Drive filesystem
	driveFs, err := drive.NewFs(ctx, "gdrive", path, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing Google Drive: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Connected to Google Drive. Listing files in '%s'...\n", path)

	// List files/folders
	entries, err := driveFs.List(ctx, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing directory: %v\n", err)
		os.Exit(1)
	}

	// Print header
	fmt.Printf("%-5s %12s %-19s %s\n", "Type", "Size", "Modified", "Name")
	fmt.Printf("%-5s %12s %-19s %s\n", "----", "----", "--------", "----")

	// Print entries
	for _, entry := range entries {
		// Format the size and date
		size := entry.Size()
		date := entry.ModTime(ctx).Format("2006-01-02 15:04:05")

		// Determine type
		entryType := "FILE"
		if _, ok := entry.(fs.Directory); ok {
			entryType = "DIR"
		}

		// Print entry info
		fmt.Printf("%-5s %12d %19s %s\n", entryType, size, date, entry.Remote())
	}
}
