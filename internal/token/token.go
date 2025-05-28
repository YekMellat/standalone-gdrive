// Package token provides the token command.
package token

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	"github.com/standalone-gdrive/lib/oauthutil"
	"golang.org/x/oauth2"
	"golang.org/x/term"
)

// Command is the command structure
type Command struct {
	tokenCommand    *CommandDef
	encryptCommand  *CommandDef
	decryptCommand  *CommandDef
	checkCommand    *CommandDef
	generateCommand *CommandDef
}

// CommandDef defines a command
type CommandDef struct {
	Use   string
	Short string
	Long  string
	Run   func(ctx context.Context, args []string) error
}

// NewCommand creates a new token command
func NewCommand() *Command {
	cmd := &Command{}

	cmd.tokenCommand = &CommandDef{
		Use:   "token",
		Short: "Manage OAuth tokens",
		Long: `This command allows you to manage OAuth tokens used by standalone-gdrive.

You can encrypt tokens for secure storage, decrypt encrypted tokens, check if a token is encrypted,
and generate a password for encrypting tokens.

For Google Drive, use GDRIVE_TOKEN_PASSWORD environment variable to specify the password for
encrypting and decrypting tokens.`,
	}

	cmd.encryptCommand = &CommandDef{
		Use:   "encrypt",
		Short: "Encrypt an OAuth token",
		Long: `This command encrypts an OAuth token using AES-256-GCM encryption.

It reads an unencrypted token file and encrypts it with the provided password.
The password can be provided via the GDRIVE_TOKEN_PASSWORD environment variable
or interactively.

Example:
  token encrypt --file token.json --output encrypted_token.json
  # Or using environment variable:
  # export GDRIVE_TOKEN_PASSWORD=your_secure_password
  # token encrypt --file token.json --output encrypted_token.json`,
		Run: cmd.encryptToken,
	}

	cmd.decryptCommand = &CommandDef{
		Use:   "decrypt",
		Short: "Decrypt an OAuth token",
		Long: `This command decrypts an encrypted OAuth token.

It reads an encrypted token file and decrypts it with the provided password.
The password can be provided via the GDRIVE_TOKEN_PASSWORD environment variable
or interactively.

Example:
  token decrypt --file encrypted_token.json --output decrypted_token.json
  # Or using environment variable:
  # export GDRIVE_TOKEN_PASSWORD=your_secure_password
  # token decrypt --file encrypted_token.json --output decrypted_token.json`,
		Run: cmd.decryptToken,
	}

	cmd.checkCommand = &CommandDef{
		Use:   "check",
		Short: "Check if a token is encrypted",
		Long: `This command checks if a token file is encrypted.

Example:
  token check --file token.json`,
		Run: cmd.checkToken,
	}

	cmd.generateCommand = &CommandDef{
		Use:   "generate",
		Short: "Generate a secure password for token encryption",
		Long: `This command generates a secure random password that can be used for token encryption.

Example:
  token generate
  token generate --length 32`,
		Run: cmd.generatePassword,
	}

	return cmd
}

// GetUsage returns usage information for the command
func (c *Command) GetUsage() string {
	return fmt.Sprintf(`token: A tool for managing OAuth tokens for standalone-gdrive

Usage:
  token [command] [flags]

Available Commands:
  encrypt     %s
  decrypt     %s
  check       %s
  generate    %s

Flags:
  -p, --password     Prompt for password (otherwise uses GDRIVE_TOKEN_PASSWORD)
  -f, --file         Token file path
  -o, --output       Output file path (for encrypt/decrypt command)
  -l, --length       Password length (for generate command, default: 24)
  -h, --help         Display this help
`, c.encryptCommand.Short, c.decryptCommand.Short, c.checkCommand.Short, c.generateCommand.Short)
}

