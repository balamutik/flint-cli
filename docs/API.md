# 📚 API Documentation

This document provides comprehensive API documentation for Flint Vault, including Go package interfaces and examples.

## 📋 Table of Contents

- [Package Overview](#package-overview)
- [Core Types](#core-types)
- [Vault Operations](#vault-operations)
- [File Operations](#file-operations)
- [Error Handling](#error-handling)
- [Examples](#examples)

## 📦 Package Overview

Flint Vault provides a clean Go API for encrypted file storage.

```go
import "flint-vault/pkg/lib/vault"
```

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

## 🔐 Vault Operations

### CreateVault

Creates a new encrypted vault file.

```go
func CreateVault(vaultPath, password string) error
```

**Example:**
```go
err := vault.CreateVault("my-vault.dat", "secure-password")
if err != nil {
    log.Fatalf("Failed to create vault: %v", err)
}
```

### ListVault

Lists all contents of an encrypted vault.

```go
func ListVault(vaultPath, password string) (*VaultData, error)
```

**Example:**
```go
data, err := vault.ListVault("my-vault.dat", "secure-password")
if err != nil {
    log.Fatalf("Failed to list vault: %v", err)
}

for _, entry := range data.Entries {
    if entry.IsDir {
        fmt.Printf("📁 %s/\n", entry.Name)
    } else {
        fmt.Printf("📄 %s (%d bytes)\n", entry.Name, entry.Size)
    }
}
```

## 📁 File Operations

### AddFileToVault

```go
func AddFileToVault(vaultPath, password, filePath string) error
```

### AddDirectoryToVault

```go
func AddDirectoryToVault(vaultPath, password, dirPath string) error
```

### ExtractVault

```go
func ExtractVault(vaultPath, password, destDir string) error
```

### ExtractSpecific

```go
func ExtractSpecific(vaultPath, password, targetPath, destDir string) error
```

### ExtractMultiple

```go
func ExtractMultiple(vaultPath, password string, targetPaths []string, destDir string) ([]string, []string, error)
```

**Parameters:**
- `vaultPath`: Path to the vault file
- `password`: Vault password
- `targetPaths`: Slice of file/directory paths to extract
- `destDir`: Destination directory for extraction

**Returns:**
- `[]string`: Successfully extracted paths
- `[]string`: Paths not found in vault
- `error`: Error if operation fails

**Usage Example:**
```go
paths := []string{"doc1.pdf", "images/", "config.json"}
extracted, notFound, err := vault.ExtractMultiple("vault.dat", "password", paths, "./output/")
if err != nil {
    log.Fatalf("Extraction failed: %v", err)
}
fmt.Printf("Extracted %d items, %d not found\n", len(extracted), len(notFound))
```

### RemoveFromVault

```go
func RemoveFromVault(vaultPath, password, targetPath string) error
```

### GetFileFromVault

```go
func GetFileFromVault(vaultPath, password, targetPath string) ([]byte, error)
```

## 💡 Complete Example

```go
package main

import (
    "fmt"
    "log"
    "flint-vault/pkg/lib/vault"
)

func main() {
    vaultPath := "example.vault"
    password := "secure-password-123"

    // Create vault
    if err := vault.CreateVault(vaultPath, password); err != nil {
        log.Fatalf("Create failed: %v", err)
    }

    // Add file
    if err := vault.AddFileToVault(vaultPath, password, "document.pdf"); err != nil {
        log.Fatalf("Add failed: %v", err)
    }

    // List contents
    data, err := vault.ListVault(vaultPath, password)
    if err != nil {
        log.Fatalf("List failed: %v", err)
    }

    fmt.Printf("Vault contains %d items\n", len(data.Entries))
} 