# 📖 User Manual

Complete guide to using Flint Vault for secure file storage and management.

## 🎯 Overview

Flint Vault provides secure, encrypted storage for your files and directories. All data is protected using military-grade AES-256-GCM encryption with strong password derivation.

**🚀 Recently Updated:**
- Unified architecture for better performance
- Memory-optimized operations for large files
- **Parallel processing** with configurable worker pools
- Enhanced compression support
- **Progress reporting** for long-running operations
- Stress-tested with multi-GB datasets

## 🔧 Command Structure

```bash
flint-vault <command> [options]
```

### Command Overview

| Command | Purpose | Key Features |
|---------|---------|--------------|
| `create` | Create new vault | AES-256-GCM encryption |
| `add` | Add files/directories | Recursive, compression |
| `list` | View vault contents | Fast metadata-only |
| `extract` | Extract all files | Full restore |
| `get` | Extract specific files | Selective extraction |
| `remove` | Remove files | Multiple targets |
| `info` | Vault information | Password-free |

## 📝 Commands

### 1. create - Create New Vault

Creates a new encrypted vault file with military-grade security.

```bash
flint-vault create --file <vault-file> [--password <password>]
```

**Options:**
- `-f, --file <path>`: Path for the new vault file
- `-p, --password <password>`: Password (prompted securely if not provided)

**Examples:**

```bash
# Create a new vault (password will be prompted securely)
flint-vault create --file my-documents.flint

# Create vault in specific directory
flint-vault create -f ~/backups/important-files.flint

# Create with password in command (NOT RECOMMENDED)
flint-vault create -f test.flint -p mypassword
```

**Output:**
```
Creating encrypted vault: my-documents.flint
✅ Vault successfully created!
🔐 Using AES-256-GCM encryption
🧂 Applied cryptographically secure salt
🔑 Key derived using PBKDF2 (100,000 iterations)
```

**Security Note:** 🔒 Always use the password prompt instead of `-p` flag to prevent password exposure in shell history.

### 2. add - Add Files and Directories

Adds files or directories to an existing vault with compression, optimization, and parallel processing.

```bash
flint-vault add --vault <vault-file> --source <source-path> [--password <password>] [--workers <num>] [--progress]
```

**Options:**
- `-v, --vault <path>`: Vault file path
- `-s, --source <path>`: File or directory to add
- `-p, --password <password>`: Password (prompted if not provided)
- `-w, --workers <num>`: Number of parallel workers (0 = auto-detect, default: 0)
- `--progress`: Show progress information (default: true)

**Examples:**

```bash
# Add a single file
flint-vault add --vault my-vault.flint --source ./document.pdf

# Add entire directory with parallel processing
flint-vault add -v my-vault.flint -s ./important-folder/ --workers 8

# Add with automatic worker detection and progress
flint-vault add -v my-vault.flint -s ./large-directory/

# Add with custom worker count
flint-vault add -v my-vault.flint -s ./project/ --workers 4 --progress

# Add without progress reporting
flint-vault add -v my-vault.flint -s ./quiet-operation/ --progress=false
```

**Performance Features:**
- **Parallel processing**: Configurable worker pools for large directories
- **Auto-detection**: Automatically determines optimal worker count (2x CPU cores)
- **Progress reporting**: Real-time status updates for long operations
- **Streaming I/O**: Memory-efficient for large files
- **Automatic compression**: Reduces vault size
- **Metadata preservation**: Timestamps, permissions
- **Batch optimization**: High-performance processing for multiple files

**Output Example:**
```
Adding directory 'large-folder/' to vault (workers: 8)...
🔄 Processing 245 files in batch mode...
🔄 Adding: folder/file001.dat
🔄 Adding: folder/file002.dat
...
✅ Successfully processed 245 files
📊 Total size: 1.2 GB, Compressed: 945 MB (21% savings)
📈 Performance: 85 MB/s average, Peak memory: 2.1 GB
⏱️ Duration: 14.2 seconds
```

### 3. list - View Vault Contents

Lists all files and directories stored in the vault with detailed metadata.

```bash
flint-vault list --vault <vault-file> [--password <password>]
```

**Options:**
- `-v, --vault <path>`: Vault file path
- `-p, --password <password>`: Password (prompted if not provided)

**Examples:**

```bash
# List vault contents
flint-vault list --vault my-vault.flint

# List with password
flint-vault list -v my-vault.flint -p mypassword
```

