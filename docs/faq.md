# Frequently Asked Questions

## General Questions

### What is this project for?
This standalone Google Drive client provides a simple, focused integration with Google Drive for Go applications without having to depend on the full rclone project.

### How is this different from the official Google Drive Go client?
This client adds several features on top of the official Google Drive API client, including:
- Simplified authentication handling
- Retry logic for better reliability
- Progress reporting for uploads/downloads
- Format conversion for Google Docs
- Higher-level abstractions for common operations

## Authentication Issues

### I'm getting "oauth token empty" errors
This means the client hasn't been authorized yet. Run the authorization method first:
```go
err := client.Authorize(context.Background())
if err != nil {
    // handle error
}
```

### How can I use this in a server environment?
For server environments, we recommend using Service Account authentication which doesn't require user interaction. See the README for details on setting up a service account.

### Why is my token expiring frequently?
Google OAuth tokens typically expire after 1 hour. The client will automatically refresh the token if a refresh token is available. If you're still having issues, make sure you're requesting the proper scopes during authorization.

## File Operations

### Why are large file uploads failing?
Large uploads might fail due to network issues or timeouts. The client includes retry logic, but you might need to adjust the retry parameters:
```go
client.SetRetryParameters(5, 10*time.Second)
```

### How do I download Google Docs as PDFs?
When downloading Google Docs, you can specify the export format:
```go
err := client.DownloadFile(ctx, fileID, localPath, drive.WithExportFormat("application/pdf"))
```

### I'm getting rate limit errors
Google Drive has API quotas. Consider implementing your own rate limiting on top of the client:
```go
time.Sleep(1 * time.Second) // Simple delay between operations
```

## Performance

### How can I improve list performance for large folders?
Use pagination when listing large folders:
```go
client.SetPageSize(100) // Default is 1000
```

### How can I optimize uploads?
For large files, consider using chunked uploads:
```go
client.UploadFile(ctx, localPath, folderID, "", drive.WithChunkSize(8*1024*1024))
```

## Error Handling

### How should I handle temporary errors?
The client handles retries for temporary errors automatically. For additional resilience, you can implement your own retry logic for operations:
```go
for attempts := 0; attempts < 3; attempts++ {
    err := client.OperationXYZ()
    if err == nil {
        break
    }
    time.Sleep(time.Duration(attempts) * time.Second)
}
```

## Team Drives

### How do I work with Team Drives (Shared Drives)?
Use the Team Drive specific methods:
```go
// List all Team Drives you have access to
teamDrives, err := client.ListTeamDrives(ctx)

// List files in a Team Drive
files, err := client.ListFilesInTeamDrive(ctx, "folder-id", "team-drive-id", false)

// Upload to a Team Drive
file, err := client.UploadFileToTeamDrive(ctx, "/local/path", "folder-id", "team-drive-id", "filename")
```

### Why can't I see all files in a Team Drive?
Make sure you're using the Team Drive specific methods which include the necessary parameters (like `SupportsAllDrives(true)`) to access Team Drive content.
