// Package vault provides secure file storage with military-grade AES-256 encryption.
// This package implements optimized streaming operations with parallel processing
// for high-performance file vault management.
package vault

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/term"
)

const (
	// Magic header for vault file identification
	VaultMagic = "FLINT001"

	// Cryptographic parameters
	KeyLength   = 32     // AES-256
	SaltLength  = 32     // 256-bit salt
	NonceLength = 12     // GCM nonce
	PBKDF2Iters = 100000 // PBKDF2 iterations (recommended minimum)

	// Current vault format version
	CurrentVaultVersion = 2

	// Buffer size for streaming operations (1MB)
	StreamBufferSize = 1024 * 1024
)

// FileEntry represents a file or directory entry in vault with optimizations
type FileEntry struct {
	Path           string    `json:"path"`            // Path to file/directory
	Name           string    `json:"name"`            // Name of file/directory
	IsDir          bool      `json:"is_dir"`          // Whether it's a directory
	Size           int64     `json:"size"`            // Original file size (0 for directories)
	CompressedSize int64     `json:"compressed_size"` // Size after compression
	Mode           uint32    `json:"mode"`            // Access permissions
	ModTime        time.Time `json:"mod_time"`        // Last modification time
	Offset         int64     `json:"offset"`          // Offset in vault file where data starts
	SHA256Hash     [32]byte  `json:"sha256_hash"`     // SHA-256 hash for integrity verification
}

// VaultDirectory contains only metadata - NO file contents in memory
type VaultDirectory struct {
	Version   uint32      `json:"version"`    // Vault format version
	Entries   []FileEntry `json:"entries"`    // File/directory metadata only
	CreatedAt time.Time   `json:"created_at"` // Vault creation time
	Comment   string      `json:"comment"`    // Vault comment
}

// VaultHeader contains vault metadata
type VaultHeader struct {
	Magic         [8]byte  // "FLINT001"
	Version       uint32   // Format version
	Iterations    uint32   // PBKDF2 iteration count
	Salt          [32]byte // Salt for key derivation
	Nonce         [12]byte // Nonce for AES-GCM
	DirectorySize uint64   // Size of encrypted directory data
}

// ParallelConfig configures parallel processing parameters
type ParallelConfig struct {
	MaxConcurrency int             // Maximum number of concurrent workers
	Timeout        time.Duration   // Timeout for individual operations
	ProgressChan   chan string     // Progress reporting channel (optional)
	Context        context.Context // Context for cancellation
}

// ParallelStats tracks parallel operation statistics
type ParallelStats struct {
	TotalFiles      int64         // Total files processed
	SuccessfulFiles int64         // Successfully processed files
	FailedFiles     int64         // Failed files
	TotalSize       int64         // Total size processed (bytes)
	Duration        time.Duration // Total processing duration
	Errors          []error       // Collection of errors encountered
	ErrorsMutex     sync.Mutex    // Mutex for thread-safe error collection
}

// DefaultParallelConfig creates default parallel processing configuration
func DefaultParallelConfig() *ParallelConfig {
	return &ParallelConfig{
		MaxConcurrency: runtime.NumCPU() * 2, // 2x CPU cores for I/O bound operations
		Timeout:        5 * time.Minute,
		Context:        context.Background(),
	}
}

// Global vault access synchronization
var vaultMutexes = make(map[string]*sync.Mutex)
var vaultMutexesLock sync.Mutex

// getVaultMutex returns a mutex for a specific vault file
func getVaultMutex(vaultPath string) *sync.Mutex {
	vaultMutexesLock.Lock()
	defer vaultMutexesLock.Unlock()

	if mutex, exists := vaultMutexes[vaultPath]; exists {
		return mutex
	}

	mutex := &sync.Mutex{}
	vaultMutexes[vaultPath] = mutex
	return mutex
}

// ========================
// VAULT CREATION
// ========================

// CreateVault creates a new optimized vault file
func CreateVault(path string, password string) error {
	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}

	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check that file doesn't exist
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("vault file already exists: %s", path)
	}

	// Create empty vault directory
	vaultDir := VaultDirectory{
		Version:   CurrentVaultVersion,
		Entries:   []FileEntry{},
		CreatedAt: time.Now(),
		Comment:   "Encrypted Flint Vault Storage (Optimized)",
	}

	return saveVaultDirectory(path, password, vaultDir)
}

