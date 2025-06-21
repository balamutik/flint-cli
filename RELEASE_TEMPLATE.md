# Release Template for GitHub

Use this template when creating releases manually on GitHub.

## Release Title Format
```
Flint Vault v1.0.0
```

## Release Description Template

```markdown
# Flint Vault v1.0.0

Secure encrypted file storage with military-grade AES-256-GCM encryption.

## 🚀 What's New
- [Add new features here]
- [Add improvements here]
- [Add bug fixes here]

## 🛡️ Security Features
- **AES-256-GCM encryption** with PBKDF2 key derivation (100,000 iterations)
- **SHA-256 integrity verification** for all files
- **Cryptographically secure** salt and nonce generation
- **Memory safety** with sensitive data clearing

## ⚡ Performance Features
- **High-performance streaming operations** for large files (tested up to 800MB+)
- **Parallel processing** for multiple files and directories
- **Memory-efficient operations** (8-10MB RAM usage for 800MB files)
- **Gzip compression** to reduce vault size
- **Optimized extraction** (260+ MB/s throughput)

## 📦 Installation

### Quick Install (Linux/macOS)
```bash
# Download and extract (replace with your platform)
wget https://github.com/yourusername/flint-vault-cli/releases/download/v1.0.0/flint-vault-v1.0.0-linux-amd64.tar.gz
tar -xzf flint-vault-v1.0.0-linux-amd64.tar.gz
cd flint-vault-v1.0.0-linux-amd64

# Install
sudo ./install.sh
```

### Manual Installation
1. Download the appropriate archive for your platform below
2. Extract the archive
3. Copy the binary to your PATH (e.g., `/usr/local/bin`)
4. Run `flint-vault --help` to get started

### Package Managers
```bash
# Homebrew (macOS/Linux) - Coming soon
brew install flint-vault

# Arch Linux AUR - Coming soon
yay -S flint-vault

# Snap - Coming soon
sudo snap install flint-vault
```

## 🔐 Quick Start
```bash
# Create a new vault
flint-vault create -f my-vault.flint

# Add files to vault
flint-vault add -v my-vault.flint -s important-file.txt

# Add entire directory with parallel processing
flint-vault add -v my-vault.flint -s my-folder/ --workers 4

# List vault contents
flint-vault list -v my-vault.flint

# Extract all files
flint-vault extract -v my-vault.flint -o extracted/

# Extract specific files with parallel processing
flint-vault extract -v my-vault.flint -o extracted/ --workers 4 file1.txt file2.txt

# Remove files from vault
flint-vault remove -v my-vault.flint -t unwanted-file.txt

# Get vault information (no password required)
flint-vault info -f my-vault.flint
```

## 📋 Platform Support
- **Linux**: x86_64, ARM64
- **macOS**: Intel (x86_64), Apple Silicon (ARM64)  
- **Windows**: x86_64

## 🧪 Tested Performance
- **800MB file processing**: 28.7 MB/s write, 264 MB/s read
- **Memory usage**: 8-10MB RAM for large file operations
- **Parallel processing**: Up to 12 concurrent operations
- **Compression**: Up to 99% reduction for repetitive data

## ✅ Verification
All binaries include SHA-256 checksums in `checksums.txt` for verification:

```bash
# Verify download (Linux/macOS)
sha256sum -c checksums.txt

# Verify download (Windows)
Get-FileHash flint-vault-v1.0.0-windows-amd64.zip
```

## 🔧 Technical Details
- **Go Version**: 1.21+
- **Dependencies**: Minimal (only stdlib + golang.org/x/crypto, golang.org/x/term)
- **Binary Size**: ~6-8MB (statically linked)
- **Test Coverage**: 77.5%
- **Memory Safety**: Extensive testing for race conditions and memory leaks

## 🐛 Known Issues
- [List any known issues]

## 📝 Full Changelog
- [Link to CHANGELOG.md or list changes]

---

**Download the appropriate binary for your platform below ⬇️**
```

## Files to Upload
When creating the release, upload these files from the `dist/` directory:
- `flint-vault-v1.0.0-linux-amd64.tar.gz`
- `flint-vault-v1.0.0-linux-arm64.tar.gz`
- `flint-vault-v1.0.0-darwin-amd64.tar.gz`
- `flint-vault-v1.0.0-darwin-arm64.tar.gz`
- `flint-vault-v1.0.0-windows-amd64.zip`
- `checksums.txt`

## Pre-release Checklist
- [ ] All tests pass (`go test ./...`)
- [ ] Build script runs successfully (`./build.sh v1.0.0`)
- [ ] Version information is correct
- [ ] README.md is updated
- [ ] CHANGELOG.md is updated (if exists)
- [ ] Security review completed
- [ ] Performance benchmarks run 