**Example Output:**
```
🔍 Vault: my-vault.flint
📁 Contents (5 items):

  📁 .  0 B  2025-06-19 22:37
  📄 documents/report.pdf  1.2 MB  2025-06-19 15:20
  📄 documents/data.xlsx  45.0 KB  2025-06-19 15:18
  📁 images/  0 B  2025-06-19 15:25
  📄 config.json  2.1 KB  2025-06-19 15:15
```

**Features:**
- **Fast operation**: Metadata-only, no decryption of file contents
- **File icons**: Visual distinction between files and directories
- **Size display**: Human-readable file sizes
- **Timestamps**: Last modification times preserved

### 4. extract - Extract All Files

Extracts all files from the vault to a destination directory with full restoration and parallel processing.

```bash
flint-vault extract --vault <vault-file> --output <destination> [--password <password>] [--files <list>] [--workers <num>] [--progress]
```

**Options:**
- `-v, --vault <path>`: Vault file path
- `-o, --output <path>`: Destination directory
- `-p, --password <password>`: Password (prompted if not provided)
- `-f, --files <list>`: Specific files to extract (optional, extracts all if not specified)
- `-w, --workers <num>`: Number of parallel workers (0 = auto-detect, default: 0)
- `--progress`: Show progress information (default: true)

**Examples:**

```bash
# Extract all files to current directory
flint-vault extract --vault my-vault.flint --output ./

# Extract with parallel processing
flint-vault extract -v my-vault.flint -o ~/restored-files/ --workers 6

# Extract specific files in parallel
flint-vault extract -v my-vault.flint -o ./output/ --files file1.txt,file2.pdf,folder/ --workers 4

# Extract with automatic optimization
flint-vault extract -v my-vault.flint -o ./backup/

# High-performance extraction for large vaults
flint-vault extract -v large-vault.flint -o ./data/ --workers 8 --progress
```

**Output Example:**
```
Extracting all files to directory: ./restored-files/
Using 6 parallel workers for extraction...
🔄 Extracting: documents/report.pdf (1.2 MB)
🔄 Extracting: images/photo.jpg (845 KB)
🔄 Extracting: data/dataset.csv (15.3 MB)
...
✅ Successfully extracted 187 files
📊 Total extracted: 2.1 GB in 8.3 seconds
📈 Performance: 253 MB/s average throughput
🔍 Integrity: 100% verified (all checksums match)
```

**Performance:**
- **High-speed extraction**: Up to 400+ MB/s with parallel processing
- **Memory efficient**: Streaming operations for large files
- **Full restoration**: Directory structure, permissions, timestamps
- **Selective extraction**: Extract only specified files for efficiency
- **Automatic optimization**: Uses optimal worker count based on file types

### 5. get - Extract Specific Files

Extracts specific files or directories from the vault. Supports multiple targets in single operation.

```bash
flint-vault get --vault <vault-file> --target <path> --output <destination> [--password <password>]
```

**Options:**
- `-v, --vault <path>`: Vault file path
- `-t, --target <path>`: File or directory path to extract
- `-o, --output <path>`: Destination directory
- `-p, --password <password>`: Password (prompted if not provided)

**Examples:**

```bash
# Extract single file
flint-vault get --vault my-vault.flint --target document.pdf --output ./

# Extract specific directory
flint-vault get -v my-vault.flint -t documents/ -o ./restored/

# Extract file from subdirectory
flint-vault get -v my-vault.flint -t documents/report.pdf -o ./

# Extract with full paths
flint-vault get -v backup.flint -t project/src/main.go -o ./restored/
```

**Output:**
```
Extracting 'documents/report.pdf' to directory: ./
✅ 'documents/report.pdf' successfully extracted to './'!
```

**Features:**
- **Selective extraction**: Only what you need
- **Path preservation**: Maintains directory structure
- **Fast operation**: Optimized for single-file extraction

### 6. remove - Remove Files from Vault

Removes specified files or directories from the vault with support for multiple targets.

```bash
flint-vault remove --vault <vault-file> --target <path> [--password <password>]
```

**Options:**
- `-v, --vault <path>`: Vault file path
- `-t, --target <path>`: File or directory path to remove
- `-p, --password <password>`: Password (prompted if not provided)

**Examples:**

```bash
# Remove single file
flint-vault remove --vault my-vault.flint --target old-document.pdf

# Remove entire directory
flint-vault remove -v my-vault.flint -t temp-folder/

# Remove file from subdirectory
flint-vault remove -v my-vault.flint -t documents/outdated.pdf
```

**Output:**
```
Removing 'old-document.pdf' from vault...
✅ 'old-document.pdf' successfully removed from vault!
```

