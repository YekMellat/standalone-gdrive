package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/standalone-gdrive/drive"
	"github.com/standalone-gdrive/fs"
)

var (
	testFs       *drive.Fs
	testFolderID string
)

func TestMain(m *testing.M) {
	// Setup
	ctx := context.Background()
	var err error

	// Initialize the FS with test credentials
	driveFs, err := drive.NewFs(ctx, "test", "root", map[string]string{
		"log_level": "INFO",
	})
	if err != nil {
		fmt.Printf("Failed to initialize drive: %v\n", err)
		os.Exit(1)
	}

	// Convert to drive.Fs type
	var ok bool
	testFs, ok = driveFs.(*drive.Fs)
	if !ok {
		fmt.Println("Failed to convert to drive.Fs type")
		os.Exit(1)
	}
	// Create a test folder with timestamp to avoid conflicts
	folderName := fmt.Sprintf("integration-test-%s", time.Now().Format("20060102150405"))
	err = testFs.Mkdir(ctx, folderName)
	if err != nil {
		fmt.Printf("Failed to create test folder: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created test folder: %s\n", folderName)

	// Extract folder ID
	dirPath := folderName
	testFolderID = dirPath

	// Run tests
	code := m.Run()

	// Cleanup
	fmt.Println("Cleaning up test resources...")
	err = testFs.Rmdir(ctx, dirPath)
	if err != nil {
		fmt.Printf("Warning: Failed to delete test folder: %v\n", err)
	}

	os.Exit(code)
}

func TestFileUploadAndDownload(t *testing.T) {
	ctx := context.Background()

	// Create a temp file for testing
	content := []byte("This is test content for file upload and download")
	tmpFile, err := ioutil.TempFile("", "gdrive-test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(content)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	// Upload file
	destPath := fmt.Sprintf("%s/%s", testFolderID, filepath.Base(tmpFile.Name()))

	f, err := testFs.NewObject(ctx, destPath)
	if err == fs.ErrorObjectNotFound { // File doesn't exist, need to create it
		in, err := os.Open(tmpFile.Name())
		require.NoError(t, err)
		defer in.Close()
		f, err = testFs.Put(ctx, in, &fs.ObjectInfoImpl{
			RemoteName:  destPath,
			FileSize:    int64(len(content)),
			FileModTime: time.Now(),
		})
		require.NoError(t, err)
	} else {
		require.NoError(t, err)
	}

	assert.NotNil(t, f)
	assert.Equal(t, int64(len(content)), f.Size())

	// Download the file
	downloadPath := filepath.Join(os.TempDir(), "gdrive-download-test.txt")
	defer os.Remove(downloadPath)

	reader, err := f.Open(ctx)
	require.NoError(t, err)

	// Read the file contents
	downloadedContent := make([]byte, f.Size())
	_, err = reader.Read(downloadedContent)
	require.NoError(t, err)
	reader.Close()

	// Save to file
	err = ioutil.WriteFile(downloadPath, downloadedContent, 0644)
	require.NoError(t, err)

	// Verify content
	assert.Equal(t, content, downloadedContent)

	// Clean up
	err = f.Remove(ctx)
	require.NoError(t, err)
}

func TestFolderOperations(t *testing.T) {
	ctx := context.Background()
	// Create a subfolder
	subfolderName := fmt.Sprintf("%s/test-subfolder", testFolderID)
	err := testFs.Mkdir(ctx, subfolderName)
	require.NoError(t, err)

	// List folders and verify
	entries, err := testFs.List(ctx, testFolderID)
	require.NoError(t, err)

	found := false
	for _, entry := range entries {
		if dir, isDir := entry.(fs.Directory); isDir {
			if dir.Remote() == subfolderName {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Created subfolder not found in listing")

	// Clean up
	err = testFs.Rmdir(ctx, subfolderName)
	require.NoError(t, err)
}
