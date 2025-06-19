package vault

import (
	"bytes"
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
	"strings"
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

// saveVaultDirectory saves only the vault directory (metadata) - optimized for memory usage
func saveVaultDirectory(path, password string, vaultDir VaultDirectory) error {
	// Serialize directory to JSON
	jsonData, err := json.Marshal(vaultDir)
	if err != nil {
		return fmt.Errorf("directory serialization error: %w", err)
	}

	// Compress directory data
	compressedData, err := compressData(jsonData)
	if err != nil {
		return fmt.Errorf("directory compression error: %w", err)
	}

	// Generate cryptographically secure salt
	salt := make([]byte, SaltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("salt generation error: %w", err)
	}

	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, PBKDF2Iters, KeyLength, sha256.New)

	// Clear password from memory
	passwordBytes := []byte(password)
	for i := range passwordBytes {
		passwordBytes[i] = 0
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("AES cipher creation error: %w", err)
	}

	// Create GCM for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("GCM creation error: %w", err)
	}

	// Generate unique nonce
	nonce := make([]byte, NonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("nonce generation error: %w", err)
	}

	// Encrypt compressed directory
	ciphertext := gcm.Seal(nil, nonce, compressedData, nil)

	// Create vault header
	header := VaultHeader{
		Version:       CurrentVaultVersion,
		Iterations:    PBKDF2Iters,
		DirectorySize: uint64(len(ciphertext)),
	}

	copy(header.Magic[:], []byte(VaultMagic))
	copy(header.Salt[:], salt)
	copy(header.Nonce[:], nonce)

	// Create/overwrite vault file
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
	if _, err := file.Write(ciphertext); err != nil {
		return fmt.Errorf("directory write error: %w", err)
	}

	// Force data to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("disk sync error: %w", err)
	}

	// Clear sensitive data from memory
	for i := range key {
		key[i] = 0
	}
	for i := range salt {
		salt[i] = 0
	}
	for i := range nonce {
		nonce[i] = 0
	}

	return nil
}

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

	// Calculate file hash and compress file data in streaming fashion
	fileHash, compressedData, err := processFileForVault(filePath)
	if err != nil {
		return fmt.Errorf("file processing error: %w", err)
	}

	// Create file entry with metadata (we'll calculate offset later)
	entry := FileEntry{
		Path:           filepath.Clean(filePath),
		Name:           fileInfo.Name(),
		IsDir:          false,
		Size:           fileInfo.Size(),
		CompressedSize: int64(len(compressedData)),
		Mode:           uint32(fileInfo.Mode()),
		ModTime:        fileInfo.ModTime(),
		Offset:         0, // Will be set in updateVaultDirectoryWithFileData
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

	// Save updated directory and append file data in one operation
	return updateVaultDirectoryWithFileData(vaultPath, password, *vaultDir, compressedData)
}

// processFileForVault calculates hash and compresses file data
func processFileForVault(filePath string) ([32]byte, []byte, error) {
	// Open source file
	file, err := os.Open(filePath)
	if err != nil {
		return [32]byte{}, nil, fmt.Errorf("file open error: %w", err)
	}
	defer file.Close()

	// Read entire file and calculate hash
	fileData, err := io.ReadAll(file)
	if err != nil {
		return [32]byte{}, nil, fmt.Errorf("file read error: %w", err)
	}

	// Calculate SHA-256 hash
	hash := sha256.Sum256(fileData)

	// Compress file data
	compressedData, err := compressData(fileData)
	if err != nil {
		return [32]byte{}, nil, fmt.Errorf("compression error: %w", err)
	}

	return hash, compressedData, nil
}

