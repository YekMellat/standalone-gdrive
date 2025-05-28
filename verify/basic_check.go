package verify

import (
	"context"
	"fmt"
	"log"

	"github.com/standalone-gdrive/drive"
)

// BasicCheck performs a basic verification of the Google Drive client functionality
func BasicCheck() error {
	ctx := context.Background()

	// Initialize the drive filesystem
	f, err := drive.NewFs(ctx, "gdrive", "", nil)
	if err != nil {
		return fmt.Errorf("failed to initialize Google Drive: %v", err)
	}

	// Test listing root directory
	fmt.Println("Testing basic Google Drive connectivity...")
	_, err = f.List(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list root directory: %v", err)
	}

	fmt.Println("Basic connectivity test passed!")
	return nil
}

// RunBasicCheck runs the basic verification and logs results
func RunBasicCheck() {
	if err := BasicCheck(); err != nil {
		log.Fatalf("Basic check failed: %v", err)
	}
	fmt.Println("All basic checks passed successfully!")
}
