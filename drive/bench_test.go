package drive

import (
	"bytes"
	"context"
	"crypto/rand"
	"os"
	"testing"
	"time"

	"github.com/standalone-gdrive/fs"
)

func skipBenchmarkIfNoCredentials(b *testing.B) {
	if os.Getenv("TEST_GDRIVE_ACCESS") == "" {
		b.Skip("Skipping benchmark - set TEST_GDRIVE_ACCESS environment variable to run")
	}
}

// Helper to create a filesystem for benchmarking
func createBenchFs(b *testing.B) fs.Fs {
	skipBenchmarkIfNoCredentials(b)

	ctx := context.Background()

	// Create temporary test directory for benchmarks
	testDir := "gdrive-bench-" + b.Name()

	// Configuration for tests
	configDir := os.Getenv("HOME") + "/.config/standalone-gdrive-test"
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		configDir = dir
	}

	// Ensure config directory exists
	os.MkdirAll(configDir, 0700)

	// Test configuration
	config := map[string]string{
		"config_dir": configDir,
	}

	// Initialize Drive filesystem
	driveFs, err := NewFs(ctx, "gdrive", testDir, config)
	if err != nil {
		b.Fatalf("Failed to create test filesystem: %v", err)
	}

	return driveFs
}

// BenchmarkSmallFileUpload benchmarks uploading small files (10KB)
func BenchmarkSmallFileUpload(b *testing.B) {
	driveFs := createBenchFs(b)
	ctx := context.Background()

	// Create test data - 10KB
	data := make([]byte, 10*1024)
	_, err := rand.Read(data)
	if err != nil {
		b.Fatalf("Failed to generate test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		info := &fs.ObjectInfoImpl{
			RemoteName:  "bench-small-" + b.Name(),
			FileSize:    int64(len(data)),
			FileModTime: time.Now(),
		}
		b.StartTimer()

		obj, err := driveFs.Put(ctx, bytes.NewReader(data), info, nil)
		if err != nil {
			b.Fatalf("Upload failed: %v", err)
		}

		b.StopTimer()
		_ = obj.Remove(ctx)
	}
}

// BenchmarkMediumFileUpload benchmarks uploading medium files (1MB)
func BenchmarkMediumFileUpload(b *testing.B) {
	driveFs := createBenchFs(b)
	ctx := context.Background()

	// Create test data - 1MB
	data := make([]byte, 1024*1024)
	_, err := rand.Read(data)
	if err != nil {
		b.Fatalf("Failed to generate test data: %v", err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		info := &fs.ObjectInfoImpl{
			RemoteName:  "bench-medium-" + b.Name(),
			FileSize:    int64(len(data)),
			FileModTime: time.Now(),
		}
		b.StartTimer()

		obj, err := driveFs.Put(ctx, bytes.NewReader(data), info, nil)
		if err != nil {
			b.Fatalf("Upload failed: %v", err)
		}

		b.StopTimer()
		_ = obj.Remove(ctx)
	}
}

// BenchmarkListDirectory benchmarks listing a directory
func BenchmarkListDirectory(b *testing.B) {
	driveFs := createBenchFs(b)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := driveFs.List(ctx, "")
		if err != nil {
			b.Fatalf("List failed: %v", err)
		}
	}
}

// BenchmarkDirLookup benchmarks looking up directories
func BenchmarkDirLookup(b *testing.B) {
	driveFs := createBenchFs(b)
	ctx := context.Background()

	// Create test directory structure
	testDirName := "bench-dir-lookup"
	err := driveFs.Mkdir(ctx, testDirName)
	if err != nil {
		b.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() {
		if purger, ok := driveFs.(fs.Purger); ok {
			_ = purger.Purge(ctx, testDirName)
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := driveFs.List(ctx, testDirName)
		if err != nil {
			b.Fatalf("List failed: %v", err)
		}
	}
}
