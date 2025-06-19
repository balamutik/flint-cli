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

### Method 1: Binary Download (Recommended)

Download pre-built binaries from the releases page:

```bash
# Linux (x64)
wget https://github.com/yourusername/flint-vault/releases/latest/download/flint-vault-linux-amd64
chmod +x flint-vault-linux-amd64
sudo mv flint-vault-linux-amd64 /usr/local/bin/flint-vault

# macOS (Intel)
wget https://github.com/yourusername/flint-vault/releases/latest/download/flint-vault-darwin-amd64
chmod +x flint-vault-darwin-amd64
sudo mv flint-vault-darwin-amd64 /usr/local/bin/flint-vault

# macOS (Apple Silicon)
wget https://github.com/yourusername/flint-vault/releases/latest/download/flint-vault-darwin-arm64
chmod +x flint-vault-darwin-arm64
sudo mv flint-vault-darwin-arm64 /usr/local/bin/flint-vault

# Windows (x64)
# Download flint-vault-windows-amd64.exe and add to PATH
```

### Method 2: Build from Source

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
git clone https://github.com/yourusername/flint-vault.git
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

### Method 3: Go Install

If you have Go installed and configured:

```bash
# Install directly from Git
go install github.com/yourusername/flint-vault/cmd@latest

# The binary will be installed to $GOPATH/bin or $HOME/go/bin
```

### Method 4: Package Managers

#### Homebrew (macOS/Linux)

```bash
# Add tap (if available)
brew tap yourusername/flint-vault
brew install flint-vault
```

#### Snap (Linux)

```bash
# Install from Snap Store
sudo snap install flint-vault
```

#### AUR (Arch Linux)

```bash
# Using yay
yay -S flint-vault

# Using paru  
paru -S flint-vault
```

## ✅ Verify Installation

After installation, verify that Flint Vault is working correctly:

```bash
# Check if command is available
flint-vault --version

# Expected output:
# Flint Vault v1.0.0 (built with Go 1.21.4)

# Test basic functionality
flint-vault --help
```

## 🔧 Configuration

### Environment Variables

Flint Vault can be configured using environment variables:

```bash
# Set default vault file location
export FLINT_VAULT_DEFAULT_FILE="$HOME/.vault/default.dat"

# Set default output directory
export FLINT_VAULT_DEFAULT_OUTPUT="$HOME/vault-extracts"

# Enable debug mode
export FLINT_VAULT_DEBUG=1
```

### Configuration File

Create a configuration file at `~/.config/flint-vault/config.yaml`:

```yaml
# Default vault file
default_vault: "~/.vault/default.dat"

# Default extraction directory
default_output: "~/vault-extracts"

# Security settings
security:
  pbkdf2_iterations: 100000
  compression_level: 6

# UI settings
ui:
  show_progress: true
  color_output: true
```

## 🐳 Docker Installation

### Using Official Image

```bash
# Pull the official image
docker pull flint-vault/flint-vault:latest

# Create alias for easy usage
echo 'alias flint-vault="docker run --rm -v $(pwd):/workspace flint-vault/flint-vault:latest"' >> ~/.bashrc
source ~/.bashrc

# Test installation
flint-vault --version
```

### Build Custom Image

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o flint-vault ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/flint-vault .
ENTRYPOINT ["./flint-vault"]
```

```bash
# Build and run
docker build -t flint-vault .
docker run --rm -v $(pwd):/workspace flint-vault --help
```

## 🔧 Development Setup

For contributors and developers:

```bash
# Clone with all submodules
git clone --recursive https://github.com/yourusername/flint-vault.git
cd flint-vault

# Install development dependencies
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securecodewarrior/semmle-code@latest

# Set up git hooks
cp scripts/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit

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

#### Memory issues on large files

**Solution**: Increase available memory or process files in smaller batches.

```bash
# For very large vaults, use streaming mode
export FLINT_VAULT_STREAMING=1
```

### Getting Help

1. **Check the FAQ**: [docs/FAQ.md](FAQ.md)
2. **Search existing issues**: [GitHub Issues](https://github.com/yourusername/flint-vault/issues)
3. **Create new issue**: Provide system info and error details
4. **Join discussions**: [GitHub Discussions](https://github.com/yourusername/flint-vault/discussions)

### System Information

Include this information when reporting issues:

```bash
# Gather system info
flint-vault --version
go version
uname -a
echo $SHELL
echo $PATH
```

## 🔄 Updating

### Binary Installation

```bash
# Download latest binary and replace existing one
wget https://github.com/yourusername/flint-vault/releases/latest/download/flint-vault-linux-amd64
chmod +x flint-vault-linux-amd64
sudo mv flint-vault-linux-amd64 /usr/local/bin/flint-vault
```

### Source Installation

```bash
cd flint-vault
git pull origin main
go build -o flint-vault ./cmd
sudo cp flint-vault /usr/local/bin/
```

### Package Manager

```bash
# Homebrew
brew upgrade flint-vault

# Go install
go install github.com/yourusername/flint-vault/cmd@latest
```

## 🗑️ Uninstallation

### Binary Installation

```bash
# Remove binary
sudo rm /usr/local/bin/flint-vault

# Remove configuration (optional)
rm -rf ~/.config/flint-vault
```

### Package Manager

```bash
# Homebrew
brew uninstall flint-vault

# Snap
sudo snap remove flint-vault
```

---

**Next Steps**: After installation, read the [User Manual](USAGE.md) to learn how to use Flint Vault. 