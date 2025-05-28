# Token Encryption Guide for standalone-gdrive

This guide explains how to use the token encryption feature in standalone-gdrive to secure your OAuth tokens.

## Overview

OAuth tokens grant access to your Google Drive account and should be kept secure. The token encryption feature allows you to encrypt these tokens using AES-256-GCM encryption, ensuring they cannot be used even if someone gains access to your token files.

## Prerequisites

- standalone-gdrive installed and configured
- Environment variable `GDRIVE_TOKEN_PASSWORD` set (optional, but recommended)

## Commands

The `token` command provides several subcommands:

```
token encrypt     # Encrypt an OAuth token
token decrypt     # Decrypt an OAuth token
token check       # Check if a token is encrypted
token generate    # Generate a secure password for token encryption
```

## Usage Examples

### Generate a Secure Password

```bash
token -generate
```

This will output a random password that you can use to encrypt your tokens. Save this password in a secure location.

### Encrypt a Token

```bash
# Set the password as an environment variable
export GDRIVE_TOKEN_PASSWORD="your-secure-password"

# Encrypt the token
token -encrypt
```

Or manually enter the password when prompted:

```bash
token -encrypt -p
```

### Decrypt a Token

```bash
# Using environment variable
export GDRIVE_TOKEN_PASSWORD="your-secure-password"
token -decrypt

# Or with password prompt
token -decrypt -p
```

### Check Token Status

```bash
token -check
```

## Automated Usage

You can integrate token encryption in your applications:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/standalone-gdrive/lib/oauthutil"
)

func main() {
	// Set password from environment variable
	password := os.Getenv("GDRIVE_TOKEN_PASSWORD")
	if password == "" {
		fmt.Println("Please set GDRIVE_TOKEN_PASSWORD environment variable")
		os.Exit(1)
	}

	// Create token manager
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "standalone-gdrive")
	tokenManager := oauthutil.NewTokenManager(configDir, "gdrive")
	tokenManager.SetPassword(password)

	// Load and use token
	token, err := tokenManager.LoadToken(nil)
	if err != nil {
		fmt.Printf("Error loading token: %v\n", err)
		return
	}

	fmt.Println("Token loaded successfully")
	fmt.Printf("Access Token: %s\n", token.AccessToken[0:10]+"...")
}
```

## Security Best Practices

1. Store your encryption password in a secure password manager
2. Use the environment variable (`GDRIVE_TOKEN_PASSWORD`) instead of hardcoding the password
3. Consider using a credential vault/manager for automated processes
4. Regularly rotate your encryption password

## Troubleshooting

- If you receive an "incorrect password" error, double-check your password
- For authentication errors, try resetting the authentication
- If unsure whether your token is encrypted, use `token -check`
