# 📦 Installation Guide

This guide covers different ways to install and set up Flint Vault on your system.

## 📋 Prerequisites

### System Requirements

- **Operating System**: Linux, macOS, Windows
- **Go Version**: 1.21 or higher
- **Memory**: 
  - Minimum: 512MB RAM (basic operations)
  - Recommended: 8GB+ RAM (for large files >1GB)
  - For optimal performance: 16GB+ RAM
- **Disk Space**: 
  - Installation: 50MB
  - Operations: ~3.2x size of files being processed (temporary)

**📊 Performance Notes:**
- Tested with 2.45 GB datasets
- Memory usage scales at 3.2:1 ratio during encryption
- Supports files larger than available RAM through streaming

### Required Dependencies

- **Go compiler** (for building from source)
- **Git** (for cloning repository)
- **Terminal/Command Line** access

### Optional Dependencies

- **golangci-lint** (for development)
- **delve** (for debugging)

## 🚀 Installation Methods

### Method 1: Build from Source (Recommended)

#### Step 1: Install Go

If you don't have Go installed:

```bash
# Linux/macOS using official installer
wget https://go.dev/dl/go1.21.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Or use package manager
# Ubuntu/Debian
sudo apt update && sudo apt install golang-go

# macOS with Homebrew
brew install go

# Arch Linux
sudo pacman -S go

# Verify Go installation
go version
```

#### Step 2: Clone and Build

```bash
# Clone the repository
git clone https://github.com/balamutik/flint-vault.git
cd flint-vault

# Download dependencies
go mod download

# Build for your platform (development)
go build -o flint-vault ./cmd

# Build optimized for production
go build -ldflags="-w -s" -o flint-vault ./cmd

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o flint-vault-linux ./cmd
GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o flint-vault-macos ./cmd
GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o flint-vault.exe ./cmd

# Install globally (optional)
sudo cp flint-vault /usr/local/bin/
```

#### Step 3: Verify Build

```bash
# Test basic functionality
./flint-vault --help
./flint-vault info --help

# Run tests to ensure everything works
go test ./...

# Optional: Run stress tests (requires large files)
go test -tags=stress ./pkg/lib/vault
```

### Method 2: Go Install

If you have Go installed and configured:

```bash
# Install directly from Git
go install github.com/balamutik/flint-vault/cmd@latest

# The binary will be installed to $GOPATH/bin or $HOME/go/bin
# Make sure this directory is in your PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### Method 3: Cross-Platform Binaries

For users who prefer not to compile:

```bash
# Download pre-built binaries (when available)
# Linux
wget https://github.com/balamutik/flint-vault/releases/latest/download/flint-vault-linux
chmod +x flint-vault-linux
sudo mv flint-vault-linux /usr/local/bin/flint-vault

# macOS
wget https://github.com/balamutik/flint-vault/releases/latest/download/flint-vault-macos
chmod +x flint-vault-macos
sudo mv flint-vault-macos /usr/local/bin/flint-vault

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/balamutik/flint-vault/releases/latest/download/flint-vault.exe" -OutFile "flint-vault.exe"
```

## ✅ Verify Installation

After installation, verify that Flint Vault is working correctly:

```bash
# Check if command is available
flint-vault --help

# Test basic functionality
flint-vault create --help
flint-vault info --help

# Check version and features
flint-vault --version  # (if implemented)

# Quick functionality test
mkdir -p test_installation
cd test_installation
echo "Hello Flint Vault!" > test.txt
flint-vault create --file test.flint
flint-vault add --vault test.flint --source test.txt
flint-vault list --vault test.flint
flint-vault info --file test.flint
cd .. && rm -rf test_installation
```

**Expected Output:**
```
Creating encrypted vault: test.flint
✅ Vault successfully created!
🔐 Using AES-256-GCM encryption
...
📦 Vault: test.flint
📁 Contents (2 items):
...
✅ File Type: Flint Vault encrypted storage
```

## 🔧 Development Setup

For contributors and developers:

```bash
# Clone with all features
git clone https://github.com/balamutik/flint-vault.git
cd flint-vault

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest

# Download dependencies
go mod download

# Verify development setup
go test ./...
go test -race ./...
golangci-lint run

# Optional: Set up for large file testing
mkdir -p stress_test_data
# Create large test files if needed
dd if=/dev/urandom of=stress_test_data/test_100mb.bin bs=1M count=100
```

### IDE Configuration

#### VS Code
```bash
# Install Go extension
code --install-extension golang.Go

# Create workspace settings
mkdir -p .vscode
cat > .vscode/settings.json << EOF
{
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "go.coverOnSave": true,
    "go.testTimeout": "30s"
}
EOF
```

## 🆘 Troubleshooting

### Common Issues

#### "command not found: flint-vault"

**Solution**: Add the binary to your PATH or use full path to executable.

```bash
# Check current PATH
echo $PATH

