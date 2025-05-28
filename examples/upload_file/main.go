package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/standalone-gdrive/drive"
	"github.com/standalone-gdrive/fs"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <local_file> <remote_path>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s myfile.txt /backup/myfile.txt\n", os.Args[0])
		os.Exit(1)
	}

	localFile := os.Args[1]
	remotePath := os.Args[2]

	// Check if local file exists
	fileInfo, err := os.Stat(localFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing local file: %v\n", err)
		os.Exit(1)
	}

	if fileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "Local path is a directory, not a file\n")
		os.Exit(1)
	}

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

	fmt.Printf("Connecting to Google Drive...\n")

	// Get parent directory path
	parentDir := filepath.Dir(remotePath)
	if parentDir == "." {
		parentDir = "/"
	}

	// Initialize Drive filesystem
	driveFs, err := drive.NewFs(ctx, "gdrive", parentDir, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing Google Drive: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Connected to Google Drive\n")
	fmt.Printf("Uploading %s to %s\n", localFile, remotePath)

	// Open the local file
	file, err := os.Open(localFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening local file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close() // Create ObjectInfoImpl for the file
	info := &fs.ObjectInfoImpl{
		RemoteName:  filepath.Base(remotePath),
		FileSize:    fileInfo.Size(),
		FileModTime: fileInfo.ModTime(),
	}

	// Start timer
	startTime := time.Now()

	// Upload the file
	obj, err := driveFs.Put(ctx, file, info, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error uploading file: %v\n", err)
		os.Exit(1)
	}

	// Calculate duration and speed
	duration := time.Since(startTime)
	speed := float64(fileInfo.Size()) / duration.Seconds() / 1024 / 1024 // MB/s

	fmt.Printf("Upload complete: %s (%.2f MB/s)\n", obj.Remote(), speed)
	fmt.Printf("Size: %d bytes\n", obj.Size())
	fmt.Printf("Modified: %s\n", obj.ModTime(ctx).Format(time.RFC3339))

	// If it's a metadata-capable object
	if m, ok := obj.(fs.MimeTyper); ok {
		fmt.Printf("MIME Type: %s\n", m.MimeType(ctx))
	}

	// If it has an ID
	if id, ok := obj.(fs.IDer); ok {
		fmt.Printf("ID: %s\n", id.ID())
	}
}
