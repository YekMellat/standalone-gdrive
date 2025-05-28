# Token Encryption Testing Plan

This document outlines how to test the token encryption functionality in the standalone Google Drive client.

## Prerequisites

- Go 1.21 or later
- Git
- Access to a Google account for OAuth testing

## Building the Project

```bash
# Clone the repository if not already done
git clone https://github.com/yourusername/standalone-gdrive.git
cd standalone-gdrive

# Build the main binary and token utility
go build -o gdrive ./cmd/gdrive
go build -o token ./cmd/token
```

## Test Suite 1: Unit Tests

Run the automated tests for token encryption:

```bash
# Run all tests in the oauthutil package
go test -v ./lib/oauthutil

# Run specific token encryption tests
go test -v ./lib/oauthutil -run TestTokenEncryption
go test -v ./drive -run TestTokenPersistence
```

## Test Suite 2: Token CLI Tool

Test the token encryption command-line tool:

```bash
# First, authenticate with Google Drive to create a token
./gdrive auth

# Check if the token is encrypted (should be plaintext initially)
./token --check

# Encrypt the token
./token --encrypt
# (enter a password when prompted)

# Check if encryption was successful
./token --check
# (should show "Token is encrypted")

# Try to use the encrypted token
./gdrive ls
# (should prompt for password)

# Decrypt the token
./token --decrypt
# (enter the password)

# Check if decryption was successful
./token --check
# (should show "Token is not encrypted")
```

## Test Suite 3: Environment Variable Support

Test using the environment variable for non-interactive usage:

```bash
# Encrypt the token first
./token --encrypt
# (enter a password, e.g., "test-password")

# Set the environment variable with the password
export GDRIVE_TOKEN_PASSWORD="test-password"

# Use the client, should not prompt for password
./gdrive ls

# Clear the environment variable
unset GDRIVE_TOKEN_PASSWORD

# Use the client again, should prompt for password
./gdrive ls
```

## Test Suite 4: Token Encryption Example

Run the token encryption example to demonstrate the feature:

```bash
# Build and run the token encryption example
go build -o token_example ./examples/token_encryption
./token_example
```

## Test Suite 5: Integration Testing

Test the complete OAuth flow with encrypted tokens:

```bash
# Reset authentication to test full flow
rm ~/.config/standalone-gdrive/gdrive.token

# Build and run the OAuth test tool
go build -o oauth_test ./cmd/oauth_test
./oauth_test --encrypt

# Should go through full OAuth flow and encrypt the resulting token
# Verify the token is encrypted
./token --check
```

## Expected Results

- All unit tests should pass successfully
- Token encryption and decryption should work as expected
- The token CLI tool should correctly identify encrypted tokens
- Environment variable support should allow non-interactive usage
- After encryption, token files should not contain plaintext credentials
- The drive client should work seamlessly with encrypted tokens

## Troubleshooting

If issues occur during testing:

1. Check that the `GDRIVE_TOKEN_PASSWORD` environment variable is set correctly if using non-interactive mode
2. Ensure token files have the correct permissions (0600)
3. For authentication errors, try resetting the authentication with `rm ~/.config/standalone-gdrive/gdrive.token`
4. If a token becomes corrupted, use the backup file created during encryption/decryption operations
