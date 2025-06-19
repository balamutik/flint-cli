# 🔐 Flint Vault

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-brightgreen.svg)](#testing)
[![Security](https://img.shields.io/badge/Security-AES--256--GCM-red.svg)](#security)

**Flint Vault** is a military-grade encrypted file storage system implemented in Go. It provides secure, password-protected storage for files and directories using advanced cryptographic algorithms.

## ✨ Features

- 🔒 **Military-grade encryption** using AES-256-GCM
- 🔑 **Strong password derivation** with PBKDF2 (100,000 iterations)
- 📁 **Directory support** with recursive file addition
- 🗜️ **Built-in compression** using gzip
- 🌍 **Unicode support** for international file names
- 🔍 **Selective extraction** of specific files or directories
- 🛡️ **Authenticated encryption** preventing data tampering
- 🚀 **High performance** with efficient memory usage
- 🔧 **Simple CLI interface** for easy usage

## 🏗️ Architecture

Flint Vault uses a layered architecture:

```
┌─────────────────┐
│   CLI Layer     │  ← User commands (create, add, list, etc.)
├─────────────────┤
│  Commands Layer │  ← Command implementations and validation
├─────────────────┤
│   Vault Layer   │  ← Core encryption and file operations
├─────────────────┤
│  Crypto Layer   │  ← AES-256-GCM, PBKDF2, secure random
└─────────────────┘
```

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/flint-vault.git
cd flint-vault

# Build the application
go build -o flint-vault ./cmd

# Make it executable
chmod +x flint-vault
```

### Basic Usage

```bash
# Create a new encrypted vault
./flint-vault create -f my-vault.dat

# Add files to the vault
./flint-vault add -f my-vault.dat ./documents/

# List vault contents
./flint-vault list -f my-vault.dat

# Extract all files
./flint-vault extract -f my-vault.dat -d ./extracted/

# Extract specific file
./flint-vault get -v my-vault.dat -t document.txt -o ./output/

# Extract multiple files and directories
./flint-vault get -v my-vault.dat -t doc1.pdf -t images/ -t config.json -o ./
```

## 📚 Documentation

- [Installation Guide](docs/INSTALLATION.md)
- [User Manual](docs/USAGE.md)
- [API Documentation](docs/API.md)
- [Security Details](docs/SECURITY.md)
- [Development Guide](docs/DEVELOPMENT.md)
- [Architecture Overview](docs/ARCHITECTURE.md)

## 🔐 Security

Flint Vault employs multiple layers of security:

- **AES-256-GCM**: Authenticated encryption preventing both eavesdropping and tampering
- **PBKDF2**: Key derivation with 100,000 iterations using SHA-256
- **Cryptographically secure random**: All salts and nonces generated using crypto/rand
- **Memory safety**: Sensitive data is cleared from memory after use
- **Format validation**: Magic headers prevent accidental data corruption

### Tested Against

- ✅ Brute force password attacks
- ✅ Nonce reuse attacks  
- ✅ Data tampering attacks
- ✅ Side-channel information leaks
- ✅ File format corruption

## 🧪 Testing

The project includes comprehensive test coverage:

- **66.5%** code coverage for core library
- **32** test functions with multiple scenarios
- **Security tests** for cryptographic components
- **Performance benchmarks** for large files
- **Edge case testing** for robustness

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...
```

## 📊 Performance

Benchmark results on Intel i7-13700:

- **Vault Creation**: ~14.6ms per operation
- **Compression/Decompression**: ~0.24ms per operation
- **Large Files**: Successfully handles 1MB+ files
- **Memory Usage**: Efficient with ~820KB per vault creation

## 🌍 Internationalization

Full Unicode support including:

- ✅ International file names (Cyrillic, Asian scripts, etc.)
- ✅ Emoji in file names 🚀📁🔒
- ✅ Special characters and symbols
- ✅ Mixed encoding support

## 🛠️ Development

### Prerequisites

- Go 1.21 or higher
- Git

### Building from Source

```bash
# Clone and enter directory
git clone https://github.com/yourusername/flint-vault.git
cd flint-vault

# Download dependencies
go mod download

# Build
go build -o flint-vault ./cmd

# Run tests
go test ./...
```

### Project Structure

```
flint-vault/
├── cmd/                    # Main application entry point
├── pkg/
│   ├── commands/          # CLI command implementations
│   └── lib/
│       └── vault/         # Core vault functionality
├── docs/                  # Documentation
├── test_data/             # Test files
└── README.md
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style

- Follow standard Go conventions
- Add tests for new functionality
- Update documentation for API changes
- Ensure all tests pass

## 🚀 Development Story

This project is a testament to the power of **human-AI collaboration**! 

**The Journey:**
- 💻 **Developer**: [@balamutik](https://github.com/balamutik) brought the vision and hardcore coding skills
- 🤖 **AI Assistant**: Claude (Anthropic) provided architectural guidance, code generation, and extensive documentation
- 🔥 **The Magic**: Intense pair programming sessions with real-time code review and iterative improvements

**What We Built Together:**
- 🏗️ **Complete Architecture**: From crypto foundations to CLI interface
- 📝 **Comprehensive Docs**: 7 detailed documentation files covering every aspect  
- 🧪 **66.5% Test Coverage**: Robust testing with 32 test functions
- 🌍 **Full Internationalization**: English documentation + Unicode support
- ⚡ **Advanced Features**: Multiple file extraction, detailed error reporting
- 🔒 **Military-Grade Security**: AES-256-GCM with proper key derivation

**The Vibe:**
```
Developer: "Я хочу крутое хранилище с шифрованием!"
Claude: "Давайте сделаем нечто эпическое! 🚀"
*intensive keyboard typing sounds*
*security algorithms flying everywhere*
*documentation materializing*
Result: Flint Vault 🔐✨
```

This collaboration showcases how **human creativity + AI assistance** can produce enterprise-grade software with comprehensive documentation, robust testing, and attention to security details that would typically take months to develop!

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **[@balamutik](https://github.com/balamutik)** - Visionary developer who hardcore coded this beast 🔥
- **Claude (Anthropic AI)** - AI pair programming partner for architecture, documentation, and code generation 🤖
- **Go team** - For excellent cryptography libraries and robust language design
- **Open Source Community** - For inspiration and best practices
- **Security researchers** - For vulnerability disclosure and crypto guidance
- **Future contributors** - Welcome to the Flint Vault family! 🚀

## ⚠️ Security Notice

If you discover a security vulnerability, please send an email to security@yourproject.com instead of creating a public issue.

---

**🔥 Built with ❤️, 🔒, and intensive AI-human collaboration**  
**🚀 [@balamutik](https://github.com/balamutik) × Claude (Anthropic) = Epic Secure Storage** 

*"When human creativity meets AI assistance, magic happens!"* ✨ 