// ========================
// PASSWORD INPUT
// ========================

// ReadPasswordSecurely securely reads password from terminal without displaying characters
func ReadPasswordSecurely(prompt string) (string, error) {
	fmt.Print(prompt)

	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input

	if err != nil {
		return "", fmt.Errorf("password read error: %w", err)
	}

	if len(password) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}

	return string(password), nil
}

// ========================
// SINGLE FILE OPERATIONS
// ========================

// AddFileToVault adds a file to vault with streaming and integrity checking
func AddFileToVault(vaultPath, password, filePath string) error {
	// Check if file is a directory
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file info error: %w", err)
	}

	if fileInfo.IsDir() {
		return AddDirectoryToVault(vaultPath, password, filePath)
	}

	// Load existing vault directory
	vaultDir, err := loadVaultDirectory(vaultPath, password)
	if err != nil {
		return fmt.Errorf("vault directory load error: %w", err)
	}

	// First pass: calculate metadata (hash and compressed size) using streaming
	fileHash, compressedSize, err := calculateFileMetadata(filePath)
	if err != nil {
		return fmt.Errorf("metadata calculation error: %w", err)
	}

	// Create file entry with metadata (we'll calculate offset later)
	entry := FileEntry{
		Path:           filepath.Clean(filePath),
		Name:           fileInfo.Name(),
		IsDir:          false,
		Size:           fileInfo.Size(),
		CompressedSize: compressedSize,
		Mode:           uint32(fileInfo.Mode()),
		ModTime:        fileInfo.ModTime(),
		Offset:         0, // Will be set in addFileToVaultStreaming
		SHA256Hash:     fileHash,
	}

	// Update vault directory
	found := false
	for i, existingEntry := range vaultDir.Entries {
		if existingEntry.Path == entry.Path {
			vaultDir.Entries[i] = entry // Update existing
			found = true
			break
		}
	}

	if !found {
		vaultDir.Entries = append(vaultDir.Entries, entry) // Add new
	}

	// Second pass: stream compressed data directly to vault file
	return addFileToVaultStreaming(vaultPath, password, *vaultDir, filePath)
}

// AddDirectoryToVault adds a directory and all its contents to the vault
func AddDirectoryToVault(vaultPath, password, dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return addDirectoryEntry(vaultPath, password, path, info, dirPath)
		}

		return AddFileToVault(vaultPath, password, path)
	})
}

// ExtractFromVault extracts all files from vault to specified directory
func ExtractFromVault(vaultPath, password, outputDir string) error {
	vaultDir, err := loadVaultDirectory(vaultPath, password)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("output directory creation error: %w", err)
	}

	for _, entry := range vaultDir.Entries {
		if entry.IsDir {
			dirPath := filepath.Join(outputDir, entry.Path)
			if err := os.MkdirAll(dirPath, os.FileMode(entry.Mode)); err != nil {
				return fmt.Errorf("directory creation error: %w", err)
			}
		} else {
			if err := extractFileEntry(vaultPath, password, entry, outputDir); err != nil {
				return fmt.Errorf("file extraction error for %s: %w", entry.Path, err)
			}
		}
	}

	return nil
}

// GetFromVault extracts specific files from vault
func GetFromVault(vaultPath, password, outputDir string, targetPaths []string) error {
	vaultDir, err := loadVaultDirectory(vaultPath, password)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("output directory creation error: %w", err)
	}

	// Create a map for fast lookup
	targetMap := make(map[string]bool)
	for _, path := range targetPaths {
		targetMap[filepath.Clean(path)] = true
	}

	for _, entry := range vaultDir.Entries {
		if targetMap[entry.Path] {
			if entry.IsDir {
				dirPath := filepath.Join(outputDir, entry.Path)
				if err := os.MkdirAll(dirPath, os.FileMode(entry.Mode)); err != nil {
					return fmt.Errorf("directory creation error: %w", err)
				}
			} else {
				if err := extractFileEntry(vaultPath, password, entry, outputDir); err != nil {
					return fmt.Errorf("file extraction error for %s: %w", entry.Path, err)
				}
			}
		}
	}

	return nil
}

// ListVault returns list of files in the vault
func ListVault(vaultPath, password string) ([]FileEntry, error) {
	vaultDir, err := loadVaultDirectory(vaultPath, password)
	if err != nil {
		return nil, err
	}

	return vaultDir.Entries, nil
}