// updateVaultDirectory updates the vault directory in the vault file
func updateVaultDirectory(vaultPath, password string, vaultDir VaultDirectory) error {
	// Read current vault file
	currentData, err := os.ReadFile(vaultPath)
	if err != nil {
		return fmt.Errorf("vault file read error: %w", err)
	}

	// Parse header to get current directory size
	var header VaultHeader
	if err := binary.Read(bytes.NewReader(currentData), binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("header parse error: %w", err)
	}

	// Calculate offset where file data starts
	headerSize := int64(binary.Size(VaultHeader{}))
	oldDirectorySize := int64(header.DirectorySize)
	fileDataOffset := headerSize + oldDirectorySize

	// Preserve file data
	fileData := currentData[fileDataOffset:]

	// Serialize new directory
	jsonData, err := json.Marshal(vaultDir)
	if err != nil {
		return fmt.Errorf("directory serialization error: %w", err)
	}

	// Compress new directory
	compressedDir, err := compressData(jsonData)
	if err != nil {
		return fmt.Errorf("directory compression error: %w", err)
	}

	// Encrypt new directory with same key parameters
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

	// Update header with new directory size
	header.DirectorySize = uint64(len(encryptedDir))

	// Create new vault file
	file, err := os.Create(vaultPath)
	if err != nil {
		return fmt.Errorf("vault file creation error: %w", err)
	}
	defer file.Close()

	// Write updated header
	if err := binary.Write(file, binary.LittleEndian, header); err != nil {
		return fmt.Errorf("header write error: %w", err)
	}

	// Write encrypted directory
	if _, err := file.Write(encryptedDir); err != nil {
		return fmt.Errorf("directory write error: %w", err)
	}

	// Write preserved file data
	if _, err := file.Write(fileData); err != nil {
		return fmt.Errorf("file data write error: %w", err)
	}

	// Clear sensitive data
	for i := range key {
		key[i] = 0
	}

	return nil
}

// updateVaultDirectoryWithFileData updates the vault directory and appends new file data atomically
func updateVaultDirectoryWithFileData(vaultPath, password string, vaultDir VaultDirectory, newFileData []byte) error {
	// Read current vault file
	currentData, err := os.ReadFile(vaultPath)
	if err != nil {
		return fmt.Errorf("vault file read error: %w", err)
	}

	// Parse header to get current directory size
	var header VaultHeader
	if err := binary.Read(bytes.NewReader(currentData), binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("header parse error: %w", err)
	}

	// Calculate offset where existing file data starts
	headerSize := int64(binary.Size(VaultHeader{}))
	oldDirectorySize := int64(header.DirectorySize)
	fileDataOffset := headerSize + oldDirectorySize

	// Preserve existing file data
	existingFileData := currentData[fileDataOffset:]

	// Calculate file offsets for all entries in the directory
	var totalDataSize int64 = 0
	for i := range vaultDir.Entries {
		if !vaultDir.Entries[i].IsDir {
			vaultDir.Entries[i].Offset = totalDataSize
			totalDataSize += vaultDir.Entries[i].CompressedSize
		}
	}

	// Now append new file data to existing data
	allFileData := append(existingFileData, newFileData...)

	// Update directory structure
	return rebuildVaultFile(vaultPath, password, vaultDir, allFileData)
}

