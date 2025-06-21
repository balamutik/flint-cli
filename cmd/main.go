// Flint Vault - Secure Encrypted File Storage
//
// Flint Vault is a command-line application that provides military-grade
// encrypted file storage using AES-256-GCM encryption with PBKDF2 key derivation.
//
// This application allows users to:
//   - Create encrypted vaults protected by passwords
//   - Add files and directories to vaults
//   - List vault contents with compression ratios
//   - Extract files from vaults (all or specific) with integrity verification
//   - Remove files from vaults
//   - Get vault information without password
//
// Security features:
//   - AES-256-GCM authenticated encryption
//   - PBKDF2 key derivation with 100,000 iterations
//   - SHA-256 integrity verification for all files
//   - Cryptographically secure random salt and nonce generation
//   - Memory safety with sensitive data clearing
//   - Secure password input (hidden from terminal)
//
// Performance features:
//   - Optimized architecture for large files
//   - Streaming I/O operations for memory efficiency
//   - Gzip compression to reduce vault size
//   - Metadata-only operations for instant listing
//   - Parallel processing for multiple files
//
// Usage:
//
//	flint-vault <command> [options]
//
// Available commands:
//
//	create  - Create new encrypted vault
//	add     - Add file or directory to vault
//	list    - Show vault contents
//	extract - Extract all files from vault
//	get     - Extract specific file or directory
//	remove  - Remove file or directory from vault
//	info    - Show vault file information
//	version - Show version information
//
// For detailed usage information, run:
//
//	flint-vault --help
//	flint-vault <command> --help
package main

import (
	"fmt"
	"os"
	"runtime"

	"flint-vault/pkg/commands"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"     // Version number
	GitCommit = "unknown" // Git commit hash
	BuildTime = "unknown" // Build timestamp
)

// main is the entry point of the Flint Vault application.
// It handles version display and delegates to the commands package for CLI parsing.
func main() {
	// Check if version is requested
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-v") {
		showVersion()
		return
	}

	// Delegate to commands package
	commands.Run()
}

// showVersion displays detailed version and build information
func showVersion() {
	fmt.Printf("Flint Vault %s\n", Version)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("\nSecure encrypted file storage with military-grade AES-256-GCM encryption.\n")
	fmt.Printf("For usage information, run: flint-vault --help\n")
}
