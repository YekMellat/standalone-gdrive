// Package drive implements a Google Drive client for standalone usage
//
// This file contains progress tracking utilities for file operations
package drive

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

// ProgressReader wraps an io.Reader to track progress
type ProgressReader struct {
	reader      io.Reader
	total       int64
	read        int64
	callback    ProgressCallback
	lastUpdate  time.Time
	updateEvery time.Duration
}

// ProgressCallback is a function that gets called with progress updates
type ProgressCallback func(bytesRead int64, total int64, percentage float64, speed float64)

// NewProgressReader returns a reader that tracks progress
func NewProgressReader(r io.Reader, size int64, callback ProgressCallback) *ProgressReader {
	return &ProgressReader{
		reader:      r,
		total:       size,
		callback:    callback,
		updateEvery: 250 * time.Millisecond,
		lastUpdate:  time.Now().Add(-24 * time.Hour), // Ensure first update happens
	}
}

// Read reads data from the underlying reader and updates progress
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if n > 0 {
		newRead := atomic.AddInt64(&pr.read, int64(n))
		now := time.Now()
		if now.Sub(pr.lastUpdate) >= pr.updateEvery {
			percentage := float64(0)
			if pr.total > 0 {
				percentage = float64(newRead) / float64(pr.total) * 100.0
			}

			// Calculate speed in bytes per second
			elapsed := now.Sub(pr.lastUpdate).Seconds()
			speed := float64(0)
			if elapsed > 0 {
				speed = float64(newRead-atomic.LoadInt64(&pr.read)) / elapsed
			}

			pr.callback(newRead, pr.total, percentage, speed)
			pr.lastUpdate = now
		}
	}
	return n, err
}

// DefaultProgressPrinter returns a default progress callback that prints to stdout
func DefaultProgressPrinter(prefix string) ProgressCallback {
	startTime := time.Now()

	return func(bytesRead int64, total int64, percentage float64, instantSpeed float64) {
		now := time.Now()
		elapsed := now.Sub(startTime).Seconds()

		// Calculate overall average speed
		avgSpeed := float64(bytesRead) / elapsed

		// Format speeds
		var avgSpeedStr, curSpeedStr string
		if avgSpeed < 1024 {
			avgSpeedStr = fmt.Sprintf("%.0f B/s", avgSpeed)
		} else if avgSpeed < 1024*1024 {
			avgSpeedStr = fmt.Sprintf("%.2f KB/s", avgSpeed/1024)
		} else {
			avgSpeedStr = fmt.Sprintf("%.2f MB/s", avgSpeed/(1024*1024))
		}

		if instantSpeed < 1024 {
			curSpeedStr = fmt.Sprintf("%.0f B/s", instantSpeed)
		} else if instantSpeed < 1024*1024 {
			curSpeedStr = fmt.Sprintf("%.2f KB/s", instantSpeed/1024)
		} else {
			curSpeedStr = fmt.Sprintf("%.2f MB/s", instantSpeed/(1024*1024))
		}

		// Calculate ETA
		eta := "Unknown"
		if avgSpeed > 0 && total > 0 {
			secondsLeft := float64(total-bytesRead) / avgSpeed
			if secondsLeft < 60 {
				eta = fmt.Sprintf("%.0fs", secondsLeft)
			} else if secondsLeft < 3600 {
				eta = fmt.Sprintf("%.1fm", secondsLeft/60)
			} else {
				eta = fmt.Sprintf("%.1fh", secondsLeft/3600)
			}
		}

		// Format completed/total
		completed := bytesRead
		var completedStr, totalStr string

		if total < 1024 {
			completedStr = fmt.Sprintf("%d B", completed)
			totalStr = fmt.Sprintf("%d B", total)
		} else if total < 1024*1024 {
			completedStr = fmt.Sprintf("%.1f KB", float64(completed)/1024)
			totalStr = fmt.Sprintf("%.1f KB", float64(total)/1024)
		} else if total < 1024*1024*1024 {
			completedStr = fmt.Sprintf("%.2f MB", float64(completed)/(1024*1024))
			totalStr = fmt.Sprintf("%.2f MB", float64(total)/(1024*1024))
		} else {
			completedStr = fmt.Sprintf("%.2f GB", float64(completed)/(1024*1024*1024))
			totalStr = fmt.Sprintf("%.2f GB", float64(total)/(1024*1024*1024))
		}
		// Update status line
		fmt.Printf("\r%s: %s/%s %.1f%% [%s, %s current, ETA %s]      ",
			prefix, completedStr, totalStr, percentage, avgSpeedStr, curSpeedStr, eta)
	}
}

// FinishProgress prints a newline to finish the progress display
func FinishProgress() {
	fmt.Println()
}
