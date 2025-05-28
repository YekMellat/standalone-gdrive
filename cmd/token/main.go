package main

import (
	"context"
	"fmt"
	"os"

	"github.com/standalone-gdrive/internal/token"
)

func main() {
	// Create a token command
	cmd := token.NewCommand()

	// Execute the command with all arguments
	ctx := context.Background()
	if err := cmd.Execute(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