// ========================
// PARALLEL OPERATIONS
// ========================

// AddMultipleFilesToVaultParallel adds multiple files to vault in parallel
func AddMultipleFilesToVaultParallel(vaultPath, password string, filePaths []string, config *ParallelConfig) (*ParallelStats, error) {
	stats := &ParallelStats{
		TotalFiles: int64(len(filePaths)),
	}
	startTime := time.Now()

	semaphore := make(chan struct{}, config.MaxConcurrency)
	var wg sync.WaitGroup
	vaultMutex := getVaultMutex(vaultPath) // Get vault-specific mutex

	for _, filePath := range filePaths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if config.ProgressChan != nil {
				config.ProgressChan <- fmt.Sprintf("Processing file: %s", path)
			}

			// Synchronize vault access
			vaultMutex.Lock()
			err := AddFileToVault(vaultPath, password, path)
			vaultMutex.Unlock()

			if err != nil {
				atomic.AddInt64(&stats.FailedFiles, 1)
				stats.ErrorsMutex.Lock()
				stats.Errors = append(stats.Errors, fmt.Errorf("failed to add %s: %w", path, err))
				stats.ErrorsMutex.Unlock()
			} else {
				atomic.AddInt64(&stats.SuccessfulFiles, 1)
				if info, err := os.Stat(path); err == nil {
					atomic.AddInt64(&stats.TotalSize, info.Size())
				}
			}
		}(filePath)
	}

	wg.Wait()
	stats.Duration = time.Since(startTime)

	if len(stats.Errors) > 0 {
		return stats, fmt.Errorf("parallel processing completed with %d errors", len(stats.Errors))
	}

	return stats, nil
}

// AddDirectoryToVaultParallel adds directory to vault with parallel processing
func AddDirectoryToVaultParallel(vaultPath, password, dirPath string, config *ParallelConfig) (*ParallelStats, error) {
	// Collect all files in directory
	var filePaths []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			filePaths = append(filePaths, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("directory traversal error: %w", err)
	}

	return AddMultipleFilesToVaultParallel(vaultPath, password, filePaths, config)
}

// ExtractMultipleFilesFromVaultParallel extracts multiple files from vault in parallel
func ExtractMultipleFilesFromVaultParallel(vaultPath, password, outputDir string, targetPaths []string, config *ParallelConfig) (*ParallelStats, error) {
	vaultDir, err := loadVaultDirectory(vaultPath, password)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("output directory creation error: %w", err)
	}

	// Create a map for fast lookup
	targetMap := make(map[string]bool)
	for _, path := range targetPaths {
		targetMap[filepath.Clean(path)] = true
	}

	// Filter entries to extract
	var entriesToExtract []FileEntry
	for _, entry := range vaultDir.Entries {
		if targetMap[entry.Path] {
			entriesToExtract = append(entriesToExtract, entry)
		}
	}

	stats := &ParallelStats{
		TotalFiles: int64(len(entriesToExtract)),
	}
	startTime := time.Now()

	semaphore := make(chan struct{}, config.MaxConcurrency)
	var wg sync.WaitGroup

	for _, entry := range entriesToExtract {
		wg.Add(1)
		go func(e FileEntry) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if config.ProgressChan != nil {
				config.ProgressChan <- fmt.Sprintf("Extracting: %s", e.Path)
			}

			if e.IsDir {
				dirPath := filepath.Join(outputDir, e.Path)
				if err := os.MkdirAll(dirPath, os.FileMode(e.Mode)); err != nil {
					atomic.AddInt64(&stats.FailedFiles, 1)
					stats.ErrorsMutex.Lock()
					stats.Errors = append(stats.Errors, fmt.Errorf("failed to create directory %s: %w", e.Path, err))
					stats.ErrorsMutex.Unlock()
				} else {
					atomic.AddInt64(&stats.SuccessfulFiles, 1)
				}
			} else {
				if err := extractFileEntry(vaultPath, password, e, outputDir); err != nil {
					atomic.AddInt64(&stats.FailedFiles, 1)
					stats.ErrorsMutex.Lock()
					stats.Errors = append(stats.Errors, fmt.Errorf("failed to extract %s: %w", e.Path, err))
					stats.ErrorsMutex.Unlock()
				} else {
					atomic.AddInt64(&stats.SuccessfulFiles, 1)
					atomic.AddInt64(&stats.TotalSize, e.Size)
				}
			}
		}(entry)
	}

	wg.Wait()
	stats.Duration = time.Since(startTime)

	if len(stats.Errors) > 0 {
		return stats, fmt.Errorf("parallel extraction completed with %d errors", len(stats.Errors))
	}

	return stats, nil
}

