# 📚 API Documentation

This document provides comprehensive API documentation for Flint Vault, including Go package interfaces and examples.

## 📋 Table of Contents

- [Package Overview](#package-overview)
- [Core Types](#core-types)
- [Vault Operations](#vault-operations)
- [Parallel Processing](#parallel-processing)
- [File Operations](#file-operations)
- [Utility Functions](#utility-functions)
- [Error Handling](#error-handling)
- [Examples](#examples)

## 📦 Package Overview

Flint Vault provides a clean, unified Go API for encrypted file storage with optimized performance and parallel processing capabilities.

```go
import "flint-vault/pkg/lib/vault"
```

**Key Features:**
- Unified API with all operations in single module
- Memory-optimized streaming operations
- **Parallel processing** with configurable worker pools
- Support for multi-GB files
- Built-in compression support
- **Progress reporting** for long-running operations
- Comprehensive error handling

## 🏗️ Core Types

### FileEntry

```go
type FileEntry struct {
    Path           string    `json:"path"`            // Path to file/directory
    Name           string    `json:"name"`            // Name of file/directory
    IsDir          bool      `json:"is_dir"`          // Whether it's a directory
    Size           int64     `json:"size"`            // Original file size
    CompressedSize int64     `json:"compressed_size"` // Size after compression
    Mode           uint32    `json:"mode"`            // Access permissions
    ModTime        time.Time `json:"mod_time"`        // Last modification time
    Offset         int64     `json:"offset"`          // Offset in vault file
    SHA256Hash     [32]byte  `json:"sha256_hash"`     // SHA-256 hash for integrity
}
```

### VaultDirectory

```go
type VaultDirectory struct {
    Version   uint32      `json:"version"`    // Vault format version
    Entries   []FileEntry `json:"entries"`    // File/directory metadata
    CreatedAt time.Time   `json:"created_at"` // Vault creation time
    Comment   string      `json:"comment"`    // Vault comment
}
```

### ParallelConfig

```go
type ParallelConfig struct {
    MaxConcurrency int             // Maximum number of concurrent workers
    Timeout        time.Duration   // Timeout for individual operations
    ProgressChan   chan string     // Progress reporting channel (optional)
    Context        context.Context // Context for cancellation
}
```

### ParallelStats

```go
type ParallelStats struct {
    TotalFiles      int64         // Total files processed
    SuccessfulFiles int64         // Successfully processed files
    FailedFiles     int64         // Failed files
    TotalSize       int64         // Total size processed (bytes)
    Duration        time.Duration // Total processing duration
    Errors          []error       // Collection of errors encountered
    ErrorsMutex     sync.Mutex    // Mutex for thread-safe error collection
}
```

### VaultInfo

```go
type VaultInfo struct {
    FilePath      string `json:"file_path"`
    FileSize      int64  `json:"file_size"`
    IsFlintVault  bool   `json:"is_flint_vault"`
    Version       int    `json:"version"`
    Iterations    int    `json:"iterations"`
}
```

## 🔐 Vault Operations

### CreateVault

Creates a new encrypted vault file with military-grade security.

```go
func CreateVault(vaultPath, password string) error
```

**Parameters:**
- `vaultPath`: Path where the vault file will be created
- `password`: Master password for vault encryption

**Security Features:**
- AES-256-GCM encryption
- PBKDF2 key derivation (100,000 iterations)
- Cryptographically secure salt generation

**Example:**
```go
err := vault.CreateVault("my-vault.flint", "secure-password")
if err != nil {
    log.Fatalf("Failed to create vault: %v", err)
}
fmt.Println("✅ Vault created successfully!")
```

### ListVault

Lists all contents of an encrypted vault with metadata.

```go
func ListVault(vaultPath, password string) ([]FileEntry, error)
```

**Returns:**
- `[]FileEntry`: Slice of vault entries with full metadata
- `error`: Error if vault cannot be opened or password is incorrect

**Example:**
```go
entries, err := vault.ListVault("my-vault.flint", "secure-password")
if err != nil {
    log.Fatalf("Failed to list vault: %v", err)
}

fmt.Printf("📦 Vault contains %d items:\n", len(entries))
for _, entry := range entries {
    icon := "📄"
    if entry.IsDir {
        icon = "📁"
    }
    fmt.Printf("  %s %s (%d bytes)\n", icon, entry.Path, entry.Size)
}
```

### ValidateVaultFile

Validates vault file format without requiring password.

```go
func ValidateVaultFile(vaultPath string) error
```

**Example:**
```go
if err := vault.ValidateVaultFile("my-vault.flint"); err != nil {
    log.Printf("⚠️  Vault validation failed: %v", err)
} else {
    fmt.Println("✅ Vault file is valid")
}
```

## 🚀 Parallel Processing

### DefaultParallelConfig

Creates default parallel processing configuration.

```go
func DefaultParallelConfig() *ParallelConfig
```

**Example:**
```go
config := vault.DefaultParallelConfig()
// Uses 2x CPU cores, 5-minute timeout
fmt.Printf("Using %d workers\n", config.MaxConcurrency)
```

### AddDirectoryToVaultParallel

Adds directory to vault with optimized parallel processing.

```go
func AddDirectoryToVaultParallel(vaultPath, password, dirPath string, config *ParallelConfig) (*ParallelStats, error)
```

**Features:**
- Configurable worker pools
- Progress reporting
- Comprehensive statistics
- Error collection and reporting

**Example:**
```go
config := vault.DefaultParallelConfig()
config.MaxConcurrency = 8 // Use 8 workers

// Optional: Set up progress reporting
progressChan := make(chan string, 100)
config.ProgressChan = progressChan

go func() {
    for msg := range progressChan {
        fmt.Printf("🔄 %s\n", msg)
    }
}()

stats, err := vault.AddDirectoryToVaultParallel(
    "my-vault.flint", 
    "password", 
    "./large-directory/", 
    config)

close(progressChan)

if err != nil {
    log.Fatalf("Parallel add failed: %v", err)
}

vault.PrintParallelStats(stats)
```

### ExtractMultipleFilesFromVaultParallel

Extracts multiple files from vault in parallel.

```go
func ExtractMultipleFilesFromVaultParallel(vaultPath, password, outputDir string, targetPaths []string, config *ParallelConfig) (*ParallelStats, error)
```

**Example:**
```go
config := vault.DefaultParallelConfig()
config.MaxConcurrency = 6

targets := []string{
    "documents/report.pdf",
    "images/photo.jpg",
    "data/large-dataset.csv",
}

stats, err := vault.ExtractMultipleFilesFromVaultParallel(
    "my-vault.flint",
    "password",
    "./extracted/",
    targets,
    config)

if err != nil {
    log.Fatalf("Parallel extraction failed: %v", err)
}

fmt.Printf("✅ Extracted %d files in %v\n", 
    stats.SuccessfulFiles, stats.Duration)
```

### PrintParallelStats

Prints detailed statistics from parallel operations.

```go
func PrintParallelStats(stats *ParallelStats)
```

**Example Output:**
```
📊 Operation Statistics:
✅ Successfully processed: 245 files
❌ Failed: 0 files
📏 Total size: 1.2 GB
⏱️  Duration: 12.3 seconds
📈 Average speed: 97.6 MB/s
🔧 Workers utilized: 8
```

## 📁 File Operations

### AddFileToVault

Adds a single file to the vault with compression and encryption.

```go
func AddFileToVault(vaultPath, password, filePath string) error
```

**Features:**
- Automatic compression for space efficiency
- Metadata preservation (timestamps, permissions)
- Memory-efficient streaming for large files

**Example:**
```go
err := vault.AddFileToVault("my-vault.flint", "password", "documents/report.pdf")
if err != nil {
    log.Fatalf("Failed to add file: %v", err)
}
fmt.Println("✅ File added successfully!")
```

### AddDirectoryToVault

Recursively adds a directory and all its contents to the vault.

```go
func AddDirectoryToVault(vaultPath, password, dirPath string) error
```

**Features:**
- Recursive directory traversal
- Preserves directory structure
- Handles large directory trees efficiently

**Example:**
```go
err := vault.AddDirectoryToVault("my-vault.flint", "password", "project/")
if err != nil {
    log.Fatalf("Failed to add directory: %v", err)
}
fmt.Println("✅ Directory added successfully!")
```

### ExtractFromVault

Extracts all files from the vault to a specified directory.

```go
func ExtractFromVault(vaultPath, password, outputDir string) error
```

**Features:**
- Recreates original directory structure
- Restores file metadata (timestamps, permissions)
- Memory-efficient streaming extraction

**Example:**
```go
err := vault.ExtractFromVault("my-vault.flint", "password", "./extracted/")
if err != nil {
    log.Fatalf("Failed to extract: %v", err)
}
fmt.Println("✅ All files extracted successfully!")
```

### GetFromVault

Extracts specific files or directories from the vault.

```go
func GetFromVault(vaultPath, password, outputDir string, targets []string) error
```

**Parameters:**
- `vaultPath`: Path to the vault file
- `password`: Vault password
- `outputDir`: Directory where files will be extracted
- `targets`: Slice of file/directory paths to extract

**Features:**
- Selective extraction
- Supports multiple targets in single operation
- Maintains directory structure for extracted items

**Example:**
```go
targets := []string{"documents/report.pdf", "images/", "config.json"}
err := vault.GetFromVault("my-vault.flint", "password", "./output/", targets)
if err != nil {
    log.Fatalf("Extraction failed: %v", err)
}
fmt.Println("✅ Selected files extracted successfully!")
```

### RemoveFromVault

Removes specified files or directories from the vault.

```go
func RemoveFromVault(vaultPath, password string, targets []string) error
```

**Features:**
- Supports multiple targets
- Directory removal includes all contents
- Efficient vault reorganization after removal

**Example:**
```go
targets := []string{"old-file.txt", "temp-directory/"}
err := vault.RemoveFromVault("my-vault.flint", "password", targets)
if err != nil {
    log.Fatalf("Removal failed: %v", err)
}
fmt.Println("✅ Files removed successfully!")
```

## 🛠️ Utility Functions

### GetVaultInfo

Retrieves vault information without requiring password.

```go
func GetVaultInfo(filePath string) (*VaultInfo, error)
```

**Features:**
- Password-free operation
- File format validation
- Metadata extraction

**Example:**
```go
info, err := vault.GetVaultInfo("my-vault.flint")
if err != nil {
    log.Fatalf("Failed to get vault info: %v", err)
}

fmt.Printf("📁 File: %s\n", info.FilePath)
fmt.Printf("📏 Size: %d bytes\n", info.FileSize)
fmt.Printf("✅ Valid Flint Vault: %v\n", info.IsFlintVault)
fmt.Printf("🔢 Version: %d\n", info.Version)
fmt.Printf("🔐 PBKDF2 Iterations: %d\n", info.Iterations)
```

### ReadPasswordSecurely

Reads password from terminal without echoing characters.

```go
func ReadPasswordSecurely(prompt string) (string, error)
```

**Features:**
- Hidden password input
- Cross-platform compatibility
- Secure memory handling

**Example:**
```go
password, err := vault.ReadPasswordSecurely("Enter vault password: ")
if err != nil {
    log.Fatalf("Failed to read password: %v", err)
}
// Use password for vault operations
```

## ⚠️ Error Handling

### Common Errors

```go
// Authentication errors
ErrInvalidPassword    = errors.New("invalid password")
ErrCorruptedVault     = errors.New("vault file is corrupted")

// File operation errors
ErrFileNotFound       = errors.New("file not found in vault")
ErrFileAlreadyExists  = errors.New("file already exists in vault")

// System errors
ErrInsufficientSpace  = errors.New("insufficient disk space")
ErrPermissionDenied   = errors.New("permission denied")
```

### Error Checking Pattern

```go
if err := vault.AddFileToVault(vaultPath, password, filePath); err != nil {
    switch {
    case strings.Contains(err.Error(), "permission denied"):
        log.Printf("❌ Permission error: %v", err)
    case strings.Contains(err.Error(), "invalid password"):
        log.Printf("❌ Authentication error: %v", err)
    case strings.Contains(err.Error(), "not found"):
        log.Printf("❌ File not found: %v", err)
    default:
        log.Printf("❌ Unexpected error: %v", err)
    }
    return err
}
```

## 💡 Complete Examples

### High-Performance Parallel Operations

```go
package main

import (
    "fmt"
    "log"
    "runtime"
    "flint-vault/pkg/lib/vault"
)

func main() {
    vaultPath := "high-performance.flint"
    password := "secure-password-123"

    // Create vault
    fmt.Println("🔐 Creating vault...")
    if err := vault.CreateVault(vaultPath, password); err != nil {
        log.Fatalf("Create failed: %v", err)
    }

    // Configure high-performance parallel processing
    config := vault.DefaultParallelConfig()
    config.MaxConcurrency = runtime.NumCPU() * 2 // 2x CPU cores
    
    // Set up progress monitoring
    progressChan := make(chan string, 100)
    config.ProgressChan = progressChan

    // Start progress reporter
    go func() {
        for msg := range progressChan {
            fmt.Printf("🔄 %s\n", msg)
        }
    }()

    // Add large directory with parallel processing
    fmt.Println("📁 Adding directory with parallel processing...")
    stats, err := vault.AddDirectoryToVaultParallel(
        vaultPath, 
        password, 
        "./large-dataset/", 
        config)

    close(progressChan)

    if err != nil {
        log.Fatalf("Parallel add failed: %v", err)
    }

    // Print performance statistics
    vault.PrintParallelStats(stats)

    // Parallel extraction of specific files
    fmt.Println("\n🔓 Parallel extraction of specific files...")
    
    progressChan2 := make(chan string, 100)
    config.ProgressChan = progressChan2

    go func() {
        for msg := range progressChan2 {
            fmt.Printf("🔄 %s\n", msg)
        }
    }()

    targets := []string{
        "data/file1.bin",
        "data/file2.bin", 
        "documents/",
    }

    extractStats, err := vault.ExtractMultipleFilesFromVaultParallel(
        vaultPath,
        password,
        "./extracted-parallel/",
        targets,
        config)

    close(progressChan2)

    if err != nil {
        log.Fatalf("Parallel extraction failed: %v", err)
    }

    vault.PrintParallelStats(extractStats)

    fmt.Println("🎉 High-performance operations completed!")
}
```

### Custom Worker Configuration

```go
func customWorkerExample() {
    // For CPU-intensive operations (heavy compression)
    cpuConfig := vault.DefaultParallelConfig()
    cpuConfig.MaxConcurrency = runtime.NumCPU()

    // For I/O-intensive operations (large files)
    ioConfig := vault.DefaultParallelConfig()
    ioConfig.MaxConcurrency = runtime.NumCPU() * 3

    // For memory-constrained environments
    memoryConfig := vault.DefaultParallelConfig()
    memoryConfig.MaxConcurrency = 2

    // For maximum throughput (sufficient resources)
    maxConfig := vault.DefaultParallelConfig()
    maxConfig.MaxConcurrency = 16
    
    // Use appropriate config based on operation type
    stats, err := vault.AddDirectoryToVaultParallel(
        "vault.flint", 
        "password", 
        "./data/", 
        ioConfig) // Use I/O optimized config
        
    // Handle results...
}
```

### Progress Monitoring and Error Handling

```go
func progressMonitoringExample() {
    config := vault.DefaultParallelConfig()
    
    // Advanced progress monitoring
    progressChan := make(chan string, 200)
    config.ProgressChan = progressChan
    
    // Statistics tracking
    var fileCount int64
    var totalSize int64
    
    go func() {
        for msg := range progressChan {
            // Custom progress processing
            if strings.Contains(msg, "Adding:") {
                fileCount++
                fmt.Printf("[%d] %s\n", fileCount, msg)
            } else if strings.Contains(msg, "Processing") {
                fmt.Printf("📊 %s\n", msg)
            }
        }
    }()
    
    stats, err := vault.AddDirectoryToVaultParallel(
        "monitored-vault.flint",
        "password",
        "./source-data/",
        config)
    
    close(progressChan)
    
    if err != nil {
        log.Printf("❌ Operation failed: %v", err)
        
        // Handle partial success
        if stats != nil && stats.SuccessfulFiles > 0 {
            fmt.Printf("⚠️  Partial success: %d files added\n", 
                stats.SuccessfulFiles)
        }
        return
    }
    
    // Success statistics
    fmt.Printf("✅ Complete success:\n")
    fmt.Printf("   Files: %d\n", stats.TotalFiles)
    fmt.Printf("   Size: %d bytes\n", stats.TotalSize)
    fmt.Printf("   Speed: %.2f MB/s\n", 
        float64(stats.TotalSize)/1024/1024/stats.Duration.Seconds())
}
```

---

*API Documentation updated for unified vault architecture*  
*Last updated: June 2025*
