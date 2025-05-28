package drive

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/standalone-gdrive/fs"
)

// BenchmarkLargeFileUpload benchmarks uploading large files (10MB)
// This tests chunked upload functionality
func BenchmarkLargeFileUpload(b *testing.B) {
	// Skip by default as this is time-consuming
	if testing.Short() {
		b.Skip("Skipping large file upload benchmark in short mode")
	}

	driveFs := createBenchFs(b)
	ctx := context.Background()

	// Create test data - 10MB (will trigger chunked upload)
	dataSize := 10 * 1024 * 1024 // 10MB
	data := make([]byte, dataSize)
	_, err := rand.Read(data)
	if err != nil {
		b.Fatalf("Failed to generate test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		info := &fs.ObjectInfoImpl{
			RemoteName:  "bench-large-" + b.Name(),
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