// PrintParallelStats prints detailed statistics from parallel operations
func PrintParallelStats(stats *ParallelStats) {
	fmt.Printf("\n📊 Processing Statistics:\n")
	fmt.Printf("  📁 Total files: %d\n", stats.TotalFiles)
	fmt.Printf("  ✅ Successful: %d\n", stats.SuccessfulFiles)
	fmt.Printf("  ❌ Failed: %d\n", stats.FailedFiles)
	fmt.Printf("  📏 Total size: %s\n", formatFileSize(stats.TotalSize))
	fmt.Printf("  ⏱️  Duration: %v\n", stats.Duration)

	if stats.TotalSize > 0 && stats.Duration > 0 {
		mbps := float64(stats.TotalSize) / (1024 * 1024) / stats.Duration.Seconds()
		fmt.Printf("  🚀 Throughput: %.1f MB/s\n", mbps)
	}

	if len(stats.Errors) > 0 {
		fmt.Printf("\n⚠️  Errors encountered:\n")
		for i, err := range stats.Errors {
			if i < 5 { // Show first 5 errors
				fmt.Printf("  - %v\n", err)
			} else if i == 5 {
				fmt.Printf("  ... and %d more errors\n", len(stats.Errors)-5)
				break
			}
		}
	}
}

// ========================
// HELPER FUNCTIONS
// ========================

// formatFileSize formats file size in human-readable format
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.1fTB", float64(size)/TB)
	case size >= GB:
		return fmt.Sprintf("%.1fGB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.1fMB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.1fKB", float64(size)/KB)
	default:
		return fmt.Sprintf("%dB", size)
	}
}

// addDirectoryEntry adds a directory entry to vault
func addDirectoryEntry(vaultPath, password, dirPath string, info os.FileInfo, basePath string) error {
	// Load existing vault directory
	vaultDir, err := loadVaultDirectory(vaultPath, password)
	if err != nil {
		return fmt.Errorf("vault directory load error: %w", err)
	}

	// Create directory entry
	relativePath, err := filepath.Rel(basePath, dirPath)
	if err != nil {
		relativePath = dirPath
	}

	entry := FileEntry{
		Path:           filepath.Clean(relativePath),
		Name:           info.Name(),
		IsDir:          true,
		Size:           0,
		CompressedSize: 0,
		Mode:           uint32(info.Mode()),
		ModTime:        info.ModTime(),
		Offset:         0,
		SHA256Hash:     [32]byte{}, // Empty hash for directories
	}

	// Update vault directory
	found := false
	for i, existingEntry := range vaultDir.Entries {
		if existingEntry.Path == entry.Path {
			vaultDir.Entries[i] = entry // Update existing
			found = true
			break
		}
	}

	if !found {
		vaultDir.Entries = append(vaultDir.Entries, entry) // Add new
	}

	return updateVaultDirectory(vaultPath, password, *vaultDir)
}

// calculateFileMetadata calculates file hash and compressed size using streaming
func calculateFileMetadata(filePath string) ([32]byte, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return [32]byte{}, 0, fmt.Errorf("file open error: %w", err)
	}
	defer file.Close()

	// Create hasher and compression size counter
	hasher := sha256.New()

	// Use a buffer to count compressed size efficiently
	var compressedSize int64
	compressedCounter := &countingWriter{count: &compressedSize}

	// Create gzip writer for size calculation
	gzipWriter := gzip.NewWriter(compressedCounter)
	defer gzipWriter.Close()

	// Create multi-writer to simultaneously hash and count compressed size
	multiWriter := io.MultiWriter(hasher, gzipWriter)

	// Stream through data
	buffer := make([]byte, StreamBufferSize)
	if _, err := io.CopyBuffer(multiWriter, file, buffer); err != nil {
		return [32]byte{}, 0, fmt.Errorf("metadata calculation error: %w", err)
	}

	// Close gzip writer to get final compressed size
	if err := gzipWriter.Close(); err != nil {
		return [32]byte{}, 0, fmt.Errorf("compression finalization error: %w", err)
	}

	// Get final hash
	var hash [32]byte
	copy(hash[:], hasher.Sum(nil))

	return hash, compressedSize, nil
}

