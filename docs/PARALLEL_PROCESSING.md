# ⚡ Parallel Processing Guide

This document provides comprehensive guidance on using Flint Vault's parallel processing capabilities for high-performance operations.

## 🎯 Overview

Flint Vault supports parallel processing for large-scale operations, providing significant performance improvements through configurable worker pools. This feature is especially beneficial for:

- **Large directories** with many files
- **Multi-GB datasets** requiring high throughput
- **Batch operations** where multiple files are processed
- **High-performance environments** with multiple CPU cores

## 🚀 Key Benefits

- **Performance**: Up to 25% speed improvement for large operations
- **Scalability**: Automatic or manual worker configuration
- **Efficiency**: Optimal resource utilization across CPU cores
- **Monitoring**: Real-time progress reporting
- **Safety**: Worker isolation prevents cross-contamination

## 🔧 Configuration

### Worker Pool Configuration

```go
// Default configuration (recommended)
config := vault.DefaultParallelConfig()
// Uses 2x CPU cores, 5-minute timeout

// Custom worker count
config.MaxConcurrency = 8

// With progress reporting
progressChan := make(chan string, 100)
config.ProgressChan = progressChan

// With timeout
config.Timeout = 10 * time.Minute

// With context for cancellation
ctx, cancel := context.WithCancel(context.Background())
config.Context = ctx
```

### CLI Configuration

```bash
# Auto-detection (recommended)
flint-vault add -v vault.flint -s ./data/

# Manual worker specification
flint-vault add -v vault.flint -s ./data/ --workers 8

# With progress reporting
flint-vault add -v vault.flint -s ./data/ --workers 4 --progress

# Memory-constrained systems
flint-vault add -v vault.flint -s ./data/ --workers 2
```

## 📊 Performance Guidelines

### Optimal Worker Counts

| System Type | CPU Cores | RAM | Recommended Workers | Use Case |
|-------------|-----------|-----|-------------------|----------|
| Laptop | 4 | 8GB | 2-4 | Conservative |
| Desktop | 8 | 16GB | 4-8 | Balanced |
| Workstation | 16+ | 32GB+ | 8-16 | High-performance |
| Server | 32+ | 64GB+ | 16+ | Maximum throughput |

### Worker Selection Strategy

```bash
# CPU-intensive operations (heavy compression)
--workers $(nproc)

# I/O-intensive operations (large files) - DEFAULT
--workers $(($(nproc) * 2))

# Memory-constrained environments
--workers 2

# High-performance systems
--workers 8
```

## 🛠️ API Usage

### Parallel Directory Addition

```go
config := vault.DefaultParallelConfig()
config.MaxConcurrency = 8

// Set up progress monitoring
progressChan := make(chan string, 100)
config.ProgressChan = progressChan

go func() {
    for msg := range progressChan {
        fmt.Printf("🔄 %s\n", msg)
    }
}()

stats, err := vault.AddDirectoryToVaultParallel(
    "vault.flint",
    "password",
    "./large-directory/",
    config)

close(progressChan)

if err != nil {
    log.Fatalf("Parallel operation failed: %v", err)
}

vault.PrintParallelStats(stats)
```

### Parallel File Extraction

```go
config := vault.DefaultParallelConfig()
config.MaxConcurrency = 6

targets := []string{
    "documents/report.pdf",
    "data/dataset.csv",
    "images/",
}

stats, err := vault.ExtractMultipleFilesFromVaultParallel(
    "vault.flint",
    "password",
    "./output/",
    targets,
    config)

if err != nil {
    log.Fatalf("Parallel extraction failed: %v", err)
}

fmt.Printf("✅ Extracted %d files in %v\n", 
    stats.SuccessfulFiles, stats.Duration)
```

## 📈 Performance Monitoring

### Real-Time Statistics

```go
type ParallelStats struct {
    TotalFiles      int64         // Total files processed
    SuccessfulFiles int64         // Successfully processed files
    FailedFiles     int64         // Failed files
    TotalSize       int64         // Total size processed (bytes)
    Duration        time.Duration // Total processing duration
    Errors          []error       // Collection of errors
}
```

