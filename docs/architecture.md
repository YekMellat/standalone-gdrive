# Standalone Google Drive Client Architecture

This document outlines the architecture and design decisions of the standalone Google Drive client.

## Overview

The standalone Google Drive client is structured as a Go library with a layered architecture:

1. **Interface layer** (`fs` package) - Defines core interfaces and types
2. **Implementation layer** (`drive` package) - Implements Google Drive API functionality
3. **Utility layer** (`lib` package) - Provides supporting utilities
4. **Command layer** (`cmd` package) - Offers command-line interface

## Core Components

### FS Interfaces (`fs` package)

The `fs` package defines the core interfaces that all operations build upon:

- `Fs` - Filesystem interface for directory-level operations (list, mkdir, etc.)
- `Object` - Object interface for file-level operations (read, write, etc.)
- `Directory` - Directory interface for directory navigation
- `IDer` - Interface for objects with unique IDs
- `MimeTyper` - Interface for objects with MIME types

These interfaces allow for a clean separation between the public API and the Google Drive implementation details.

### Google Drive Implementation (`drive` package)

The `drive` package implements the interfaces defined in the `fs` package:

- `Fs` - Implemented by `Drive` struct
- `Object` - Implemented by `Object` struct
- `Directory` - Implemented by `Directory` struct

Key features include:

- Chunked uploads for large files
- Handling of Google Workspace document types
- Team Drive / Shared Drive support
- Resource key handling for shared files
- Path resolution and navigation

### OAuth Authentication (`lib/oauthutil` package)

OAuth2 authentication is handled through the `oauthutil` package which provides:

- Token acquisition through web browser flow
- Persistent token storage
- Automatic token refresh
- Service account support

### Directory Cache (`lib/dircache` package)

For efficient directory operations, the `dircache` package provides:

- In-memory caching of directory listings
- Path resolution optimization
- Cache invalidation strategies

### Rate Limiting (`lib/pacer` package)

The `pacer` package implements rate limiting with:

- Adaptive backoff strategies
- Configurable retry limits
- API quota management

## Data Flow

1. **Authentication**:
   - OAuth2 flow -> Token acquisition -> Token storage
   - Tokens automatically refreshed when needed

2. **File Operations**:
   - Path resolution through directory cache
   - API calls through authenticated client
   - Response parsing and object creation

3. **Error Handling**:
   - Transient errors trigger retries with backoff
   - Permanent errors propagated with context

## Design Decisions

### API Version Usage

The client uses both v2 and v3 of the Google Drive API:

- v3 for general operations (better performance and more modern)
- v2 for specific functions where v3 lacks required features

### Chunked Uploads

For large files, the client automatically uses chunked uploads:

1. Upload is initiated with metadata
2. File is uploaded in configurable chunks
3. Upload can be resumed if interrupted

### Resource Key Handling

Shared files with resource keys are handled transparently:

1. Resource keys stored in persistent cache
2. Automatically added to API requests
3. New keys discovered during operations are cached

## Performance Considerations

1. **Directory Caching**:
   - Minimizes API calls for path resolution
   - Cache invalidation on write operations

2. **Connection Reuse**:
   - HTTP connections pooled for better performance
   - Keep-alive enabled for persistent connections

3. **Retry Strategy**:
   - Adaptive backoff for transient failures
   - Rate limit awareness to avoid quota issues

## Testing Strategy

The testing approach includes:

1. **Unit Tests**:
   - Pure functions tested in isolation
   - Mocked dependencies for complex operations

2. **Integration Tests**:
   - Real API interactions with test accounts
   - Controlled by TEST_GDRIVE_ACCESS environment variable

3. **Benchmarks**:
   - Performance testing for critical operations
   - Comparison metrics for various file sizes

## Future Improvements

Potential areas for enhancement:

1. **Concurrency**:
   - Parallel uploads for multi-file operations
   - Worker pool for better resource management

2. **Metadata**:
   - Enhanced custom property support
   - Better integration with Google Workspace features

3. **Caching**:
   - More sophisticated cache invalidation
   - Optional persistent cache for offline capabilities
