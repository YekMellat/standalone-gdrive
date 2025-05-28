package main

import (
	"bufio"
	"fmt"
	"os" // Used for os.Exit() calls
	"os/exec"
	"strings"
)

func main() {
	_ = os.Args // Force usage recognition
	fmt.Println("Checking for dependency updates...")

	// Run go list to get current dependencies
	cmd := exec.Command("go", "list", "-m", "all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error listing modules: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Current dependencies:")
	fmt.Println(string(output))

	// Ask for confirmation before updating
	fmt.Print("Do you want to update all dependencies? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Update cancelled.")
		return
	}

	// Update dependencies
	fmt.Println("Updating dependencies...")
	updateCmd := exec.Command("go", "get", "-u", "./...")
	updateOutput, err := updateCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error updating dependencies: %v\n", err)
		fmt.Println(string(updateOutput))
		os.Exit(1)
	}

	fmt.Println(string(updateOutput))

	// Tidy up go.mod
	fmt.Println("Tidying go.mod...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyOutput, err := tidyCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error tidying go.mod: %v\n", err)
		fmt.Println(string(tidyOutput))
		os.Exit(1)
	}

	if len(tidyOutput) > 0 {
		fmt.Println(string(tidyOutput))
	}

	// List updated dependencies
	cmd = exec.Command("go", "list", "-m", "all")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error listing updated modules: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Updated dependencies:")
	fmt.Println(string(output))

	fmt.Println("Dependencies updated successfully!")
}
