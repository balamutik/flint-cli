# 🔄 Migration Guide

This guide helps users upgrade to the latest version of Flint Vault with parallel processing capabilities.

## 🚀 What's New

### Major Improvements
- **Parallel Processing**: Configurable worker pools for faster operations
- **Go 1.24+ Requirement**: Updated from Go 1.21+
- **Enhanced Performance**: Up to 25% speed improvement
- **Progress Reporting**: Real-time status for long operations
- **Auto-Optimization**: Intelligent worker count detection

## 📋 Migration Checklist

### 1. Update Go Version
```bash
# Check current Go version
go version

# If below 1.24, update Go
# Download from: https://golang.org/dl/
# Or use package manager
brew install go  # macOS
sudo apt install golang-go  # Ubuntu
```

### 2. Update Flint Vault
```bash
# Pull latest changes
git pull origin main

# Rebuild application
go build -o flint-vault ./cmd

# Verify update
./flint-vault --help
```

### 3. Learn New Features
```bash
# Check new parallel processing options
flint-vault add --help
flint-vault extract --help

# Test with small dataset
flint-vault add -v test.flint -s ./small-folder/ --workers 2 --progress
```

## ⚡ New Command Options

### Before (Still Works)
```bash
# Original commands continue to work
flint-vault add -v vault.flint -s ./data/
flint-vault extract -v vault.flint -o ./output/
```

### After (Enhanced Performance)
```bash
# New parallel processing options
flint-vault add -v vault.flint -s ./data/ --workers 8 --progress
flint-vault extract -v vault.flint -o ./output/ --workers 4

# Auto-detection (recommended)
flint-vault add -v vault.flint -s ./data/  # Uses optimal worker count
```

## 🔧 Configuration Migration

### Old Approach
```bash
# Manual optimization
export GOMAXPROCS=$(nproc)
flint-vault add -v vault.flint -s ./large-data/
```

### New Approach
```bash
# Built-in optimization
flint-vault add -v vault.flint -s ./large-data/ --workers auto

# Or manual specification
flint-vault add -v vault.flint -s ./large-data/ --workers 8
```

## 📊 Performance Optimization

### Recommended Worker Counts

| Your System | Old Performance | New Recommended | Expected Improvement |
|-------------|----------------|-----------------|---------------------|
| 4 cores, 8GB RAM | Baseline | `--workers 4` | +15-20% |
| 8 cores, 16GB RAM | Baseline | `--workers 8` | +20-25% |
| 16+ cores, 32GB+ RAM | Baseline | `--workers 12-16` | +25%+ |

### Memory Considerations
```bash
# Check available memory before large operations
free -h

# For memory-constrained systems
flint-vault add -v vault.flint -s ./data/ --workers 2

# For high-memory systems
flint-vault add -v vault.flint -s ./data/ --workers 8
```

## 🔄 Script Migration

### Old Bash Scripts
```bash
#!/bin/bash
# Old approach
for dir in */; do
    flint-vault add -v "backup-$(date +%Y%m%d).flint" -s "$dir"
done
```

### New Optimized Scripts
```bash
#!/bin/bash
# New parallel approach
WORKERS=$(( $(nproc) * 2 ))

for dir in */; do
    flint-vault add -v "backup-$(date +%Y%m%d).flint" -s "$dir" \
        --workers "$WORKERS" --progress
done
```

## 🛡️ Backward Compatibility

### Vault Files
- ✅ **Fully compatible**: All existing vault files work unchanged
- ✅ **No migration required**: Vault format unchanged
- ✅ **Same security**: All encryption remains identical

### CLI Commands
- ✅ **All old commands work**: No breaking changes
- ✅ **New flags optional**: Default behavior maintained
- ✅ **Gradual adoption**: Use new features when ready

### API Compatibility
```go
// Old API calls still work
err := vault.CreateVault(path, password)
err := vault.AddFileToVault(vaultPath, password, filePath)
entries, err := vault.ListVault(vaultPath, password)

// New parallel API available
config := vault.DefaultParallelConfig()
stats, err := vault.AddDirectoryToVaultParallel(vaultPath, password, dirPath, config)
```

## 🎯 Adoption Strategy

### Phase 1: Basic Update
1. Update Go to 1.24+
2. Rebuild Flint Vault
3. Continue using existing commands

### Phase 2: Test Parallel Features
1. Try `--workers 2` on small datasets
2. Monitor performance improvements
3. Test progress reporting

### Phase 3: Full Optimization
1. Use auto-detection for worker counts
2. Optimize scripts with parallel processing
3. Monitor system resources

## 🔍 Troubleshooting

### Performance Issues
```bash
# If parallel processing doesn't improve performance
flint-vault add -v vault.flint -s ./data/ --workers 1  # Single worker
flint-vault add -v vault.flint -s ./data/ --workers 2  # Conservative

# Monitor resource usage
top -p $(pgrep flint-vault)
```

### Memory Issues
```bash
# Reduce worker count if out of memory
flint-vault add -v vault.flint -s ./large-data/ --workers 2

# Check memory requirements
# Rule: 3.2x vault size + (200MB × workers)
```

### Compatibility Issues
```bash
# If any issues, fall back to single-threaded mode
flint-vault add -v vault.flint -s ./data/ --workers 1

# Or use original commands (no --workers flag)
flint-vault add -v vault.flint -s ./data/
```

## 📈 Performance Testing

### Before/After Comparison
```bash
# Test with your data
echo "Testing old approach..."
time flint-vault add -v test1.flint -s ./test-data/

echo "Testing new parallel approach..."
time flint-vault add -v test2.flint -s ./test-data/ --workers 4

# Compare vault sizes (should be identical)
ls -la test*.flint
```

## 🎉 Migration Complete

After migration, you should have:
- ✅ Go 1.24+ installed
- ✅ Latest Flint Vault built
- ✅ Understanding of new parallel options
- ✅ Improved performance for large operations
- ✅ All existing vault files working unchanged

## 📞 Support

If you encounter issues during migration:
1. Check [Troubleshooting Guide](INSTALLATION.md#troubleshooting)
2. Review [Parallel Processing Guide](PARALLEL_PROCESSING.md)
3. Create an issue on GitHub with system details

---

*Migration completed? See the [Parallel Processing Guide](PARALLEL_PROCESSING.md) for advanced usage!* 