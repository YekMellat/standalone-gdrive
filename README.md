# Standalone Google Drive Client

This is a standalone implementation of Google Drive service extracted from the rclone project. It provides a complete, lightweight Go library for interacting with Google Drive, with all the features from the original rclone implementation.

[![Build and Test](https://github.com/standalone-gdrive/standalone-gdrive/actions/workflows/go.yml/badge.svg)](https://github.com/standalone-gdrive/standalone-gdrive/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/standalone-gdrive/standalone-gdrive)](https://goreportcard.com/report/github.com/standalone-gdrive/standalone-gdrive)
[![GoDoc](https://godoc.org/github.com/standalone-gdrive/standalone-gdrive?status.svg)](https://godoc.org/github.com/standalone-gdrive/standalone-gdrive)

## Features

- Full Google Drive API support for files and folders
- Team Drive / Shared Drive support with dedicated API methods
- Google Docs import/export with configurable formats
- Chunked uploads for large files with resumable upload support
- Metadata support
- Command-line interface with progress tracking for file operations
- File integrity verification with checksum validation
- Persistent OAuth token storage and automatic refresh
- Robust error handling and retry logic with contextualized errors
- Support for accessing shared files with resource keys
- Multiple authentication methods:
  - OAuth2 authentication
  - Service account support
  - API Key authentication
  - Token encryption for secure storage
- Rate limiting with automatic backoff
- Configurable logging system with multiple log levels
- Granular access control with different Google Drive scopes
- Timeout support for long-running operations
- Version tracking and user-agent customization
- Comprehensive documentation and examples

## Installation

### Prerequisites

- Go 1.19 or later

### Building from source

```bash
git clone https://github.com/yourusername/standalone-gdrive.git
cd standalone-gdrive
go build -o gdrive ./cmd/gdrive
```

## Usage

### As a library

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/standalone-gdrive/drive"
    "github.com/standalone-gdrive/fs"
)

func main() {
    ctx := context.Background()
      // Create a new Google Drive client with configuration
    config := map[string]string{
        "config_dir": os.Getenv("HOME") + "/.config/standalone-gdrive",
    }
    
    driveFs, err := drive.NewFs(ctx, "gdrive", "/", config)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create drive fs: %v\n", err)
        os.Exit(1)
    }
    
    // List files in the root directory
    entries, err := driveFs.List(ctx, "")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to list directory: %v\n", err)
        os.Exit(1)
    }
    
    for _, entry := range entries {
        fmt.Println(entry.Remote())
    }
}
```

### As a command-line tool

```
$ gdrive -command=ls -path=/Documents
D           0 2023-05-10 15:04:05 Project1
D           0 2023-05-12 10:30:22 Project2
F     1048576 2023-05-15 09:45:18 Report.docx
F       12345 2023-05-16 14:22:33 Notes.txt
```

## Available Commands

- `ls` - List files and directories
- `mkdir` - Create directory
- `rm` - Remove file or directory
- `cat` - Display file content
- `cp` - Copy file
- `mv` - Move file
- `info` - Show file information
- `about` - Show account information
- `upload` - Upload file with progress tracking
- `download` - Download file with progress tracking and checksum verification
- `version` - Show version information
- `encrypt` - Encrypt a local file with password protection
- `decrypt` - Decrypt a previously encrypted file
- `mv` - Move file
- `info` - Show information about a file
- `about` - Show information about the Google Drive account
- `upload` - Upload a file to Google Drive
- `download` - Download a file from Google Drive

## Authentication

The client supports multiple authentication methods:

1. **OAuth2 Authentication** - Standard OAuth flow requiring user interaction:
   ```go
   // The client will automatically open a browser for authentication
   options := map[string]string{
       "client_id": "your-client-id",
       "client_secret": "your-client-secret",
   }
   ```

2. **Service Account Authentication** - For server environments without user interaction:
   ```go
   options := map[string]string{
       "service_account_file": "/path/to/service-account.json",
   }
   ```

3. **API Key Authentication** - Limited access using just an API key:
   ```go
   options := map[string]string{
       "api_key": "your-api-key",
   }
   ```

4. **Token Encryption** - Secure storage of OAuth tokens:
   ```go
   options := map[string]string{
       "config_dir": "/path/to/config",
       "encrypt_tokens": "true",
   }
   os.Setenv("GDRIVE_TOKEN_PASSWORD", "your-encryption-key")
   ```

## Configuration

Configuration options can be provided when creating a new client:

```go
options := map[string]string{
    "scope": "drive",                           // API scope
    "root_folder_id": "0ABCdef123456",          // Root folder ID
    "team_drive": "0ABCdef123456",              // Team Drive ID
    "service_account_file": "service-key.json", // Service account credentials
    "log_level": "INFO",                        // Logging level (SILENT, ERROR, WARN, INFO, DEBUG, TRACE)
    "log_output": "/path/to/logfile.log",       // Log file path (empty for stderr)
}

driveFs, err := drive.NewFs(ctx, "gdrive", "/", options)
```

## Logging

The client includes a comprehensive logging system with multiple log levels:

```go
// Get the filesystem instance
fs, ok := driveFs.(*drive.Fs)
if ok {
    // Set log level
    fs.SetLogLevel(drive.LogLevelDebug)
    
    // Use log methods
    fs.LogInfo("Starting file upload operation")
    fs.LogDebug("Processing file: %s", filename)
    
    // Get the logger for custom configuration
    logger := fs.GetLogger()
    logger.SetOutput(customWriter)
}
```

Available log levels:
- `LogLevelSilent` - No logging
- `LogLevelError` - Error messages only
- `LogLevelWarn` - Warnings and errors
- `LogLevelInfo` - Informational messages (default)
- `LogLevelDebug` - Debug information
- `LogLevelTrace` - Very verbose debugging

## Testing

### Running Tests

The project includes both unit tests and integration tests:

```bash
# Run all tests
go test ./...

# Run only unit tests (no API calls)
go test -short ./...

# Run tests with race detection
go test -race ./...
```

### Integration Tests

Integration tests interact with the Google Drive API and require authentication. To run these tests:

1. Set the `TEST_GDRIVE_ACCESS` environment variable to any value
2. Ensure you have valid OAuth credentials in your config directory

```bash
# Run all tests including integration tests
TEST_GDRIVE_ACCESS=1 go test ./...
```

### OAuth Flow Testing

To test the OAuth authentication flow specifically:

```bash
# Build the OAuth test tool
go build -o oauth_test ./cmd/oauth_test

# Run the test (will use existing token if available)
./oauth_test

# Force re-authentication
./oauth_test --reset
```

### Benchmarks

Performance benchmarks are available for critical operations:

```bash
# Run all benchmarks
TEST_GDRIVE_ACCESS=1 go test -bench=. ./...

# Run specific benchmarks
TEST_GDRIVE_ACCESS=1 go test -bench=BenchmarkSmallFileUpload ./drive
```

## Team Drive Support

This client includes enhanced support for Google Shared Drives (formerly Team Drives):

```go
// List all available Team Drives
teamDrives, err := fs.ListTeamDrives(ctx)
if err != nil {
    // handle error
}

for _, teamDrive := range teamDrives {
    fmt.Printf("Team Drive: %s (ID: %s)\n", teamDrive.Name, teamDrive.ID)
}

// Get a specific Team Drive
teamDrive, err := fs.GetTeamDrive(ctx, "team-drive-id")
if err != nil {
    // handle error
}
fmt.Printf("Team Drive: %s, Description: %s\n", teamDrive.Name, teamDrive.Description)

// List files in a Team Drive folder
files, err := fs.ListFilesInTeamDrive(ctx, "folder-id", "team-drive-id", false)
if err != nil {
    // handle error
}

// Upload a file to a Team Drive
file, err := fs.UploadFileToTeamDrive(ctx, "local-file.txt", "parent-folder-id", "team-drive-id", "remote-filename.txt")
if err != nil {
    // handle error
}
```

## Enhanced Security Features

### Token Encryption

This client includes enhanced security features for protecting OAuth tokens:

```bash
# Encrypt your OAuth token
gdrive token --encrypt

# Decrypt an encrypted token
gdrive token --decrypt

# Check if a token is encrypted
gdrive token --check

# Generate a secure random password
gdrive token --generate --length=32
```

To use encrypted tokens in your applications:

```go
// Set password via environment variable (recommended for automated scripts)
os.Setenv("GDRIVE_TOKEN_PASSWORD", "your-secure-password")

// The client will automatically detect encrypted tokens and decrypt them
driveFs, err := drive.NewFs(ctx, "gdrive", "/", config)

// For interactive applications, if GDRIVE_TOKEN_PASSWORD is not set,
// the client will prompt for a password when needed
```

Benefits of token encryption:

- Protect OAuth credentials from unauthorized access
- Secure token storage on shared systems or backup media
- Support for scripted operation via environment variables
- Automatic detection of encrypted tokens
- Interactive password prompts for user-facing applications

## Credit

This project is based on the Google Drive implementation from [rclone](https://github.com/rclone/rclone), simplified and extracted as a standalone package.