// countingWriter counts bytes written to it
type countingWriter struct {
	count *int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n := len(p)
	*c.count += int64(n)
	return n, nil
}

// addFileToVaultStreaming adds file to vault using true streaming approach
func addFileToVaultStreaming(vaultPath, password string, vaultDir VaultDirectory, filePath string) error {
	// Create temporary file for the new vault
	tempPath := vaultPath + ".tmp"
	defer os.Remove(tempPath) // Clean up temp file

	// Open original vault file for reading
	originalFile, err := os.Open(vaultPath)
	if err != nil {
		return fmt.Errorf("original file open error: %w", err)
	}
	defer originalFile.Close()

	// Read original header
	var originalHeader VaultHeader
	if err := binary.Read(originalFile, binary.LittleEndian, &originalHeader); err != nil {
		return fmt.Errorf("header read error: %w", err)
	}

	// Calculate file offsets for all entries in the directory
	var totalDataSize int64 = 0
	for i := range vaultDir.Entries {
		if !vaultDir.Entries[i].IsDir {
			vaultDir.Entries[i].Offset = totalDataSize
			totalDataSize += vaultDir.Entries[i].CompressedSize
		}
	}

	// Serialize and compress new directory
	jsonData, err := json.Marshal(vaultDir)
	if err != nil {
		return fmt.Errorf("directory serialization error: %w", err)
	}

	compressedDir, err := compressData(jsonData)
	if err != nil {
		return fmt.Errorf("directory compression error: %w", err)
	}

	// Encrypt directory with existing parameters
	key := pbkdf2.Key([]byte(password), originalHeader.Salt[:], int(originalHeader.Iterations), KeyLength, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("AES cipher creation error: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("GCM creation error: %w", err)
	}

	encryptedDir := gcm.Seal(nil, originalHeader.Nonce[:], compressedDir, nil)

	// Create new header with updated directory size
	newHeader := originalHeader
	newHeader.DirectorySize = uint64(len(encryptedDir))

	// Create temporary file
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("temp file creation error: %w", err)
	}
	defer tempFile.Close()

	// Write new header
	if err := binary.Write(tempFile, binary.LittleEndian, newHeader); err != nil {
		return fmt.Errorf("header write error: %w", err)
	}

	// Write encrypted directory
	if _, err := tempFile.Write(encryptedDir); err != nil {
		return fmt.Errorf("directory write error: %w", err)
	}

	// Stream existing file data from original vault
	headerSize := int64(binary.Size(VaultHeader{}))
	originalDirectorySize := int64(originalHeader.DirectorySize)
	fileDataOffset := headerSize + originalDirectorySize

	if _, err := originalFile.Seek(fileDataOffset, io.SeekStart); err != nil {
		return fmt.Errorf("original file seek error: %w", err)
	}

	buffer := make([]byte, StreamBufferSize)
	if _, err := io.CopyBuffer(tempFile, originalFile, buffer); err != nil {
		return fmt.Errorf("existing file data copy error: %w", err)
	}

	// Now stream new file data directly from source file with compression
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("source file open error: %w", err)
	}
	defer sourceFile.Close()

	// Create gzip writer to compress directly to vault
	gzipWriter := gzip.NewWriter(tempFile)
	if _, err := io.CopyBuffer(gzipWriter, sourceFile, buffer); err != nil {
		gzipWriter.Close()
		return fmt.Errorf("file compression streaming error: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("compression finalization error: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("temp file sync error: %w", err)
	}

	tempFile.Close()
	originalFile.Close()

	// Replace original file with temporary file
	if err := os.Rename(tempPath, vaultPath); err != nil {
		return fmt.Errorf("file replacement error: %w", err)
	}

	// Clear sensitive data
	for i := range key {
		key[i] = 0
	}

	return nil
}