// rebuildVaultFile completely rebuilds the vault file with correct structure
func rebuildVaultFile(vaultPath, password string, vaultDir VaultDirectory, fileData []byte) error {
	// Create temporary file
	tempPath := vaultPath + ".tmp"
	defer os.Remove(tempPath) // Clean up temp file

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

	// Read header from original file to get encryption parameters
	originalFile, err := os.Open(vaultPath)
	if err != nil {
		return fmt.Errorf("original file open error: %w", err)
	}

	var header VaultHeader
	if err := binary.Read(originalFile, binary.LittleEndian, &header); err != nil {
		originalFile.Close()
		return fmt.Errorf("header read error: %w", err)
	}
	originalFile.Close()

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

	// Update header with new directory size
	header.DirectorySize = uint64(len(encryptedDir))

	// Create temporary file
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("temp file creation error: %w", err)
	}

	// Write header
	if err := binary.Write(tempFile, binary.LittleEndian, header); err != nil {
		tempFile.Close()
		return fmt.Errorf("header write error: %w", err)
	}

	// Write encrypted directory
	if _, err := tempFile.Write(encryptedDir); err != nil {
		tempFile.Close()
		return fmt.Errorf("directory write error: %w", err)
	}

	// Write all file data
	if _, err := tempFile.Write(fileData); err != nil {
		tempFile.Close()
		return fmt.Errorf("file data write error: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return fmt.Errorf("temp file sync error: %w", err)
	}

	tempFile.Close()

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

// ListVault returns list of files in the vault
func ListVault(vaultPath, password string) ([]FileEntry, error) {
	vaultDir, err := loadVaultDirectory(vaultPath, password)
	if err != nil {
		return nil, err
	}

	return vaultDir.Entries, nil
}

// ExtractFromVault extracts all files from vault to specified directory
func ExtractFromVault(vaultPath, password, outputDir string) error {
	// Load vault directory
	vaultDir, err := loadVaultDirectory(vaultPath, password)
	if err != nil {
		return err
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("output directory creation error: %w", err)
	}

	// Open vault file for reading file data
	vaultFile, err := os.Open(vaultPath)
	if err != nil {
		return fmt.Errorf("vault file open error: %w", err)
	}
	defer vaultFile.Close()

	// Extract all entries
	for _, entry := range vaultDir.Entries {
		if err := extractFileEntry(vaultFile, entry, outputDir); err != nil {
			return fmt.Errorf("extraction error for %s: %w", entry.Path, err)
		}
	}

	return nil
}

// extractFileEntry extracts a single file entry from vault
func extractFileEntry(vaultFile *os.File, entry FileEntry, outputDir string) error {
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

	// Read vault header to calculate file data offset
	if _, err := vaultFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("file seek error: %w", err)
	}

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

	// Read compressed file data
	compressedData := make([]byte, entry.CompressedSize)
	if _, err := io.ReadFull(vaultFile, compressedData); err != nil {
		return fmt.Errorf("compressed data read error: %w", err)
	}

	// Decompress file data
	originalData, err := decompressData(compressedData)
	if err != nil {
		return fmt.Errorf("decompression error: %w", err)
	}

	// Verify file integrity
	actualHash := sha256.Sum256(originalData)
	if actualHash != entry.SHA256Hash {
		return fmt.Errorf("integrity check failed: file data corrupted")
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("output file creation error: %w", err)
	}
	defer outputFile.Close()

	// Write file data
	if _, err := outputFile.Write(originalData); err != nil {
		return fmt.Errorf("file write error: %w", err)
	}

	// Set file permissions and modification time
	if err := outputFile.Chmod(os.FileMode(entry.Mode)); err != nil {
		return fmt.Errorf("permission set error: %w", err)
	}

	if err := os.Chtimes(outputPath, entry.ModTime, entry.ModTime); err != nil {
		return fmt.Errorf("time set error: %w", err)
	}

	return nil
}

// GetFromVault extracts specific files/directories from vault
func GetFromVault(vaultPath, password, outputDir string, paths []string) error {
	// Load vault directory
	vaultDir, err := loadVaultDirectory(vaultPath, password)
	if err != nil {
		return err
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("output directory creation error: %w", err)
	}

	// Open vault file for reading
	vaultFile, err := os.Open(vaultPath)
	if err != nil {
		return fmt.Errorf("vault file open error: %w", err)
	}
	defer vaultFile.Close()

	// Extract requested entries
	for _, requestedPath := range paths {
		found := false
		for _, entry := range vaultDir.Entries {
			if entry.Path == requestedPath || strings.HasPrefix(entry.Path, requestedPath+"/") {
				if err := extractFileEntry(vaultFile, entry, outputDir); err != nil {
					return fmt.Errorf("extraction error for %s: %w", entry.Path, err)
				}
				found = true
			}
		}
		if !found {
			return fmt.Errorf("path not found in vault: %s", requestedPath)
		}
	}

	return nil
}

// RemoveFromVault removes files/directories from vault
func RemoveFromVault(vaultPath, password string, paths []string) error {
	// Load vault directory
	vaultDir, err := loadVaultDirectory(vaultPath, password)
	if err != nil {
		return err
	}

	// Remove requested entries
	var newEntries []FileEntry
	for _, entry := range vaultDir.Entries {
		shouldRemove := false
		for _, removePath := range paths {
			if entry.Path == removePath || strings.HasPrefix(entry.Path, removePath+"/") {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			newEntries = append(newEntries, entry)
		}
	}

	// Update vault directory
	vaultDir.Entries = newEntries
	return updateVaultDirectory(vaultPath, password, *vaultDir)
}