**Performance:**
- **High-speed removal**: Up to 272 MB/s throughput
- **Efficient reorganization**: Vault optimization after removal
- **Memory efficient**: 2.5 GB peak memory for large operations

**Warning:** ⚠️ Removal is permanent and cannot be undone!

### 7. info - Vault Information

Displays vault file information without requiring password.

```bash
flint-vault info --file <vault-file>
```

**Options:**
- `-f, --file <path>`: Vault file path

**Examples:**

```bash
# Get vault information
flint-vault info --file my-vault.flint

# Check if file is valid vault
flint-vault info -f unknown-file.dat
```

**Example Output:**
```
🔍 Analyzing file: my-vault.flint

📁 File Path: my-vault.flint
📏 File Size: 2.4 GB
✅ File Type: Flint Vault encrypted storage
🔢 Format Version: 1
🔐 PBKDF2 Iterations: 100,000
✅ Validation: Passed

💡 This file can be opened with 'flint-vault list' command
```

**Features:**
- **Password-free**: No authentication required
- **Format validation**: Checks file integrity
- **Metadata display**: Version, iterations, size
- **Quick verification**: Instant format checking

## 🔐 Security Features

### Password Security

**Strong Passwords:**
- Use at least 12 characters
- Include uppercase, lowercase, numbers, and symbols
- Avoid dictionary words and personal information
- Consider using a password manager

**Example Strong Passwords:**
```
MySecur3_Vault#2025!
Tr0ub4dor&3_Flint
C0mpl3x_P@ssw0rd_2025
```

### Secure Password Entry

Flint Vault automatically hides password input:

```bash
# Password prompt (recommended)
$ flint-vault create --file secure.flint
Enter password for new vault: [hidden input]
✅ Vault successfully created!
```

### Advanced Security Features

- **AES-256-GCM**: Authenticated encryption preventing tampering
- **PBKDF2**: 100,000 iterations for password derivation
- **Secure random**: Cryptographically secure salt and nonce generation
- **Memory safety**: Sensitive data cleared after use
- **Format validation**: Magic headers prevent corruption

## 📊 Performance Guide

### Tested Performance (Real-World)

**System: Linux 6.14.8 (June 2025)**

| Operation | Speed | Workers | Memory Usage | Notes |
|-----------|-------|---------|--------------|-------|
| Vault Creation | <1s | - | 4 MB | Instant |
| Adding Files | 61-85 MB/s | Auto/8 | 3.2:1 ratio | Parallel optimized |
| Extracting Files | 245-400+ MB/s | Auto/4-8 | Minimal | 4-6x faster |
| Extracting Specific | 400+ MB/s | 4 | Low | Selective processing |
| Removing Files | 272 MB/s | - | 2.5:1 ratio | Fastest |
| Listing Contents | <1s | - | Minimal | Metadata only |

### Parallel Processing Optimization

**Worker Configuration Guidelines:**

```bash
# For CPU-intensive operations (compression-heavy)
--workers $(nproc)  # Equal to CPU cores

# For I/O-intensive operations (large files)
--workers $(($(nproc) * 2))  # 2x CPU cores (default auto-detect)

# For memory-constrained systems
--workers 2  # Conservative approach

# For maximum throughput (sufficient RAM)
--workers 8  # High-performance setup
```

**Performance Tips:**
- **Auto-detection**: Let Flint Vault choose optimal workers (recommended)
- **Large directories**: Use 4-8 workers for best performance
- **Small files**: 2-4 workers prevent overhead
- **Large files**: Auto-detection works best
- **Progress reporting**: Minimal performance impact

**Resource Requirements:**
```bash
# Memory calculation for parallel operations
# Base memory: 3.2x data size
# Per worker: +100-200 MB overhead
# Example: 1GB vault with 4 workers = ~3.5-4GB RAM

# Check available resources
free -h  # Check memory
nproc    # Check CPU cores
```

### Large File Handling

**Successfully Tested:**
- **2.45 GB datasets**: Multiple files 400MB-800MB each
- **Memory efficiency**: 3.2:1 memory-to-data ratio
- **100% data integrity**: All files preserved perfectly

**Performance Tips:**
```bash
# For very large files (>1GB each)
# Memory requirements: ~3.2x file size during encryption
# Recommended: 8GB+ RAM for files over 2GB

# Monitor resource usage
top -p $(pgrep flint-vault)
```

## 📁 Working with Different File Types

### Text Files and Documents

