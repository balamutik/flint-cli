# 📖 User Manual

Complete guide to using Flint Vault for secure file storage and management.

## 🎯 Overview

Flint Vault provides secure, encrypted storage for your files and directories. All data is protected using military-grade AES-256-GCM encryption with strong password derivation.

## 🔧 Command Structure

```bash
flint-vault <command> [options]
```

### Global Options

- `-f, --file <path>`: Vault file path
- `-p, --password <password>`: Password (not recommended for security)
- `-h, --help`: Show help information
- `--version`: Show version information

## 📝 Commands

### 1. create - Create New Vault

Creates a new encrypted vault file.

```bash
flint-vault create -f <vault-file>
```

**Options:**
- `-f, --file <path>`: Path for the new vault file
- `-p, --password <password>`: Password (prompted securely if not provided)

**Examples:**

```bash
# Create a new vault (password will be prompted securely)
flint-vault create -f my-documents.vault

# Create vault in specific directory
flint-vault create -f ~/backups/important-files.vault

# Create with password in command (NOT RECOMMENDED)
flint-vault create -f test.vault -p mypassword
```

**Security Note:** 🔒 Always use the password prompt instead of `-p` flag to prevent password exposure in shell history.

### 2. add - Add Files and Directories

Adds files or directories to an existing vault.

```bash
flint-vault add -f <vault-file> <source-path>
```

**Options:**
- `-f, --file <path>`: Vault file path
- `-p, --password <password>`: Password (prompted if not provided)

**Examples:**

```bash
# Add a single file
flint-vault add -f my-vault.dat ./document.pdf

# Add entire directory recursively
flint-vault add -f my-vault.dat ./important-folder/

# Add multiple files
flint-vault add -f my-vault.dat file1.txt file2.txt folder/

# Add with absolute paths
flint-vault add -f my-vault.dat /home/user/documents/
```

**Behavior:**
- Files are added with their full paths preserved
- Directories are added recursively
- Existing files are updated with new content
- File permissions and timestamps are preserved

### 3. list - View Vault Contents

Lists all files and directories stored in the vault.

```bash
flint-vault list -f <vault-file>
```

**Options:**
- `-f, --file <path>`: Vault file path
- `-p, --password <password>`: Password (prompted if not provided)

**Examples:**

```bash
# List vault contents
flint-vault list -f my-vault.dat

# Example output:
# Vault Contents:
# Created: 2023-12-01 15:30:45
# Comment: Encrypted Flint Vault Storage
# 
# Files and Directories:
# drwxr-xr-x  0 B     2023-12-01 15:25  documents/
# -rw-r--r--  1.2 MB  2023-12-01 15:20  documents/report.pdf
# -rw-r--r--  45 KB   2023-12-01 15:18  documents/data.xlsx
# -rw-r--r--  2.1 KB  2023-12-01 15:15  config.json
# 
# Total: 4 items, 1.2 MB compressed
```

### 4. extract - Extract All Files

Extracts all files from the vault to a destination directory.

```bash
flint-vault extract -f <vault-file> -d <destination>
```

**Options:**
- `-f, --file <path>`: Vault file path
- `-d, --destination <path>`: Destination directory
- `-p, --password <password>`: Password (prompted if not provided)

**Examples:**

```bash
# Extract all files to current directory
flint-vault extract -f my-vault.dat -d ./

# Extract to specific directory
flint-vault extract -f my-vault.dat -d ~/restored-files/

# Extract with password
flint-vault extract -f my-vault.dat -d ./backup/ -p mypassword
```

**Behavior:**
- Creates destination directory if it doesn't exist
- Preserves original file structure
- Restores file permissions and timestamps
- Overwrites existing files

### 5. get - Extract Specific Files

Extracts specific files or directories from the vault. Supports extracting multiple files and directories in a single command.

```bash
flint-vault get -v <vault-file> -t <target1> -t <target2> ... -o <destination>
```

**Options:**
- `-v, --vault <path>`: Vault file path
- `-t, --targets <path>`: File or directory paths to extract (can be specified multiple times)
- `-o, --output <path>`: Destination directory
- `-p, --password <password>`: Password (prompted if not provided)

