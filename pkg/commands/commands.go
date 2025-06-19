// Package commands implements the command-line interface for Flint Vault.
// This package provides all CLI commands and their implementations using the urfave/cli framework.
//
// Available commands:
//   - create: Create new encrypted vault
//   - add: Add files or directories to vault
//   - list: Show vault contents
//   - extract: Extract all files from vault
//   - get: Extract specific files or directories
//   - remove: Remove files or directories from vault
//
// All commands support secure password input and provide comprehensive error handling.
package commands

import (
	"context"
	"fmt"
	"log"
	"os"

	"flint-vault/pkg/lib/vault"

	"github.com/urfave/cli/v3"
)

// Run initializes and executes the Flint Vault CLI application.
// This function sets up all available commands with their flags and handlers,
// then processes the command line arguments and executes the appropriate command.
//
// The function does not return - it either successfully executes a command
// or terminates the program with an error via log.Fatal.
//
// Command structure:
//   - Each command has its own set of flags for configuration
//   - Password input is secured by default (hidden from terminal)
//   - All commands provide comprehensive help text
//   - Error messages are user-friendly and descriptive
func Run() {
	app := &cli.Command{
		Name:  "flint-vault",
		Usage: "Secure storage with AES-256 encryption",
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create new encrypted vault",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "Path to vault file",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "Encryption password (NOT RECOMMENDED, better to enter interactively for security)",
						Required: false,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					file := cmd.String("file")
					password := cmd.String("password")

					// If password not specified, request it securely
					if password == "" {
						var err error
						password, err = vault.ReadPasswordSecurely("Enter password for new vault: ")
						if err != nil {
							return err
						}
					}

					fmt.Printf("Creating encrypted vault: %s\n", file)

					if err := vault.CreateVault(file, password); err != nil {
						return fmt.Errorf("vault creation error: %w", err)
					}

					fmt.Println("✅ Vault successfully created!")
					fmt.Println("🔐 Using AES-256-GCM encryption")
					fmt.Println("🧂 Applied cryptographically secure salt")
					fmt.Println("🔑 Key derived using PBKDF2 (100,000 iterations)")

					return nil
				},
			},
			{
				Name:  "add",
				Usage: "Add file or directory to vault",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "vault",
						Aliases:  []string{"v"},
						Usage:    "Path to vault file",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "Vault password (NOT RECOMMENDED, better to enter interactively for security)",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "source",
						Aliases:  []string{"s"},
						Usage:    "Path to file or directory to add",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					vaultPath := cmd.String("vault")
					password := cmd.String("password")
					sourcePath := cmd.String("source")

					if password == "" {
						var err error
						password, err = vault.ReadPasswordSecurely("Enter vault password: ")
						if err != nil {
							return err
						}
					}

					// Check that source exists
					info, err := os.Stat(sourcePath)
					if err != nil {
						return fmt.Errorf("file or directory not found: %s", sourcePath)
					}

					fmt.Printf("Adding %s to vault...\n", sourcePath)

					if info.IsDir() {
						err = vault.AddDirectoryToVault(vaultPath, password, sourcePath)
					} else {
						err = vault.AddFileToVault(vaultPath, password, sourcePath)
					}

					if err != nil {
						return fmt.Errorf("add error: %w", err)
					}

					if info.IsDir() {
						fmt.Printf("✅ Directory '%s' successfully added to vault!\n", sourcePath)
					} else {
						fmt.Printf("✅ File '%s' successfully added to vault!\n", sourcePath)
					}

					return nil
				},
			},
			{
				Name:  "list",
				Usage: "Show vault contents",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "vault",
						Aliases:  []string{"v"},
						Usage:    "Path to vault file",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "Vault password (NOT RECOMMENDED, better to enter interactively for security)",
						Required: false,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					vaultPath := cmd.String("vault")
					password := cmd.String("password")

					if password == "" {
						var err error
						password, err = vault.ReadPasswordSecurely("Enter vault password: ")
						if err != nil {
							return err
						}
					}

					data, err := vault.ListVault(vaultPath, password)
					if err != nil {
						return fmt.Errorf("vault read error: %w", err)
					}

					fmt.Printf("📦 Vault: %s\n", vaultPath)
					fmt.Printf("📅 Created: %s\n", data.CreatedAt.Format("2006-01-02 15:04:05"))
					fmt.Printf("💬 Comment: %s\n", data.Comment)
					fmt.Printf("📁 Contents (%d items):\n\n", len(data.Entries))

					if len(data.Entries) == 0 {
						fmt.Println("  Vault is empty")
						return nil
					}

					// Display list of files and directories
					for _, entry := range data.Entries {
						icon := "📄"
						if entry.IsDir {
							icon = "📁"
						}

						size := formatSize(entry.Size)
						fmt.Printf("  %s %s  %s  %s\n",
							icon,
							entry.Path,
							size,
							entry.ModTime.Format("2006-01-02 15:04"))
					}

					return nil
				},
			},
			{
				Name:  "extract",
				Usage: "Extract all files from vault",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "vault",
						Aliases:  []string{"v"},
						Usage:    "Path to vault file",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "Vault password (NOT RECOMMENDED, better to enter interactively for security)",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "Directory to extract files",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					vaultPath := cmd.String("vault")
					password := cmd.String("password")
					outputDir := cmd.String("output")

					if password == "" {
						var err error
						password, err = vault.ReadPasswordSecurely("Enter vault password: ")
						if err != nil {
							return err
						}
					}

					fmt.Printf("Extracting all files to directory: %s\n", outputDir)

					if err := vault.ExtractVault(vaultPath, password, outputDir); err != nil {
						return fmt.Errorf("extraction error: %w", err)
					}

					fmt.Printf("✅ All files successfully extracted to '%s'!\n", outputDir)

					return nil
				},
			},
			{
				Name:  "get",
				Usage: "Extract specific file or directory from vault",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "vault",
						Aliases:  []string{"v"},
						Usage:    "Path to vault file",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "Vault password (NOT RECOMMENDED, better to enter interactively for security)",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "target",
						Aliases:  []string{"t"},
						Usage:    "Path to file or directory in vault to extract",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "Directory to extract files",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					vaultPath := cmd.String("vault")
					password := cmd.String("password")
					targetPath := cmd.String("target")
					outputDir := cmd.String("output")

					if password == "" {
						var err error
						password, err = vault.ReadPasswordSecurely("Enter vault password: ")
						if err != nil {
							return err
						}
					}

					fmt.Printf("Extracting '%s' to directory: %s\n", targetPath, outputDir)

					if err := vault.ExtractSpecific(vaultPath, password, targetPath, outputDir); err != nil {
						return fmt.Errorf("extraction error: %w", err)
					}

					fmt.Printf("✅ '%s' successfully extracted to '%s'!\n", targetPath, outputDir)

					return nil
				},
			},
			{
				Name:  "remove",
				Usage: "Remove file or directory from vault",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "vault",
						Aliases:  []string{"v"},
						Usage:    "Path to vault file",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "Vault password (NOT RECOMMENDED, better to enter interactively for security)",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "target",
						Aliases:  []string{"t"},
						Usage:    "Path to file or directory in vault to remove",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					vaultPath := cmd.String("vault")
					password := cmd.String("password")
					targetPath := cmd.String("target")

					if password == "" {
						var err error
						password, err = vault.ReadPasswordSecurely("Enter vault password: ")
						if err != nil {
							return err
						}
					}

					fmt.Printf("Removing '%s' from vault...\n", targetPath)

					if err := vault.RemoveFromVault(vaultPath, password, targetPath); err != nil {
						return fmt.Errorf("removal error: %w", err)
					}

					fmt.Printf("✅ '%s' successfully removed from vault!\n", targetPath)

					return nil
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

// formatSize formats file size in human-readable form
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