```bash
# Add configuration files
flint-vault add -v config.flint -s .bashrc
flint-vault add -v config.flint -s .vimrc
flint-vault add -v config.flint -s config/

# Add office documents
flint-vault add -v docs.flint -s reports/
flint-vault add -v docs.flint -s presentations/
```

### Binary Files and Media

```bash
# Add images and media (large files supported)
flint-vault add -v media.flint -s ./photos/
flint-vault add -v media.flint -s ./videos/

# Add applications and executables
flint-vault add -v apps.flint -s ./bin/
flint-vault add -v apps.flint -s ./tools/
```

### Source Code and Projects

```bash
# Add entire project
flint-vault add -v project.flint -s ./my-project/

# Selective project backup
flint-vault add -v project.flint -s ./src/
flint-vault add -v project.flint -s ./docs/
flint-vault add -v project.flint -s README.md
flint-vault add -v project.flint -s package.json
```

## 🌍 International Support

### Unicode File Names

Flint Vault fully supports international characters:

```bash
# Russian files
flint-vault add -v docs.flint -s ./документы/

# Chinese files  
flint-vault add -v files.flint -s ./文档/

# Japanese files
flint-vault add -v data.flint -s ./データ/

# Emoji in names (supported!)
flint-vault add -v fun.flint -s ./🚀_projects/
flint-vault add -v fun.flint -s ./📄_documents/
```

### Special Characters

Supported characters in file names:
- **Spaces**: `my document.txt`
- **Dashes**: `file-name.txt`
- **Underscores**: `file_name.txt`
- **Dots**: `file.backup.txt`
- **Unicode**: `文档.txt`, `файл.doc`, `ファイル.txt`

## 💡 Best Practices

### Vault Organization

```bash
# Separate vaults by purpose
flint-vault create -f documents.flint    # For documents
flint-vault create -f media.flint        # For photos/videos
flint-vault create -f backup.flint       # For system backups

# Use descriptive names with dates
flint-vault create -f work-2025.flint
flint-vault create -f personal-docs-june.flint
```

### Efficient Workflows

```bash
# Regular document backup
flint-vault add -v daily-docs.flint -s ~/Documents/
flint-vault add -v daily-docs.flint -s ~/Desktop/

# Project versioning
flint-vault create -f project-v1.0.flint
flint-vault add -v project-v1.0.flint -s ./project/

# Selective extraction
flint-vault get -v archive.flint -t needed-file.pdf -o ./
```

### Security Hygiene

1. **Never share passwords** via insecure channels
2. **Use different passwords** for different vaults
3. **Keep vault files secure** - they contain encrypted data
4. **Regular password changes** for sensitive vaults
5. **Test extraction** periodically to ensure data integrity
6. **Backup vault files** themselves to multiple locations

## 🔧 Advanced Usage

### Batch Operations

```bash
# Create multiple themed vaults
for category in docs media code configs; do
    flint-vault create -f ${category}-$(date +%Y%m%d).flint
done

# Batch file addition
find ./important -name "*.pdf" -exec flint-vault add -v docs.flint -s {} \;
```

### Scripting with Flint Vault

```bash
#!/bin/bash
# automated-backup.sh

DATE=$(date +%Y%m%d)
VAULT_FILE="daily-backup-${DATE}.flint"
SOURCE_DIR="/home/user/important"

echo "🔐 Creating vault: $VAULT_FILE"
flint-vault create -f "$VAULT_FILE"

echo "📁 Adding files from: $SOURCE_DIR"
flint-vault add -v "$VAULT_FILE" -s "$SOURCE_DIR"

echo "📊 Vault information:"
flint-vault info -f "$VAULT_FILE"

echo "✅ Backup completed: $VAULT_FILE"
```

### Environment-Based Usage

```bash
# Set vault location
export VAULT_DIR="$HOME/.vaults"
mkdir -p "$VAULT_DIR"

# Create vault with environment
flint-vault create -f "$VAULT_DIR/personal.flint"

# Conditional operations
if flint-vault info -f "$VAULT_DIR/backup.flint" >/dev/null 2>&1; then
    echo "Vault exists, adding files..."
    flint-vault add -v "$VAULT_DIR/backup.flint" -s "$HOME/new-files/"
else
    echo "Creating new vault..."
    flint-vault create -f "$VAULT_DIR/backup.flint"
fi
```

## 🚨 Troubleshooting

### Common Error Messages

#### "Authentication failed" / "Invalid password"

```bash
# Error message:
Error: authentication failed: invalid password or corrupted vault

# Solutions:
1. Double-check password spelling
2. Verify vault file isn't corrupted
3. Try with backup vault file
4. Use 'info' command to verify vault integrity
```