// saveVaultDirectory saves initial vault directory to file
func saveVaultDirectory(path, password string, vaultDir VaultDirectory) error {
	// Serialize directory
	jsonData, err := json.Marshal(vaultDir)
	if err != nil {
		return fmt.Errorf("directory serialization error: %w", err)
	}

	// Compress directory
	compressedDir, err := compressData(jsonData)
	if err != nil {
		return fmt.Errorf("directory compression error: %w", err)
	}

	// Generate cryptographic parameters
	var salt [SaltLength]byte
	var nonce [NonceLength]byte

	if _, err := rand.Read(salt[:]); err != nil {
		return fmt.Errorf("salt generation error: %w", err)
	}

	if _, err := rand.Read(nonce[:]); err != nil {
		return fmt.Errorf("nonce generation error: %w", err)
	}

	// Derive key
	key := pbkdf2.Key([]byte(password), salt[:], PBKDF2Iters, KeyLength, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("AES cipher creation error: %w", err)
	}

	// Create GCM for encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("GCM creation error: %w", err)
	}

	// Encrypt directory
	encryptedDir := gcm.Seal(nil, nonce[:], compressedDir, nil)

	// Create header
	header := VaultHeader{
		Version:       CurrentVaultVersion,
		Iterations:    PBKDF2Iters,
		Salt:          salt,
		Nonce:         nonce,
		DirectorySize: uint64(len(encryptedDir)),
	}
	copy(header.Magic[:], VaultMagic)

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("file creation error: %w", err)
	}
	defer file.Close()

	// Write header
	if err := binary.Write(file, binary.LittleEndian, header); err != nil {
		return fmt.Errorf("header write error: %w", err)
	}

	// Write encrypted directory
	if _, err := file.Write(encryptedDir); err != nil {
		return fmt.Errorf("directory write error: %w", err)
	}

	// Clear sensitive data
	for i := range key {
		key[i] = 0
	}

	return nil
}

// ========================
// STORAGE FUNCTIONS (from common.go)
// ========================

// loadVaultDirectory loads only the vault directory (metadata) - memory efficient
func loadVaultDirectory(path, password string) (*VaultDirectory, error) {
	// First validate the vault file format
	if err := ValidateVaultFile(path); err != nil {
		return nil, err
	}

	// Open vault file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("file open error: %w", err)
	}
	defer file.Close()

	// Read header
	var header VaultHeader
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("header read error: %w", err)
	}

	// Derive key from password
	key := pbkdf2.Key([]byte(password), header.Salt[:], int(header.Iterations), KeyLength, sha256.New)

	// Clear password from memory
	passwordBytes := []byte(password)
	for i := range passwordBytes {
		passwordBytes[i] = 0
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AES cipher creation error: %w", err)
	}

	// Create GCM for decryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCM creation error: %w", err)
	}

	// Read encrypted directory data
	encryptedDir := make([]byte, header.DirectorySize)
	if _, err := io.ReadFull(file, encryptedDir); err != nil {
		return nil, fmt.Errorf("encrypted directory read error: %w", err)
	}

	// Decrypt directory data
	compressedData, err := gcm.Open(nil, header.Nonce[:], encryptedDir, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: invalid password or corrupted data")
	}

	// Decompress directory data
	jsonData, err := decompressData(compressedData)
	if err != nil {
		return nil, fmt.Errorf("directory decompression error: %w", err)
	}

	// Deserialize JSON
	var vaultDir VaultDirectory
	if err := json.Unmarshal(jsonData, &vaultDir); err != nil {
		return nil, fmt.Errorf("directory deserialization error: %w", err)
	}

	// Clear key from memory
	for i := range key {
		key[i] = 0
	}

	return &vaultDir, nil
}

// updateVaultDirectory updates the vault directory in the vault file
func updateVaultDirectory(vaultPath, password string, vaultDir VaultDirectory) error {
	// Use optimized streaming version
	return updateVaultDirectoryStreamingOptimized(vaultPath, password, vaultDir)
}