### Progress Reporting

```go
progressChan := make(chan string, 100)
config.ProgressChan = progressChan

// Custom progress handler
go func() {
    var fileCount int64
    for msg := range progressChan {
        if strings.Contains(msg, "Adding:") {
            fileCount++
            fmt.Printf("[%d] %s\n", fileCount, msg)
        } else if strings.Contains(msg, "Processing") {
            fmt.Printf("📊 %s\n", msg)
        }
    }
}()
```

### Performance Metrics

```bash
# Example output from parallel operations
📊 Operation Statistics:
✅ Successfully processed: 245 files
❌ Failed: 0 files
📏 Total size: 1.2 GB
⏱️  Duration: 12.3 seconds
📈 Average speed: 97.6 MB/s
🔧 Workers utilized: 8
```

## 🔍 Troubleshooting

### Common Issues

#### Performance Not Improving
```bash
# Check if operation is CPU or I/O bound
# CPU-bound: reduce workers to CPU count
flint-vault add -v vault.flint -s ./data/ --workers $(nproc)

# I/O-bound: use more workers
flint-vault add -v vault.flint -s ./data/ --workers $(($(nproc) * 2))
```

#### Memory Issues
```bash
# Check available memory
free -h

# Reduce worker count for large operations
flint-vault add -v vault.flint -s ./large-data/ --workers 2

# Calculate memory requirements
# Base: 3.2x vault size + (100-200MB per worker)
```

#### System Resource Monitoring
```bash
# Monitor during operations
top -p $(pgrep flint-vault)

# Check I/O utilization
iostat 1

# Monitor memory usage
watch -n 1 'free -h'
```

### Optimization Tips

1. **Start with auto-detection**: Let Flint Vault choose optimal workers
2. **Monitor resource usage**: Watch CPU and memory during operations
3. **Scale gradually**: Increase workers incrementally for large datasets
4. **Consider file types**: Text files compress better, binary files process faster
5. **Use progress reporting**: Monitor operations for performance tuning

## 🎯 Use Cases

### Large Directory Processing
```bash
# Adding 1000+ files efficiently
flint-vault add -v archive.flint -s ./project-archive/ --workers 8 --progress
```

### Selective High-Speed Extraction
```bash
# Extract specific files with maximum speed
flint-vault extract -v large-vault.flint -o ./output/ \
    --files "data/,documents/important.pdf,logs/" --workers 6
```

### Batch Archive Creation
```bash
# Process multiple directories in parallel
for dir in project1 project2 project3; do
    flint-vault add -v "archive-$(date +%Y%m%d).flint" -s "./$dir/" --workers 4 &
done
wait
```

### High-Performance Backup
```bash
# Maximum throughput backup
flint-vault add -v "backup-$(date +%Y%m%d).flint" -s /data/ \
    --workers $(($(nproc) * 2)) --progress
```

## 🔒 Security Considerations

### Worker Isolation
- Each worker operates in isolated memory space
- No cross-contamination between workers
- Secure cleanup of sensitive data per worker

### Performance vs Security
- All security features maintained with parallel processing
- No reduction in encryption strength
- Consistent timing across all worker configurations

## 🚀 Best Practices

1. **Auto-detection first**: Use default configuration for most cases
2. **Monitor and adjust**: Watch performance and adjust worker count
3. **Consider system resources**: Don't exceed available CPU/memory
4. **Use progress reporting**: Monitor long-running operations
5. **Test configurations**: Find optimal settings for your use case

## 📋 Summary

**Parallel processing in Flint Vault provides:**
- **15-25% performance improvement** for large operations
- **Automatic optimization** with sensible defaults
- **Full security preservation** across all worker configurations
- **Comprehensive monitoring** and error handling
- **Scalable architecture** for enterprise use

**Recommended approach:**
1. Start with auto-detection
2. Monitor performance
3. Adjust worker count based on results
4. Use progress reporting for large operations
5. Scale resources appropriately for your workload

---

*Parallel Processing Guide - June 2025*
*Tested and validated with 2.45 GB datasets* 