**Examples:**

```bash
# Extract single file
flint-vault get -v my-vault.dat -t document.pdf -o ./

# Extract specific directory
flint-vault get -v my-vault.dat -t documents/ -o ./restored/

# Extract multiple files at once
flint-vault get -v my-vault.dat -t document1.pdf -t document2.pdf -t config.json -o ./

# Extract multiple directories and files
flint-vault get -v my-vault.dat -t documents/ -t images/ -t readme.txt -o ./restored/

# Extract file from subdirectory
flint-vault get -v my-vault.dat -t documents/report.pdf -o ./

# Extract mixed content with detailed feedback
flint-vault get -v my-vault.dat -t src/ -t README.md -t config/ -t nonexistent.txt -o ./backup/
# Output shows:
# ✅ Successfully extracted 3 items:
#   ✓ src/
#   ✓ README.md  
#   ✓ config/
# ⚠️  1 items not found in vault:
#   ✗ nonexistent.txt
```

**Behavior:**
- **Single target**: Uses optimized single-file extraction
- **Multiple targets**: Processes all targets and provides detailed feedback
- **Not found items**: Lists missing files/directories but continues extracting found items
- **Overlapping paths**: Automatically handles duplicate entries (e.g., extracting both `dir/` and `dir/file.txt`)
- **Directory structure**: Preserves original file organization and permissions

### 6. remove - Remove Files from Vault

Removes files or directories from the vault.

```bash
flint-vault remove -f <vault-file> -n <file-name>
```

**Options:**
- `-f, --file <path>`: Vault file path
- `-n, --name <path>`: File or directory name to remove
- `-p, --password <password>`: Password (prompted if not provided)

**Examples:**

```bash
# Remove single file
flint-vault remove -f my-vault.dat -n old-document.pdf

# Remove entire directory
flint-vault remove -f my-vault.dat -n temp-folder/

# Remove with confirmation
flint-vault remove -f my-vault.dat -n important.doc
```

**Warning:** ⚠️ Removal is permanent and cannot be undone!

## 🔐 Security Features

### Password Security

**Strong Passwords:**
- Use at least 12 characters
- Include uppercase, lowercase, numbers, and symbols
- Avoid dictionary words and personal information
- Consider using a password manager

**Example Strong Passwords:**
```
MySecur3_Vault#2023!
Tr0ub4dor&3_Flint
C0mpl3x_P@ssw0rd_2023
```

### Secure Password Entry

Flint Vault automatically hides password input:

```bash
# Password prompt (recommended)
$ flint-vault create -f secure.vault
Enter password: [hidden input]
Confirm password: [hidden input]
Vault created successfully!
```

### File Path Security

All file paths are stored with full information:
- Original file permissions
- Timestamps (creation, modification)
- Directory structure
- Unicode file names supported

## 📁 Working with Different File Types

### Text Files

```bash
# Add configuration files
flint-vault add -f config.vault .bashrc .vimrc config.json

# Add source code
flint-vault add -f source.vault ./src/
```

### Binary Files

```bash
# Add images and media
flint-vault add -f media.vault ./photos/ ./videos/

# Add executables
flint-vault add -f apps.vault ./bin/ ./tools/
```

### Archives and Backups

```bash
# Add compressed archives
flint-vault add -f backup.vault backup.tar.gz database.sql.gz

# Full system backup
flint-vault add -f system.vault /etc/ /home/user/
```

## 🌍 International Support

### Unicode File Names

Flint Vault fully supports international characters:

```bash
# Russian files
flint-vault add -f docs.vault ./документы/

# Chinese files  
flint-vault add -f files.vault ./文档/

# Emoji in names
flint-vault add -f fun.vault ./🚀_projects/ ./📄_documents/
```

### Special Characters

Supported characters in file names:
- Spaces: `my document.txt`
- Dashes: `file-name.txt`
- Underscores: `file_name.txt`
- Dots: `file.backup.txt`
- Unicode: `文档.txt`, `файл.doc`

## 💡 Best Practices

### Vault Organization

