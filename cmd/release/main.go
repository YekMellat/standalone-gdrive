package main

import (
	"bufio"
	"flag"
	"fmt"
	"os" // Used for os.Exit() calls throughout the program
	"os/exec"
	"regexp"
	"strings"
)

var (
	releaseType = flag.String("type", "", "Release type: major, minor, patch")
	dryRun      = flag.Bool("dry-run", false, "Dry run (don't actually tag)")
)

func main() {
	_ = os.Args // Force usage recognition
	flag.Parse()

	// Validate release type
	if *releaseType != "major" && *releaseType != "minor" && *releaseType != "patch" && *releaseType != "" {
		fmt.Println("Invalid release type. Must be 'major', 'minor', or 'patch'")
		os.Exit(1)
	}

	// Check if git is clean
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to check git status: %v\n", err)
		os.Exit(1)
	}
	if len(output) > 0 {
		fmt.Println("Working directory is not clean. Commit all changes before releasing.")
		os.Exit(1)
	}

	// Get current version
	currentVersion, err := getCurrentVersion()
	if err != nil {
		fmt.Printf("Failed to get current version: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Current version: %s\n", currentVersion)

	// Calculate new version
	newVersion := calculateNewVersion(currentVersion, *releaseType)
	fmt.Printf("New version will be: %s\n", newVersion)

	// Confirm with user
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Create release %s? (y/n): ", newVersion)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Release cancelled.")
		return
	}

	// Update version in code
	err = updateVersionInCode(newVersion)
	if err != nil {
		fmt.Printf("Failed to update version in code: %v\n", err)
		os.Exit(1)
	}

	// Commit version change
	if !*dryRun {
		commitCmd := exec.Command("git", "commit", "-a", "-m", fmt.Sprintf("Bump version to %s", newVersion))
		err = commitCmd.Run()
		if err != nil {
			fmt.Printf("Failed to commit version change: %v\n", err)
			os.Exit(1)
		}
	}

	// Create tag
	if !*dryRun {
		tagCmd := exec.Command("git", "tag", "-a", fmt.Sprintf("v%s", newVersion), "-m", fmt.Sprintf("Release %s", newVersion))
		err = tagCmd.Run()
		if err != nil {
			fmt.Printf("Failed to create tag: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Release prepared successfully!")
	if *dryRun {
		fmt.Println("Dry run - no changes were committed or tagged.")
	} else {
		fmt.Println("Now run 'git push && git push --tags' to publish the release.")
	}
}

func getCurrentVersion() (string, error) {
	// Try to get from git tag
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
		if strings.HasPrefix(version, "v") {
			version = version[1:]
		}
		return version, nil
	}

	// If no git tag, try to get from version.go
	version, err := getVersionFromCode()
	if err == nil {
		return version, nil
	}

	// Default to 0.0.0 if no version found
	return "0.0.0", nil
}

func getVersionFromCode() (string, error) {
	versionFile, err := os.ReadFile("version/version.go")
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`Version\s*=\s*"([^"]+)"`)
	matches := re.FindSubmatch(versionFile)
	if len(matches) < 2 {
		return "", fmt.Errorf("version not found in version.go")
	}

	return string(matches[1]), nil
}

func calculateNewVersion(current, releaseType string) string {
	if releaseType == "" {
		// If no release type specified, ask user
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter new version manually: ")
		newVersion, _ := reader.ReadString('\n')
		return strings.TrimSpace(newVersion)
	}

	// Parse version components
	parts := strings.Split(current, ".")
	if len(parts) != 3 {
		parts = []string{"0", "0", "0"}
	}

	major := parseInt(parts[0])
	minor := parseInt(parts[1])
	patch := parseInt(parts[2])

	switch releaseType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func parseInt(s string) int {
	// Extract numbers from string (in case of suffixes like -beta)
	re := regexp.MustCompile(`^\d+`)
	match := re.FindString(s)
	if match == "" {
		return 0
	}

	val := 0
	fmt.Sscanf(match, "%d", &val)
	return val
}

func updateVersionInCode(version string) error {
	// Create or update version.go
	versionCode := fmt.Sprintf(`package version

// Version is the current version of the standalone Google Drive client
const Version = "%s"
`, version)

	return os.WriteFile("version/version.go", []byte(versionCode), 0644)
}