# Find where Go installs binaries
go env GOPATH

# Add directory to PATH temporarily
export PATH=$PATH:$(go env GOPATH)/bin

# Add to PATH permanently (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

#### "permission denied"

**Solution**: Make sure the binary has execute permissions.

```bash
chmod +x flint-vault

# If using system-wide installation
sudo chmod +x /usr/local/bin/flint-vault
```

#### "cannot find module" or build errors

**Solution**: Ensure Go modules are properly configured.

```bash
# Clean and re-download modules
go clean -modcache
go mod download
go mod tidy

# Verify Go version
go version  # Should be 1.21+

# Try building again
go build -v ./cmd
```

#### Memory issues with large files

**Solution**: Check available system memory.

```bash
# Check available memory
free -h  # Linux
vm_stat | head -5  # macOS
Get-ComputerInfo | Select-Object TotalPhysicalMemory,AvailablePhysicalMemory  # Windows PowerShell

# For files >1GB, ensure you have sufficient RAM
# Rule of thumb: 3.2x file size in available memory
```

#### Performance issues

**Solution**: Check system resources and configuration.

```bash
# Check CPU info
lscpu  # Linux
sysctl -n machdep.cpu.brand_string  # macOS

# Check for hardware AES support (improves performance)
grep -m1 -o aes /proc/cpuinfo  # Linux - should output "aes"

# Monitor resource usage during operations
top
htop  # if available
```

### Getting Help

1. **Read Documentation**: Check all files in `docs/` directory
2. **Search Issues**: [GitHub Issues](https://github.com/balamutik/flint-vault/issues)
3. **Create Issue**: Include system info and error details
4. **Performance Issues**: Include memory usage and file sizes

### System Information for Bug Reports

Include this information when reporting issues:

```bash
# Gather comprehensive system info
echo "=== Flint Vault Info ==="
flint-vault --help | head -5
echo ""

echo "=== System Info ==="
go version
uname -a
echo "Shell: $SHELL"
echo "PATH: $PATH"
echo ""

echo "=== Memory Info ==="
free -h 2>/dev/null || vm_stat | head -5 2>/dev/null || echo "Memory info not available"
echo ""

echo "=== CPU Info ==="
lscpu | grep -E "(Model name|CPU\(s\)|Thread)" 2>/dev/null || \
sysctl -n machdep.cpu.brand_string 2>/dev/null || \
echo "CPU info not available"
echo ""

echo "=== Go Environment ==="
go env GOVERSION GOOS GOARCH GOROOT GOPATH
```

## 🔄 Updating

### Source Installation

```bash
cd flint-vault
git pull origin main
go mod download
go build -ldflags="-w -s" -o flint-vault ./cmd

# Update globally installed binary
sudo cp flint-vault /usr/local/bin/

# Verify update
flint-vault --help
```

### Go Install

```bash
go install github.com/balamutik/flint-vault/cmd@latest

# Verify update
flint-vault --help
```

### Check for Updates

```bash
# Check current version (if available)
flint-vault --version

# Check latest release on GitHub
curl -s https://api.github.com/repos/balamutik/flint-vault/releases/latest | grep tag_name
```

## 🗑️ Uninstallation

### Remove Binary

```bash
# Remove globally installed binary
sudo rm /usr/local/bin/flint-vault

# Remove Go-installed binary
rm $(go env GOPATH)/bin/flint-vault

# Remove source directory (if cloned)
rm -rf flint-vault
```

### Clean Up Go Cache

```bash
# Clean Go module cache (optional)
go clean -modcache

# Clean build cache
go clean -cache
```

## 🚀 Performance Optimization

### System Tuning for Large Files

```bash
# Increase file descriptor limits (Linux)
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Optimize for SSD (Linux)
sudo echo mq-deadline > /sys/block/sda/queue/scheduler

# Increase memory overcommit (if needed)
sudo sysctl vm.overcommit_memory=1
```

### Environment Variables

```bash
# Optional: Set Go-specific optimizations
export GOGC=100              # Garbage collection target
export GOMAXPROCS=$(nproc)    # Use all CPU cores

# For development
export GOPROXY=direct         # Bypass proxy for private repos
export GOSUMDB=off           # Disable checksum verification
```

---

**🚀 Installation Complete!**

**Next Steps:** 
- Read the [User Manual](USAGE.md) to learn basic operations
- Check [API Documentation](API.md) for programmatic usage  
- See [Security Guide](SECURITY.md) for best practices
- Review [Performance Guide](MEMORY_OPTIMIZATION.md) for large files

**📊 Tested with 2.45 GB datasets - Ready for production use!**

*Installation guide updated: June 2025* 