// updateVaultDirectoryStreamingOptimized optimized version for memory efficiency
func updateVaultDirectoryStreamingOptimized(vaultPath, password string, vaultDir VaultDirectory) error {
	// Recalculate file offsets in new structure
	var totalDataSize int64 = 0
	for i := range vaultDir.Entries {
		if !vaultDir.Entries[i].IsDir {
			vaultDir.Entries[i].Offset = totalDataSize
			totalDataSize += vaultDir.Entries[i].CompressedSize
		}
	}

	// Open source file only for reading header
	sourceFile, err := os.Open(vaultPath)
	if err != nil {
		return fmt.Errorf("source file open error: %w", err)
	}
	defer sourceFile.Close()

	// Read only header (not entire file!)
	var header VaultHeader
	if err := binary.Read(sourceFile, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("header read error: %w", err)
	}

	// Serialize and compress new directory
	jsonData, err := json.Marshal(vaultDir)
	if err != nil {
		return fmt.Errorf("directory serialization error: %w", err)
	}

	compressedDir, err := compressData(jsonData)
	if err != nil {
		return fmt.Errorf("directory compression error: %w", err)
	}

	// Encrypt directory with existing parameters
	key := pbkdf2.Key([]byte(password), header.Salt[:], int(header.Iterations), KeyLength, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("AES cipher creation error: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("GCM creation error: %w", err)
	}

	encryptedDir := gcm.Seal(nil, header.Nonce[:], compressedDir, nil)

	// Update header
	header.DirectorySize = uint64(len(encryptedDir))

	// Create temporary file
	tempPath := vaultPath + ".tmp"
	defer os.Remove(tempPath)

	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("temp file creation error: %w", err)
	}
	defer tempFile.Close()

	// Write new header
	if err := binary.Write(tempFile, binary.LittleEndian, header); err != nil {
		return fmt.Errorf("header write error: %w", err)
	}

	// Write encrypted directory
	if _, err := tempFile.Write(encryptedDir); err != nil {
		return fmt.Errorf("directory write error: %w", err)
	}

	// KEY OPTIMIZATION: streaming copy only needed file data
	if err := copyNeededFileDataStreaming(sourceFile, tempFile, vaultDir.Entries, &header); err != nil {
		return fmt.Errorf("file data streaming error: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("temp file sync error: %w", err)
	}

	tempFile.Close()
	sourceFile.Close()

	// Atomic file replacement
	if err := os.Rename(tempPath, vaultPath); err != nil {
		return fmt.Errorf("file replacement error: %w", err)
	}

	// Clear sensitive data
	for i := range key {
		key[i] = 0
	}

	return nil
}

// copyNeededFileDataStreaming streams copy only needed file data
func copyNeededFileDataStreaming(sourceFile, targetFile *os.File, entries []FileEntry, originalHeader *VaultHeader) error {
	// Calculate original data offset
	headerSize := int64(binary.Size(VaultHeader{}))
	originalDirectorySize := int64(originalHeader.DirectorySize)
	originalDataOffset := headerSize + originalDirectorySize

	// Stream copy each needed file
	buffer := make([]byte, StreamBufferSize)

	for _, entry := range entries {
		if entry.IsDir {
			continue // Skip directories
		}

		// Seek to file position in source
		sourceOffset := originalDataOffset + entry.Offset
		if _, err := sourceFile.Seek(sourceOffset, io.SeekStart); err != nil {
			return fmt.Errorf("source seek error for %s: %w", entry.Path, err)
		}

		// Copy compressed file data
		limitedReader := io.LimitReader(sourceFile, entry.CompressedSize)
		if _, err := io.CopyBuffer(targetFile, limitedReader, buffer); err != nil {
			return fmt.Errorf("file data copy error for %s: %w", entry.Path, err)
		}
	}

	return nil
}

// ========================
// COMPRESSION FUNCTIONS
// ========================

// compressData compresses data using gzip compression
func compressData(data []byte) ([]byte, error) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return compressed.Bytes(), nil
}

// decompressData decompresses gzip-compressed data
func decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// decompressDataStreaming creates streaming gzip reader for decompression
func decompressDataStreaming(reader io.Reader) (*gzip.Reader, error) {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("gzip reader creation error: %w", err)
	}
	return gzipReader, nil
}

// getOptimalBufferSizeForFile returns optimal buffer size for file
func getOptimalBufferSizeForFile(fileSize int64) int {
	switch {
	case fileSize < 1024*1024: // < 1MB
		return 64 * 1024 // 64KB buffer
	case fileSize < 100*1024*1024: // < 100MB
		return 1024 * 1024 // 1MB buffer
	default: // > 100MB
		return 4 * 1024 * 1024 // 4MB buffer
	}
}

