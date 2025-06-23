# 📂 File Extraction Modes Guide

This guide explains the different file extraction modes available in Flint Vault.

## 🎯 Overview

Flint Vault now supports two extraction modes:
- **Flat extraction** (default): Extracts only filenames without directory structure
- **Full-path extraction**: Preserves the original directory structure

## 🔧 How to Use

### CLI Command

```bash
# Flat extraction (default)
flint-vault extract -v vault.flint -o ./output/

# Full-path extraction  
flint-vault extract -v vault.flint -o ./output/ --extract-full-path
```

### API Usage

```go
// Flat extraction (default)
err := vault.ExtractFromVaultWithOptions(vaultPath, password, outputDir, false)

// Full-path extraction
err := vault.ExtractFromVaultWithOptions(vaultPath, password, outputDir, true)
```

## 📁 Extraction Examples

### Example Vault Contents
```
documents/
├── reports/
│   ├── 2025-report.pdf
│   └── summary.txt
├── images/
│   ├── chart1.png
│   └── diagram.svg
└── config.json
```

### Flat Extraction (Default)
```bash
flint-vault extract -v example.flint -o ./flat-output/
```

**Result:**
```
flat-output/
├── 2025-report.pdf
├── summary.txt
├── chart1.png
├── diagram.svg
└── config.json
```

### Full-Path Extraction
```bash
flint-vault extract -v example.flint -o ./structured-output/ --extract-full-path
```

**Result:**
```
structured-output/
└── documents/
    ├── reports/
    │   ├── 2025-report.pdf
    │   └── summary.txt
    ├── images/
    │   ├── chart1.png
    │   └── diagram.svg
    └── config.json
```

## 🎯 When to Use Each Mode

### Flat Extraction (Default) 
**Use when:**
- ✅ You need quick access to all files
- ✅ Directory structure is not important
- ✅ Processing files individually 
- ✅ Working with scripts that expect flat structure
- ✅ Analyzing file contents without organization

**Examples:**
```bash
# Quick file analysis
flint-vault extract -v logs.flint -o ./temp/
grep "ERROR" ./temp/*.log

# Batch processing
flint-vault extract -v images.flint -o ./process/
for img in ./process/*.jpg; do convert "$img" "${img%.jpg}.png"; done

# Simple backup restoration
flint-vault extract -v backup.flint -o ./restore/
```

### Full-Path Extraction
**Use when:**
- ✅ You need to preserve project structure
- ✅ Restoring archives or backups
- ✅ Directory organization is important
- ✅ Working with projects that depend on file locations
- ✅ Maintaining relative file references

**Examples:**
```bash
# Project restoration
flint-vault extract -v project-backup.flint -o ./restored-project/ --extract-full-path

# Website backup
flint-vault extract -v website.flint -o ./site/ --extract-full-path

# Configuration backup
flint-vault extract -v system-config.flint -o ./ --extract-full-path
```

## 🚀 Advanced Usage

### Selective Extraction with Modes

```bash
# Extract specific files with flat structure
flint-vault extract -v vault.flint -o ./output/ --files report.pdf,data.csv

# Extract specific files preserving structure
flint-vault extract -v vault.flint -o ./output/ --files docs/report.pdf,project/ --extract-full-path
```

### Parallel Processing with Modes

```bash
# High-performance flat extraction
flint-vault extract -v large-vault.flint -o ./flat/ --workers 8

# High-performance structured extraction
flint-vault extract -v large-vault.flint -o ./structured/ --workers 8 --extract-full-path
```

### API Examples

```go
// Extract different file sets with different modes
func extractWithModes() {
    // Quick access files (flat)
    quickFiles := []string{"config.json", "readme.txt", "version.info"}
    err := vault.ExtractMultipleFilesFromVaultParallelWithOptions(
        vaultPath, password, "./quick/", quickFiles, config, false)
    
    // Project files (structured)  
    projectFiles := []string{"src/", "docs/", "tests/"}
    err = vault.ExtractMultipleFilesFromVaultParallelWithOptions(
        vaultPath, password, "./project/", projectFiles, config, true)
}
```

## 💡 Best Practices

### Workflow Recommendations

1. **Development workflow**:
   ```bash
   # Quick file access for editing
   flint-vault extract -v dev.flint -o ./temp/ --files main.go,config.yaml
   
   # Full project restore
   flint-vault extract -v dev.flint -o ./project/ --extract-full-path
   ```

2. **Backup workflow**:
   ```bash
   # Create structured backup
   flint-vault add -v backup.flint -s ./important-project/
   
   # Restore with structure
   flint-vault extract -v backup.flint -o ./restored/ --extract-full-path
   ```

3. **Analysis workflow**:
   ```bash
   # Extract for processing
   flint-vault extract -v data.flint -o ./analysis/
   
   # Process all files in flat structure
   python analyze.py ./analysis/*.csv
   ```

### Performance Considerations

- **Flat extraction**: Slightly faster due to simpler path operations
- **Full-path extraction**: May require creating directories, slightly slower
- **Both modes**: Support parallel processing for optimal performance

### File Name Conflicts

**In flat extraction**, if multiple files have the same name:
```
vault contains:
- docs/readme.txt
- src/readme.txt

flat extraction results in:
- readme.txt (last one overwrites)
```

**Solution**: Use full-path extraction when file names might conflict.

## 🔄 Migration Guide

### From Previous Versions

Previous behavior (always full-path) is now available with `--extract-full-path`:

```bash
# Old command (implicitly full-path)
flint-vault extract -v vault.flint -o ./output/

# New equivalent command (explicitly full-path)
flint-vault extract -v vault.flint -o ./output/ --extract-full-path

# New default behavior (flat)
flint-vault extract -v vault.flint -o ./output/
```

### API Migration

```go
// Old API (always full-path)
err := vault.ExtractFromVault(vaultPath, password, outputDir)

// New API - explicit control
err := vault.ExtractFromVaultWithOptions(vaultPath, password, outputDir, true)  // full-path
err := vault.ExtractFromVaultWithOptions(vaultPath, password, outputDir, false) // flat
```

---

**📋 Summary:**
- **Default behavior**: Flat extraction for quick file access
- **Optional flag**: `--extract-full-path` preserves directory structure  
- **API support**: Both modes available programmatically
- **Performance**: Both modes support parallel processing
- **Backward compatibility**: Previous behavior available with explicit flag 