```bash
# Separate vaults by purpose
flint-vault create -f documents.vault    # For documents
flint-vault create -f media.vault        # For photos/videos
flint-vault create -f backup.vault       # For system backups

# Use descriptive names
flint-vault create -f work-2023.vault
flint-vault create -f personal-docs.vault
```

### Regular Backups

```bash
# Backup vault files themselves
cp important.vault important.vault.backup

# Create multiple copies
flint-vault extract -f main.vault -d ./backup-$(date +%Y%m%d)/
```

### Security Hygiene

1. **Never share passwords** via insecure channels
2. **Use different passwords** for different vaults
3. **Keep vault files secure** - they contain encrypted data
4. **Regular password changes** for sensitive vaults
5. **Test extraction** periodically to ensure data integrity

## 🔧 Advanced Usage

### Batch Operations

```bash
# Create multiple vaults
for category in docs media code; do
    flint-vault create -f ${category}.vault
done

# Add files in batches
find . -name "*.pdf" -exec flint-vault add -f docs.vault {} \;
```

### Scripting with Flint Vault

```bash
#!/bin/bash
# backup-script.sh

VAULT_FILE="daily-backup-$(date +%Y%m%d).vault"
SOURCE_DIR="/home/user/important"

# Create vault
flint-vault create -f "$VAULT_FILE"

# Add files
flint-vault add -f "$VAULT_FILE" "$SOURCE_DIR"

echo "Backup completed: $VAULT_FILE"
```

### Environment Variables

```bash
# Set default vault
export FLINT_VAULT_DEFAULT_FILE="$HOME/.vault/default.dat"

# Use in commands
flint-vault list  # Uses default vault file
```

## 🚨 Troubleshooting

### Common Error Messages

#### "Invalid password or corrupted data"

```bash
# Error message:
Error: decryption failed: invalid password or corrupted data

# Solutions:
1. Check password spelling
2. Verify vault file isn't corrupted
3. Try with original vault file backup
```

#### "Vault file already exists"

```bash
# Error message:
Error: vault file already exists: my-vault.dat

# Solutions:
1. Use different file name
2. Remove existing file if intended
3. Use 'add' command to modify existing vault
```

#### "File not found in vault"

```bash
# Error message:
Error: file or directory 'missing.txt' not found in vault

# Solutions:
1. Use 'list' command to see available files
2. Check exact file path and name
3. Verify file was actually added to vault
```

### Performance Tips

#### Large Files

```bash
# For very large files, consider:
1. Splitting large files before adding
2. Using compression before vault storage
3. Processing in smaller batches
```

#### Memory Usage

```bash
# Monitor memory usage
top -p $(pgrep flint-vault)

# For memory-constrained systems
ulimit -m 512000  # Limit memory to 512MB
```

## 📊 Examples and Scenarios

### Personal Document Management

```bash
# Create personal vault
flint-vault create -f personal.vault

# Add important documents
flint-vault add -f personal.vault ~/Documents/passport.pdf
flint-vault add -f personal.vault ~/Documents/tax-returns/
flint-vault add -f personal.vault ~/Documents/certificates/

# List contents
flint-vault list -f personal.vault

# Extract when needed
flint-vault get -f personal.vault -n passport.pdf -d ./
```

### Development Project Backup

```bash
# Create project vault
flint-vault create -f project-backup.vault

# Add source code (excluding build artifacts)
flint-vault add -f project-backup.vault ./src/
flint-vault add -f project-backup.vault ./docs/
flint-vault add -f project-backup.vault README.md LICENSE

# Restore on new machine
flint-vault extract -f project-backup.vault -d ./restored-project/
```

### System Configuration Backup

```bash
# Create system config vault
flint-vault create -f system-config.vault

# Add configuration files
flint-vault add -f system-config.vault ~/.bashrc
flint-vault add -f system-config.vault ~/.vimrc
flint-vault add -f system-config.vault ~/.gitconfig
flint-vault add -f system-config.vault ~/.ssh/

# Restore configurations
flint-vault extract -f system-config.vault -d ~/
```

---

**Next**: Learn about [API Documentation](API.md) for programmatic usage. 