// streamCopyWithIntegrityCheck performs streaming copy with integrity check
func streamCopyWithIntegrityCheck(outputFile *os.File, source io.Reader, entry FileEntry, bufferSize int) error {
	// Create hasher for integrity check
	hasher := sha256.New()

	// MultiWriter writes simultaneously to file and hasher
	multiWriter := io.MultiWriter(outputFile, hasher)

	// Use controlled buffer for streaming copy
	buffer := make([]byte, bufferSize)

	// Streaming copy with optimal buffer
	_, err := io.CopyBuffer(multiWriter, source, buffer)
	if err != nil {
		return fmt.Errorf("streaming copy error: %w", err)
	}

	// Check integrity after copy
	actualHash := hasher.Sum(nil)
	var expectedHash [32]byte = entry.SHA256Hash

	if !compareHashesConstantTime(actualHash, expectedHash[:]) {
		return fmt.Errorf("integrity check failed: file data corrupted")
	}

	// Set permissions and modification time
	if err := outputFile.Chmod(os.FileMode(entry.Mode)); err != nil {
		return fmt.Errorf("permission set error: %w", err)
	}

	if err := os.Chtimes(outputFile.Name(), entry.ModTime, entry.ModTime); err != nil {
		return fmt.Errorf("time set error: %w", err)
	}

	return nil
}

// compareHashesConstantTime compares hashes with constant time execution
func compareHashesConstantTime(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	result := byte(0)
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// ========================
// EXTRACTION FUNCTIONS
// ========================

// extractFileEntry extracts a single file entry from vault using STREAMING processing
func extractFileEntry(vaultPath, password string, entry FileEntry, outputDir string) error {
	outputPath := filepath.Join(outputDir, entry.Path)

	if entry.IsDir {
		// Create directory
		return os.MkdirAll(outputPath, os.FileMode(entry.Mode))
	}

	// Create parent directories
	parentDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("parent directory creation error: %w", err)
	}

	// Open vault file
	vaultFile, err := os.Open(vaultPath)
	if err != nil {
		return fmt.Errorf("vault file open error: %w", err)
	}
	defer vaultFile.Close()

	// Read vault header to calculate file data offset
	var header VaultHeader
	if err := binary.Read(vaultFile, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("header read error: %w", err)
	}

	// Calculate absolute offset in vault file
	headerSize := int64(binary.Size(VaultHeader{}))
	directorySize := int64(header.DirectorySize)
	absoluteOffset := headerSize + directorySize + entry.Offset

	// Seek to file data
	if _, err := vaultFile.Seek(absoluteOffset, io.SeekStart); err != nil {
		return fmt.Errorf("file data seek error: %w", err)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("output file creation error: %w", err)
	}
	defer outputFile.Close()

	// CRITICAL OPTIMIZATION: streaming processing instead of loading to memory
	// Read compressed data in chunks, not loading all to memory
	limitedReader := io.LimitReader(vaultFile, entry.CompressedSize)

	// Create gzip reader for streaming decompression
	gzipReader, err := decompressDataStreaming(limitedReader)
	if err != nil {
		return fmt.Errorf("streaming decompression setup error: %w", err)
	}
	defer gzipReader.Close()

	// Streaming copy with integrity check and optimal buffer
	bufferSize := getOptimalBufferSizeForFile(entry.Size)
	return streamCopyWithIntegrityCheck(outputFile, gzipReader, entry, bufferSize)
}

// RemoveFromVault removes files/directories from vault
func RemoveFromVault(vaultPath, password string, paths []string) error {
	// Validate inputs
	if vaultPath == "" {
		return fmt.Errorf("vault path cannot be empty")
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if len(paths) == 0 {
		return fmt.Errorf("no paths specified for removal")
	}

	// Synchronize vault access for thread safety
	vaultMutex := getVaultMutex(vaultPath)
	vaultMutex.Lock()
	defer vaultMutex.Unlock()

	// Load vault directory
	vaultDir, err := loadVaultDirectory(vaultPath, password)
	if err != nil {
		return fmt.Errorf("vault directory load error: %w", err)
	}

	// Track which entries to keep
	var entriesToKeep []FileEntry
	removedPaths := make(map[string]bool)

	// Mark paths for removal (normalize first)
	for _, path := range paths {
		cleanPath := filepath.Clean(path)
		removedPaths[cleanPath] = true
	}

	// Filter entries to keep
	for _, entry := range vaultDir.Entries {
		if !removedPaths[entry.Path] {
			entriesToKeep = append(entriesToKeep, entry)
		}
	}

	// Check if any files were actually removed
	if len(entriesToKeep) == len(vaultDir.Entries) {
		return fmt.Errorf("no matching files found for removal")
	}

	// Update directory with remaining entries
	vaultDir.Entries = entriesToKeep

	// Update vault with optimized streaming approach
	return updateVaultDirectoryStreamingOptimized(vaultPath, password, *vaultDir)
}