#### "File not found in vault"

```bash
# Error message:
Error: file or directory 'missing.txt' not found in vault

# Solutions:
1. Use 'list' command to see available files
flint-vault list -v my-vault.flint

2. Check exact file path and name (case-sensitive)
3. Verify file was actually added to vault
```

#### "Insufficient memory" / Out of Memory

```bash
# For very large files on memory-constrained systems:

# Solutions:
1. Ensure sufficient RAM (3.2x file size)
2. Process files individually instead of directories
3. Use system with more memory for large operations
4. Split large files before adding
```

### Performance Optimization

#### For Large Files

```bash
# Best practices for files >1GB:
1. Ensure sufficient RAM (recommended: 8GB+ for 2GB+ files)
2. Use SSD storage for better I/O performance
3. Close other applications to free memory
4. Process large files individually

# Monitor during operation:
watch -n 1 'free -h; echo "---"; ps aux | grep flint-vault'
```

#### Memory Management

```bash
# Check available memory before large operations
free -h

# Set memory limits if needed (for scripting)
ulimit -v 4194304  # Limit to 4GB virtual memory

# Monitor vault operations
top -p $(pgrep flint-vault)
```

## 📊 Examples and Scenarios

### Personal Document Management

```bash
# Create personal document vault
flint-vault create -f personal-docs.flint

# Add important documents
flint-vault add -v personal-docs.flint -s ~/Documents/passport.pdf
flint-vault add -v personal-docs.flint -s ~/Documents/tax-returns/
flint-vault add -v personal-docs.flint -s ~/Documents/certificates/

# Check contents
flint-vault list -v personal-docs.flint

# Extract specific document when needed
flint-vault get -v personal-docs.flint -t passport.pdf -o ./
```

### Development Project Backup

```bash
# Create comprehensive project backup
flint-vault create -f project-backup-v2.flint

# Add source code and documentation
flint-vault add -v project-backup-v2.flint -s ./src/
flint-vault add -v project-backup-v2.flint -s ./docs/
flint-vault add -v project-backup-v2.flint -s ./tests/
flint-vault add -v project-backup-v2.flint -s README.md
flint-vault add -v project-backup-v2.flint -s package.json

# Verify backup contents
flint-vault list -v project-backup-v2.flint

# Restore on new machine
flint-vault extract -v project-backup-v2.flint -o ./restored-project/
```

### System Configuration Backup

```bash
# Create system configuration vault
flint-vault create -f system-config.flint

# Add various configuration files
flint-vault add -v system-config.flint -s ~/.bashrc
flint-vault add -v system-config.flint -s ~/.vimrc
flint-vault add -v system-config.flint -s ~/.ssh/config
flint-vault add -v system-config.flint -s ~/.gitconfig

# Add configuration directories
flint-vault add -v system-config.flint -s ~/.config/
flint-vault add -v system-config.flint -s /etc/nginx/

# Check vault info
flint-vault info -f system-config.flint
```

### Media Archive Management

```bash
# Create media archive (large files)
flint-vault create -f media-archive-2025.flint

# Add photo collections (handles large directories efficiently)
flint-vault add -v media-archive-2025.flint -s ~/Pictures/2025/
flint-vault add -v media-archive-2025.flint -s ~/Pictures/vacation/

# Add video files (streaming handles large files)
flint-vault add -v media-archive-2025.flint -s ~/Videos/important/

# Extract specific albums
flint-vault get -v media-archive-2025.flint -t "Pictures/vacation/" -o ./restored/
```

## 🎯 Migration and Upgrade

### From Other Tools

```bash
# Extract from other encrypted storage
tar -xzf old-archive.tar.gz
flint-vault create -f migrated-data.flint
flint-vault add -v migrated-data.flint -s ./extracted-data/

# Verify migration
flint-vault list -v migrated-data.flint
```

### Vault Maintenance

```bash
# Regular maintenance routine
flint-vault info -f important.flint  # Check integrity
flint-vault list -v important.flint  # Verify contents

# Create backup copy
cp important.flint important-backup-$(date +%Y%m%d).flint

# Test extraction (to temporary location)
mkdir -p /tmp/vault-test
flint-vault extract -v important.flint -o /tmp/vault-test/
rm -rf /tmp/vault-test
```

---

**📚 This manual covers all features of Flint Vault unified architecture**  
**🚀 Tested with multi-GB datasets - ready for production use**  
**🔒 Security-first design with proven cryptographic algorithms**

*Last updated: June 2025* 