// promptForPassword prompts the user for a password
func promptForPassword() (string, error) {
	fmt.Print("Enter password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add a newline after the password input
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(passwordBytes), nil
}

// getPassword gets the password either from the environment or by prompting
func getPassword(prompt bool) (string, error) {
	if prompt {
		return promptForPassword()
	}

	// Try to get password from environment
	password := os.Getenv("GDRIVE_TOKEN_PASSWORD")
	if password == "" {
		return "", fmt.Errorf("password not set - use -p flag or set GDRIVE_TOKEN_PASSWORD environment variable")
	}
	return password, nil
}

// encryptToken encrypts a token file
func (c *Command) encryptToken(ctx context.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("missing required arguments")
	}

	// Parse flags
	var inputFile, outputFile string
	var promptPassword bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-f", "--file":
			if i+1 < len(args) {
				inputFile = args[i+1]
				i++
			}
		case "-o", "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "-p", "--password":
			promptPassword = true
		}
	}

	if inputFile == "" {
		return fmt.Errorf("input file path not specified")
	}

	if outputFile == "" {
		outputFile = inputFile
	}

	// Get password
	password, err := getPassword(promptPassword)
	if err != nil {
		return err
	}

	// Read token file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}

	tokenData := string(data)
	if oauthutil.IsEncryptedToken(tokenData) {
		return fmt.Errorf("token is already encrypted")
	}

	// Check if the content is valid JSON
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return fmt.Errorf("token file does not contain valid JSON: %w", err)
	}

	// Encrypt token
	encryptedToken, err := oauthutil.EncryptToken(tokenData, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}

	// Write encrypted token
	err = os.WriteFile(outputFile, []byte(encryptedToken), 0600)
	if err != nil {
		return fmt.Errorf("failed to write encrypted token: %w", err)
	}

	fmt.Printf("Token encrypted and saved to: %s\n", outputFile)
	return nil
}

// decryptToken decrypts a token file
func (c *Command) decryptToken(ctx context.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("missing required arguments")
	}

	// Parse flags
	var inputFile, outputFile string
	var promptPassword bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-f", "--file":
			if i+1 < len(args) {
				inputFile = args[i+1]
				i++
			}
		case "-o", "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "-p", "--password":
			promptPassword = true
		}
	}

	if inputFile == "" {
		return fmt.Errorf("input file path not specified")
	}

	if outputFile == "" {
		outputFile = inputFile
	}

	// Get password
	password, err := getPassword(promptPassword)
	if err != nil {
		return err
	}

	// Read token file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}

	tokenData := string(data)
	if !oauthutil.IsEncryptedToken(tokenData) {
		return fmt.Errorf("token is not encrypted")
	}

	// Decrypt token
	decryptedToken, err := oauthutil.DecryptToken(tokenData, password)
	if err != nil {
		if err == oauthutil.ErrWrongPassword {
			return fmt.Errorf("incorrect password")
		}
		return fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Write decrypted token
	err = os.WriteFile(outputFile, []byte(decryptedToken), 0600)
	if err != nil {
		return fmt.Errorf("failed to write decrypted token: %w", err)
	}

	fmt.Printf("Token decrypted and saved to: %s\n", outputFile)
	return nil
}

// checkToken checks if a token is encrypted
func (c *Command) checkToken(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("missing required arguments")
	}

	// Parse flags
	var inputFile string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-f", "--file":
			if i+1 < len(args) {
				inputFile = args[i+1]
				i++
			}
		}
	}

	if inputFile == "" {
		return fmt.Errorf("file path not specified")
	}

	// Read token file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}

	tokenData := string(data)
	if oauthutil.IsEncryptedToken(tokenData) {
		fmt.Println("Token is encrypted")
	} else {
		fmt.Println("Token is not encrypted")
	}

	return nil
}

// generatePassword generates a secure password
func (c *Command) generatePassword(ctx context.Context, args []string) error {
	// Parse flags
	length := 24

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-l", "--length":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &length)
				i++
			}
		}
	}

	password, err := oauthutil.GeneratePassword(length)
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	fmt.Println("Generated password:")
	fmt.Println(password)
	fmt.Println("\nYou can use this with the GDRIVE_TOKEN_PASSWORD environment variable:")
	fmt.Println("export GDRIVE_TOKEN_PASSWORD='" + password + "'")

	return nil
}

// Execute runs the token command
func (c *Command) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Println(c.GetUsage())
		return nil
	}

	command := args[0]
	commandArgs := args

	switch command {
	case "encrypt":
		return c.encryptCommand.Run(ctx, commandArgs)
	case "decrypt":
		return c.decryptCommand.Run(ctx, commandArgs)
	case "check":
		return c.checkCommand.Run(ctx, commandArgs)
	case "generate":
		return c.generateCommand.Run(ctx, commandArgs)
	default:
		fmt.Println(c.GetUsage())
		return fmt.Errorf("unknown command: %s", command)
	}
}
