// Package commands implements the command-line interface for Flint Vault.
// This package provides all CLI commands and their implementations using the urfave/cli framework.
//
// Available commands:
//   - create: Create new encrypted vault
//   - add: Add files or directories to vault (with high-performance batch processing)
//   - list: Show vault contents
//   - extract: Extract files from vault (with parallel processing)
//   - remove: Remove files or directories from vault
//   - info: Show vault file information without password
//
// All commands use optimized batch processing and provide comprehensive error handling.
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
		Usage: "Military-grade encrypted file storage with AES-256",
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
						Usage:    "Encryption password (NOT RECOMMENDED, better to enter interactively)",
						Required: false,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					file := cmd.String("file")
					password := cmd.String("password")

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
				Usage: "Add files or directories to vault with high-performance batch processing",
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
						Usage:    "Vault password (NOT RECOMMENDED, better to enter interactively)",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "source",
						Aliases:  []string{"s"},
						Usage:    "Path to file or directory to add",
						Required: true,
					},
					&cli.IntFlag{
						Name:    "workers",
						Aliases: []string{"w"},
						Usage:   "Number of parallel workers (0 = auto-detect)",
						Value:   0,
					},
					&cli.BoolFlag{
						Name:  "progress",
						Usage: "Show progress information",
						Value: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					vaultPath := cmd.String("vault")
					password := cmd.String("password")
					sourcePath := cmd.String("source")
					workers := cmd.Int("workers")
					showProgress := cmd.Bool("progress")

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
						return fmt.Errorf("source not found: %s", sourcePath)
					}

					// Configure parallel processing
					config := vault.DefaultParallelConfig()
					if workers > 0 {
						config.MaxConcurrency = workers
					}

					var progressChan chan string
					if showProgress {
						progressChan = make(chan string, 100)
						config.ProgressChan = progressChan

						go func() {
							for msg := range progressChan {
								fmt.Printf("🔄 %s\n", msg)
							}
						}()
					}

					if info.IsDir() {
						fmt.Printf("Adding directory '%s' to vault (workers: %d)...\n", sourcePath, config.MaxConcurrency)

						var stats *vault.ParallelStats
						var err error

						stats, err = vault.AddDirectoryToVaultParallel(vaultPath, password, sourcePath, config)

						if showProgress {
							close(progressChan)
						}

						if err != nil {
							return fmt.Errorf("directory add error: %w", err)
						}

						vault.PrintParallelStats(stats)
					} else {
						fmt.Printf("Adding file '%s' to vault...\n", sourcePath)
						if err := vault.AddFileToVault(vaultPath, password, sourcePath); err != nil {
							return fmt.Errorf("file add error: %w", err)
						}
						fmt.Printf("✅ File successfully added to vault!\n")
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
						Usage:    "Vault password (NOT RECOMMENDED, better to enter interactively)",
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

					entries, err := vault.ListVault(vaultPath, password)
					if err != nil {
						return fmt.Errorf("vault read error: %w", err)
					}

					fmt.Printf("📦 Vault: %s\n", vaultPath)
					fmt.Printf("📁 Contents (%d items):\n\n", len(entries))

					if len(entries) == 0 {
						fmt.Println("  Vault is empty")
						return nil
					}

					for _, entry := range entries {
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
				Usage: "Extract files from vault (optimized with parallel processing)",
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
						Usage:    "Vault password (NOT RECOMMENDED, better to enter interactively)",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "Directory to extract files",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:    "files",
						Aliases: []string{"f"},
						Usage:   "Specific files to extract (if not specified, extracts all)",
					},
					&cli.IntFlag{
						Name:    "workers",
						Aliases: []string{"w"},
						Usage:   "Number of parallel workers (0 = auto-detect)",
						Value:   0,
					},
					&cli.BoolFlag{
						Name:  "progress",
						Usage: "Show progress information",
						Value: true,
					},
					&cli.BoolFlag{
						Name:  "extract-full-path",
						Usage: "Extract files with full directory structure (default: false, extracts only filenames)",
						Value: false,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					vaultPath := cmd.String("vault")
					password := cmd.String("password")
					outputDir := cmd.String("output")
					specificFiles := cmd.StringSlice("files")
					workers := cmd.Int("workers")
					showProgress := cmd.Bool("progress")
					extractFullPath := cmd.Bool("extract-full-path")

					if password == "" {
						var err error
						password, err = vault.ReadPasswordSecurely("Enter vault password: ")
						if err != nil {
							return err
						}
					}

					// Configure parallel processing
					config := vault.DefaultParallelConfig()
					if workers > 0 {
						config.MaxConcurrency = workers
					}

					var progressChan chan string
					if showProgress {
						progressChan = make(chan string, 100)
						config.ProgressChan = progressChan

						go func() {
							for msg := range progressChan {
								fmt.Printf("🔄 %s\n", msg)
							}
						}()
					}

					if len(specificFiles) > 0 {
						// Extract specific files in parallel
						fmt.Printf("Extracting %d specific files (workers: %d, full-path: %v)...\n", len(specificFiles), config.MaxConcurrency, extractFullPath)
						stats, err := vault.ExtractMultipleFilesFromVaultParallelWithOptions(vaultPath, password, outputDir, specificFiles, config, extractFullPath)

						if showProgress {
							close(progressChan)
						}

						if err != nil {
							return fmt.Errorf("parallel extraction error: %w", err)
						}

						vault.PrintParallelStats(stats)
					} else {
						// Extract all files using optimized streaming
						fmt.Printf("Extracting all files to: %s (full-path: %v)\n", outputDir, extractFullPath)
						if err := vault.ExtractFromVaultWithOptions(vaultPath, password, outputDir, extractFullPath); err != nil {
							return fmt.Errorf("extraction error: %w", err)
						}
						fmt.Printf("✅ All files successfully extracted!\n")
					}

					return nil
				},
			},
			{
				Name:  "remove",
				Usage: "Remove files or directories from vault",
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
						Usage:    "Vault password (NOT RECOMMENDED, better to enter interactively)",
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

					if err := vault.RemoveFromVault(vaultPath, password, []string{targetPath}); err != nil {
						return fmt.Errorf("removal error: %w", err)
					}

					fmt.Printf("✅ '%s' successfully removed from vault!\n", targetPath)
					return nil
				},
			},
			{
				Name:  "info",
				Usage: "Show vault file information without requiring password",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "Path to file to check",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					filePath := cmd.String("file")

					fmt.Printf("🔍 Analyzing file: %s\n\n", filePath)

					info, err := vault.GetVaultInfo(filePath)
					if err != nil {
						return fmt.Errorf("file analysis error: %w", err)
					}

					fmt.Printf("📁 File Path: %s\n", info.FilePath)
					fmt.Printf("📏 File Size: %s\n", formatSize(info.FileSize))

					if info.IsFlintVault {
						fmt.Printf("✅ File Type: Flint Vault encrypted storage\n")
						fmt.Printf("🔢 Format Version: %d\n", info.Version)
						fmt.Printf("🔐 PBKDF2 Iterations: %s\n", formatNumber(int64(info.Iterations)))

						if err := vault.ValidateVaultFile(filePath); err != nil {
							fmt.Printf("⚠️  Validation: Failed - %v\n", err)
						} else {
							fmt.Printf("✅ Validation: Passed\n")
						}

						fmt.Printf("\n💡 This file can be opened with 'flint-vault list' command\n")
					} else {
						fmt.Printf("❌ File Type: Not a Flint Vault file\n")
						fmt.Printf("\n💡 This file cannot be opened by Flint Vault\n")
					}

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

// formatNumber formats large numbers with thousand separators
func formatNumber(num int64) string {
	str := fmt.Sprintf("%d", num)
	if len(str) <= 3 {
		return str
	}

	var result []rune
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, digit)
	}
	return string(result)
}

// formatFileSize formats file size in human-readable format
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.1fTB", float64(size)/TB)
	case size >= GB:
		return fmt.Sprintf("%.1fGB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.1fMB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.1fKB", float64(size)/KB)
	default:
		return fmt.Sprintf("%dB", size)
	}
}
