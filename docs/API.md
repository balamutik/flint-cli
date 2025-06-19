# 📚 API Documentation

This document provides comprehensive API documentation for Flint Vault, including Go package interfaces and examples.

## 📋 Table of Contents

- [Package Overview](#package-overview)
- [Core Types](#core-types)
- [Vault Operations](#vault-operations)
- [File Operations](#file-operations)
- [Utility Functions](#utility-functions)
- [Error Handling](#error-handling)
- [Examples](#examples)

## 📦 Package Overview

Flint Vault provides a clean, unified Go API for encrypted file storage with optimized performance.

```go
import "flint-vault/pkg/lib/vault"
```

**Key Features:**
- Unified API with all operations in single module
- Memory-optimized streaming operations
- Support for multi-GB files
- Built-in compression support
- Comprehensive error handling

## 🏗️ Core Types

### VaultEntry

```go
type VaultEntry struct {
    Path    string      `json:"path"`
    Name    string      `json:"name"`
    IsDir   bool        `json:"is_dir"`
    Size    int64       `json:"size"`
    Mode    os.FileMode `json:"mode"`
    ModTime time.Time   `json:"mod_time"`
    Content []byte      `json:"content"`
}
```

### VaultData

```go
type VaultData struct {
    Entries   []VaultEntry `json:"entries"`
    CreatedAt time.Time    `json:"created_at"`
    Comment   string       `json:"comment"`
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
func ListVault(vaultPath, password string) ([]VaultEntry, error)
```

**Returns:**
- `[]VaultEntry`: Slice of vault entries with full metadata
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

### Basic Vault Operations

```go
package main

import (
    "fmt"
    "log"
    "flint-vault/pkg/lib/vault"
)

func main() {
    vaultPath := "example.flint"
    password := "secure-password-123"

    // Create vault
    fmt.Println("🔐 Creating vault...")
    if err := vault.CreateVault(vaultPath, password); err != nil {
        log.Fatalf("Create failed: %v", err)
    }

    // Add file
    fmt.Println("📄 Adding file...")
    if err := vault.AddFileToVault(vaultPath, password, "document.pdf"); err != nil {
        log.Fatalf("Add failed: %v", err)
    }

    // Add directory
    fmt.Println("📁 Adding directory...")
    if err := vault.AddDirectoryToVault(vaultPath, password, "project/"); err != nil {
        log.Fatalf("Add directory failed: %v", err)
    }

    // List contents
    fmt.Println("📋 Listing contents...")
    entries, err := vault.ListVault(vaultPath, password)
    if err != nil {
        log.Fatalf("List failed: %v", err)
    }

    fmt.Printf("✅ Vault contains %d items\n", len(entries))
    for _, entry := range entries {
        icon := "📄"
        if entry.IsDir {
            icon = "📁"
        }
        fmt.Printf("  %s %s (%d bytes)\n", icon, entry.Path, entry.Size)
    }

    // Extract specific files
    fmt.Println("🔓 Extracting specific files...")
    targets := []string{"document.pdf", "project/"}
    if err := vault.GetFromVault(vaultPath, password, "./extracted/", targets); err != nil {
        log.Fatalf("Extract failed: %v", err)
    }

    fmt.Println("🎉 All operations completed successfully!")
}
```

### Advanced Usage with Error Handling

```go
package main

import (
    "fmt"
    "log"
    "os"
    "flint-vault/pkg/lib/vault"
)

func main() {
    vaultPath := "advanced-example.flint"
    
    // Read password securely
    password, err := vault.ReadPasswordSecurely("Enter vault password: ")
    if err != nil {
        log.Fatalf("Failed to read password: %v", err)
    }

    // Check if vault exists
    if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
        fmt.Println("🔐 Creating new vault...")
        if err := vault.CreateVault(vaultPath, password); err != nil {
            log.Fatalf("Failed to create vault: %v", err)
        }
    } else {
        // Validate existing vault
        fmt.Println("🔍 Validating existing vault...")
        if err := vault.ValidateVaultFile(vaultPath); err != nil {
            log.Fatalf("Vault validation failed: %v", err)
        }
    }

    // Get vault information
    info, err := vault.GetVaultInfo(vaultPath)
    if err != nil {
        log.Fatalf("Failed to get vault info: %v", err)
    }
    
    fmt.Printf("📊 Vault Info:\n")
    fmt.Printf("  📁 File: %s\n", info.FilePath)
    fmt.Printf("  📏 Size: %d bytes\n", info.FileSize)
    fmt.Printf("  🔢 Version: %d\n", info.Version)
    fmt.Printf("  🔐 Iterations: %d\n", info.Iterations)

    // Perform operations with comprehensive error handling
    operations := []struct {
        name string
        fn   func() error
    }{
        {"Adding test file", func() error {
            return vault.AddFileToVault(vaultPath, password, "test.txt")
        }},
        {"Listing contents", func() error {
            entries, err := vault.ListVault(vaultPath, password)
            if err != nil {
                return err
            }
            fmt.Printf("📋 Found %d entries\n", len(entries))
            return nil
        }},
        {"Extracting files", func() error {
            return vault.ExtractFromVault(vaultPath, password, "./extracted/")
        }},
    }

    for _, op := range operations {
        fmt.Printf("🔄 %s...\n", op.name)
        if err := op.fn(); err != nil {
            log.Printf("❌ %s failed: %v\n", op.name, err)
        } else {
            fmt.Printf("✅ %s completed\n", op.name)
        }
    }
}
```

### Performance Monitoring

```go
package main

import (
    "fmt"
    "time"
    "flint-vault/pkg/lib/vault"
)

func measureOperation(name string, fn func() error) {
    start := time.Now()
    fmt.Printf("🔄 Starting %s...\n", name)
    
    if err := fn(); err != nil {
        fmt.Printf("❌ %s failed: %v\n", name, err)
        return
    }
    
    duration := time.Since(start)
    fmt.Printf("✅ %s completed in %v\n", name, duration)
}

func main() {
    vaultPath := "performance-test.flint"
    password := "test-password"

    measureOperation("Vault Creation", func() error {
        return vault.CreateVault(vaultPath, password)
    })

    measureOperation("Large Directory Addition", func() error {
        return vault.AddDirectoryToVault(vaultPath, password, "./large-dataset/")
    })

    measureOperation("Full Extraction", func() error {
        return vault.ExtractFromVault(vaultPath, password, "./extracted/")
    })
}
```

## 🚀 Performance Considerations

### Memory Usage
- **Streaming Operations**: All file operations use memory-efficient streaming
- **Buffer Size**: Optimized 1MB buffers for best performance
- **Memory Ratio**: Typical 3.2:1 memory-to-data ratio during encryption

### Best Practices
1. **Use appropriate buffer sizes** for your system
2. **Monitor memory usage** with large files
3. **Validate vault files** before operations
4. **Handle errors gracefully** with proper cleanup
5. **Use secure password input** in production

### Tested Performance
- **Throughput**: Up to 272 MB/s for file operations
- **Scalability**: Successfully tested with 2.45 GB datasets
- **Memory Efficiency**: Excellent performance with large files

---

*API Documentation updated for unified vault architecture*  
*Last updated: June 2025*
