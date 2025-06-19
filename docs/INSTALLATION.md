# 📦 Installation Guide

This guide covers different ways to install and set up Flint Vault on your system.

## 📋 Prerequisites

### System Requirements

- **Operating System**: Linux, macOS, Windows
- **Go Version**: 1.21 or higher
- **Memory**: Minimum 512MB RAM
- **Disk Space**: 50MB for installation

### Required Dependencies

- **Go compiler** (for building from source)
- **Git** (for cloning repository)
- **Terminal/Command Line** access

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
```

#### Step 2: Clone and Build

```bash
# Clone the repository
git clone https://github.com/balamutik/flint-vault.git
cd flint-vault

# Download dependencies
go mod download

# Build for your platform
go build -o flint-vault ./cmd

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -o flint-vault-linux ./cmd
GOOS=darwin GOARCH=amd64 go build -o flint-vault-macos ./cmd
GOOS=windows GOARCH=amd64 go build -o flint-vault.exe ./cmd

# Install globally (optional)
sudo cp flint-vault /usr/local/bin/
```

### Method 2: Go Install

If you have Go installed and configured:

```bash
# Install directly from Git
go install github.com/balamutik/flint-vault/cmd@latest

# The binary will be installed to $GOPATH/bin or $HOME/go/bin
```

## ✅ Verify Installation

After installation, verify that Flint Vault is working correctly:

```bash
# Check if command is available
flint-vault --help

# Test basic functionality
flint-vault create --help
```

## 🔧 Development Setup

For contributors and developers:

```bash
# Clone with all submodules
git clone --recursive https://github.com/balamutik/flint-vault.git
cd flint-vault

# Install development dependencies
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests to verify setup
go test ./...
```

## 🆘 Troubleshooting

### Common Issues

#### "command not found: flint-vault"

**Solution**: Add the binary to your PATH or use full path to executable.

```bash
# Check current PATH
echo $PATH

# Add directory to PATH temporarily
export PATH=$PATH:/path/to/flint-vault

# Add to PATH permanently (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
```

#### "permission denied"

**Solution**: Make sure the binary has execute permissions.

```bash
chmod +x flint-vault
```

#### "cannot find module"

**Solution**: Download Go modules.

```bash
go mod download
go mod tidy
```

### Getting Help

1. **Check the FAQ**: [docs/FAQ.md](FAQ.md)
2. **Search existing issues**: [GitHub Issues](https://github.com/balamutik/flint-vault/issues)
3. **Create new issue**: Provide system info and error details
4. **Join discussions**: [GitHub Discussions](https://github.com/balamutik/flint-vault/discussions)

### System Information

Include this information when reporting issues:

```bash
# Gather system info
flint-vault --help
go version
uname -a
echo $SHELL
echo $PATH
```

## 🔄 Updating

### Source Installation

```bash
cd flint-vault
git pull origin main
go build -o flint-vault ./cmd
sudo cp flint-vault /usr/local/bin/
```

### Go Install

```bash
go install github.com/balamutik/flint-vault/cmd@latest
```

## 🗑️ Uninstallation

### Binary Installation

```bash
# Remove binary
sudo rm /usr/local/bin/flint-vault
```

---

**Next Steps**: After installation, read the [User Manual](USAGE.md) to learn how to use Flint Vault. 