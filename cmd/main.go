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
//
// For detailed usage information, run:
//
//	flint-vault --help
//	flint-vault <command> --help
package main

import "flint-vault/pkg/commands"

// main is the entry point of the Flint Vault application.
// It delegates to the commands package which handles CLI parsing and execution.
func main() {
	commands.Run()
}
