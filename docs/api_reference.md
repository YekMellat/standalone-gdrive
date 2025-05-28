# API Reference

This document provides a detailed reference for the standalone Google Drive client API.

## Core Interfaces

### `fs.Fs` Interface

The `Fs` interface represents a remote filing system and provides directory-level operations.

```go
type Fs interface {
    Name() string                                             // Name of the remote
    Root() string                                             // Root path
    String() string                                           // String representation of the remote
    Precision() time.Duration                                 // Precision of timestamps
    Features() *Features                                      // Optional features
    List(ctx context.Context, dir string) ([]DirEntry, error) // List files and directories
    NewObject(ctx context.Context, remote string) (Object, error) // Get a single object
    Mkdir(ctx context.Context, dir string) error              // Make a directory
    Rmdir(ctx context.Context, dir string) error              // Remove a directory
    Put(ctx context.Context, in io.Reader, src *ObjectInfo, options ...OpenOption) (Object, error) // Upload a file
}
```

### `fs.Object` Interface

The `Object` interface represents a file and provides file-level operations.

```go
type Object interface {
    DirEntry                                      // Object is also a DirEntry
    Fs() Fs                                       // Get the parent Fs
    Hash(ctx context.Context, ty Hash) (string, error) // Get file hash
    Open(ctx context.Context, options ...OpenOption) (io.ReadCloser, error) // Open for reading
    Update(ctx context.Context, in io.Reader, src *ObjectInfo, options ...OpenOption) error // Update file contents
    Remove(ctx context.Context) error             // Remove this object
    SetModTime(ctx context.Context, t time.Time) error // Set modification time
}
```

### `fs.Directory` Interface

The `Directory` interface represents a directory.

```go
type Directory interface {
    DirEntry
    Items() int64 // Return number of items in directory
}
```

## Main Functionality

### Creating a New Filesystem

To interact with Google Drive, first create a new filesystem:

```go
driveFs, err := drive.NewFs(ctx, "gdrive", "/path", config)
```

Parameters:
- `ctx` - Context for the request
- `name` - Name of the remote (typically "gdrive")
- `path` - Path within Google Drive to use as root
- `config` - Configuration options

Configuration options:

| Option | Description | Default |
|--------|-------------|---------|
| `client_id` | OAuth client ID | Built-in |
| `client_secret` | OAuth client secret | Built-in |
| `token` | OAuth token in JSON format | `""`|
| `service_account_file` | Path to service account key file | `""` |
| `team_drive` | Team Drive ID | `""` |
| `root_folder_id` | ID of the root folder | `""` |
| `scope` | OAuth scope | `"drive"` |
| `chunk_size` | Upload chunk size | `8 MB` |
| `acknowledge_abuse` | Download known abusive files | `false` |
| `config_dir` | Config directory | Platform-specific |

### Listing Files

```go
entries, err := driveFs.List(ctx, "/path/to/folder")
for _, entry := range entries {
    fmt.Println(entry.Remote())
    
    // Check if it's a file or directory
    if _, ok := entry.(fs.Directory); ok {
        fmt.Println("Directory")
    } else {
        fmt.Println("File")
    }
}
```

### Uploading Files

```go
file, _ := os.Open("local-file.txt")
defer file.Close()

info := &fs.ObjectInfo{
    Remote:  "remote-file.txt",
    Size:    fileInfo.Size(),
    ModTime: time.Now(),
}

obj, err := driveFs.Put(ctx, file, info, nil)
```

### Downloading Files

```go
obj, err := driveFs.NewObject(ctx, "file.txt")
if err != nil {
    return err
}

reader, err := obj.Open(ctx)
if err != nil {
    return err
}
defer reader.Close()

// Read from reader
data, err := io.ReadAll(reader)
```

### Creating Directories

```go
err := driveFs.Mkdir(ctx, "new-folder")
```

### Removing Objects

```go
obj, err := driveFs.NewObject(ctx, "file.txt")
if err != nil {
    return err
}

err = obj.Remove(ctx)
```

### Removing Directories

```go
err := driveFs.Rmdir(ctx, "empty-folder")
```

For non-empty directories, use the `Purge` method if the filesystem implements `fs.Purger`:

```go
if purger, ok := driveFs.(fs.Purger); ok {
    err := purger.Purge(ctx, "non-empty-folder")
}
```

## Advanced Functionality

### Working with Google Workspace Documents

Google Workspace documents (Docs, Sheets, etc.) have special MIME types and can be exported in different formats:

```go
obj, err := driveFs.NewObject(ctx, "document.docx")
if exportable, ok := obj.(drive.Exportable); ok {
    formats, _ := exportable.Formats(ctx)
    fmt.Println("Available formats:", formats)
    
    reader, err := exportable.Export(ctx, "application/pdf")
    // Read PDF content
}
```

### Using Team Drives / Shared Drives

```go
config := map[string]string{
    "team_drive": "0ABCdef123456",
}

driveFs, err := drive.NewFs(ctx, "gdrive", "/", config)
```

### Accessing Shared Files with Resource Keys

The client automatically handles resource keys for shared files:

```go
// The resource key will be automatically applied if needed
obj, err := driveFs.NewObject(ctx, "shared-file.txt")
```

You can also manually set resource keys:

```go
SetResourceKey("fileId", "resourceKey")
```

## Authentication and Token Management

### TokenManager

The `TokenManager` provides secure handling of OAuth2 tokens, including encryption support.

```go
type TokenManager struct {
    // contains filtered or unexported fields
}

// NewTokenManager creates a new token manager
func NewTokenManager(configDir string, name string) *TokenManager

// LoadToken loads a token from the configured path, handling encryption if present
func (tm *TokenManager) LoadToken(ctx context.Context) (*oauth2.Token, error)

// SaveToken saves a token to the configured path, encrypting if password is set
func (tm *TokenManager) SaveToken(token *oauth2.Token) error
```

### Token Encryption Functions

The library provides functions for encrypting and decrypting OAuth2 tokens:

```go
// EncryptToken encrypts an OAuth2 token with a password
func EncryptToken(token *oauth2.Token, password string) (string, error)

// DecryptToken decrypts an encrypted token string
func DecryptToken(encryptedData string, password string) (*oauth2.Token, error)

// SaveEncryptedToken saves an encrypted token to a file
func SaveEncryptedToken(path string, token *oauth2.Token, password string) error

// LoadEncryptedToken loads and decrypts a token from a file
func LoadEncryptedToken(path string, password string) (*oauth2.Token, error)

// IsTokenEncrypted checks if a token file is encrypted
func IsTokenEncrypted(path string) (bool, error)
```

### Using Encrypted Tokens

To use encrypted tokens:

1. **Environment Variable**: Set the `GDRIVE_TOKEN_PASSWORD` environment variable with your encryption password.
   
   ```bash
   export GDRIVE_TOKEN_PASSWORD="your-secure-password"
   ```

2. **Token CLI Tool**: Use the token management tool to encrypt/decrypt tokens:

   ```bash
   token --encrypt   # Encrypts an existing token
   token --decrypt   # Decrypts an encrypted token
   token --check     # Checks if a token is encrypted
   token --generate  # Generates a secure random password
   ```

3. **Non-interactive Usage**: For automated scripts, always provide the password via the environment variable.

4. **Interactive Usage**: If no password is provided via environment variable, the application will prompt for a password when needed.
```
