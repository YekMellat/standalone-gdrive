# OAuth Authentication Guide

This document explains the OAuth authentication flow used in the standalone Google Drive client.

## Overview

The standalone Google Drive client uses OAuth 2.0 to authenticate with Google's APIs. This allows the client to access Google Drive on behalf of the user without needing the user's password.

## Authentication Methods

The client supports multiple authentication methods:

1. **Interactive Browser Authentication** - The default method that opens a web browser for user consent
2. **Service Account Authentication** - For server applications using a service account key file
3. **Existing Token Authentication** - Uses a previously acquired token

## Interactive Browser Authentication Flow

The standard authentication flow works as follows:

1. **Authorization Request**:
   - Client initiates authentication with the Google OAuth2 server
   - A browser window opens showing Google's consent screen
   - User logs in if necessary and grants permissions
   - Google redirects to a local callback server with an authorization code

2. **Token Acquisition**:
   - Client exchanges the code for access and refresh tokens
   - Tokens are stored in the configured location (default: `~/.config/standalone-gdrive/token.json`)

3. **Token Usage**:
   - Access token is included in API requests as a bearer token
   - When the access token expires, the refresh token is used to obtain a new one

### Code Example

```go
package main

import (
    "context"
    "log"

    "github.com/standalone-gdrive/drive"
)

func main() {
    ctx := context.Background()
    
    // Configuration with default OAuth flow
    config := map[string]string{
        "config_dir": "~/.config/standalone-gdrive",
    }
    
    // This will trigger the OAuth flow if no token exists
    driveFs, err := drive.NewFs(ctx, "gdrive", "/", config)
    if err != nil {
        log.Fatalf("Failed to authenticate: %v", err)
    }
    
    // Now authenticated and ready to use
}
```

## Service Account Authentication

For server applications or automation where user interaction isn't possible, service account authentication can be used:

1. **Create a Service Account**:
   - Go to the Google Cloud Console
   - Create a new project or select an existing one
   - Enable the Google Drive API
   - Create a service account
   - Download the service account key file (.json)

2. **Configure the Client**:

```go
config := map[string]string{
    "service_account_file": "path/to/service-account.json",
}

driveFs, err := drive.NewFs(ctx, "gdrive", "/", config)
```

### Scopes

The client uses the following OAuth scopes by default:

- `https://www.googleapis.com/auth/drive` - Full access to Google Drive

You can request more limited scopes by setting the `scope` option:

```go
config := map[string]string{
    "scope": "drive.readonly", // Read-only access
}
```

Available scope shortcuts:

- `drive` - Full access (default)
- `drive.readonly` - Read-only access
- `drive.file` - Access to files created or opened by the app only
- `drive.appfolder` - Access to the application data folder only
- `drive.metadata.readonly` - Read-only access to file metadata

## Token Storage

By default, tokens are stored in:

- `~/.config/standalone-gdrive/token.json` on Unix-like systems
- `%APPDATA%\standalone-gdrive\token.json` on Windows

You can customize the location with the `config_dir` option:

```go
config := map[string]string{
    "config_dir": "/custom/path/to/config",
}
```

### Token Format

The token file is a JSON file containing:

- `access_token` - Used to authenticate requests
- `refresh_token` - Used to obtain new access tokens
- `token_type` - Almost always "Bearer"
- `expiry` - When the access token expires

Example:

```json
{
  "access_token": "ya29.a0AfB_byC...",
  "token_type": "Bearer",
  "refresh_token": "1//0eF...",
  "expiry": "2023-05-18T13:45:30.123456789Z"
}
```

## Troubleshooting

Common authentication issues:

1. **"Token has been revoked"**:
   - The refresh token has been invalidated
   - Delete the token file and re-authenticate

2. **"Invalid credentials"**:
   - Check that your OAuth client ID and secret are correct
   - Ensure the Google Drive API is enabled in your Google Cloud project

3. **"Access denied"**:
   - The requested scopes were denied by the user
   - Try again and accept all permission requests

4. **"No browser found"**:
   - The client couldn't open a browser for authentication
   - Set `use_browser` to `false` and manually open the displayed URL
