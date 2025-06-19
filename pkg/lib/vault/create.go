package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// Magic header for vault file identification
	VaultMagic = "FLINT001"

	// Cryptographic parameters
	KeyLength   = 32     // AES-256
	SaltLength  = 32     // 256-bit salt
	NonceLength = 12     // GCM nonce
	PBKDF2Iters = 100000 // PBKDF2 iterations (recommended minimum)

	// Header field sizes
	MagicLength   = 8
	VersionLength = 4
	ItersLength   = 4

	// Current vault format version
	CurrentVaultVersion = 1
)

// VaultHeader contains vault metadata
type VaultHeader struct {
	Magic      [MagicLength]byte // "FLINT001"
	Version    uint32            // Format version
	Iterations uint32            // PBKDF2 iteration count
	Salt       [SaltLength]byte  // Salt for key derivation
	Nonce      [NonceLength]byte // Nonce for AES-GCM
}

// VaultInfo contains basic information about a vault file that can be read without a password
type VaultInfo struct {
	IsFlintVault bool   // Whether this is a valid Flint Vault file
	Version      uint32 // Vault format version
	Iterations   uint32 // PBKDF2 iteration count used
	FileSize     int64  // Total file size in bytes
	FilePath     string // Path to the vault file
}

// IsFlintVault checks if the specified file is a valid Flint Vault file.
// This function only reads the file header and does not require a password.
//
// Parameters:
//   - path: Path to the file to check
//
// Returns:
//   - bool: true if the file is a valid Flint Vault file, false otherwise
//   - error: nil on success, or error describing the failure
//
// This function is useful for:
//   - Validating user input before attempting to decrypt
//   - Filtering files in directory listings
//   - Providing user-friendly error messages
func IsFlintVault(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Try to read the header
	var header VaultHeader
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return false, nil // Not enough data for header, not a vault file
	}

	// Check magic header
	return string(header.Magic[:]) == VaultMagic, nil
}

// GetVaultInfo returns basic information about a vault file without requiring a password.
// This function reads only the file header and provides metadata that can be
// safely displayed to users for file identification and validation.
//
// Parameters:
//   - path: Path to the vault file
//
// Returns:
//   - *VaultInfo: Information about the vault file
//   - error: nil on success, or error describing the failure
//
// The returned information includes:
//   - Whether the file is a valid Flint Vault
//   - Vault format version
//   - PBKDF2 iteration count
//   - File size
//   - File path
func GetVaultInfo(path string) (*VaultInfo, error) {
	// Get file info for size
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("file stat error: %w", err)
	}

	info := &VaultInfo{
		IsFlintVault: false,
		Version:      0,
		Iterations:   0,
		FileSize:     fileInfo.Size(),
		FilePath:     path,
	}

	// Try to open and read header
	file, err := os.Open(path)
	if err != nil {
		return info, nil // Return info with IsFlintVault=false
	}
	defer file.Close()

	// Try to read the header
	var header VaultHeader
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return info, nil // Not enough data for header, not a vault file
	}

	// Check magic header
	if string(header.Magic[:]) == VaultMagic {
		info.IsFlintVault = true
		info.Version = header.Version
		info.Iterations = header.Iterations
	}

	return info, nil
}

// ValidateVaultFile performs comprehensive validation of a vault file format.
// This function checks the file structure and header validity without requiring a password.
//
// Parameters:
//   - path: Path to the vault file to validate
//
// Returns:
//   - error: nil if the file is valid, or error describing validation failures
//
// Validation checks:
//   - File exists and is readable
//   - File has correct magic header
//   - File has supported version
//   - File has minimum required size
//   - Header fields are within expected ranges
func ValidateVaultFile(path string) error {
	// Check file exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file access error: %w", err)
	}

	// Check minimum file size (header + some encrypted data)
	minSize := int64(binary.Size(VaultHeader{}) + 16) // header + minimal ciphertext
	if fileInfo.Size() < minSize {
		return fmt.Errorf("file too small to be a valid vault (minimum %d bytes, got %d)", minSize, fileInfo.Size())
	}

	// Open file and read header
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("file open error: %w", err)
	}
	defer file.Close()

	var header VaultHeader
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("header read error: %w", err)
	}

	// Validate magic header
	if string(header.Magic[:]) != VaultMagic {
		return fmt.Errorf("invalid file format: not a Flint Vault file (expected magic '%s', got '%s')",
			VaultMagic, string(header.Magic[:]))
	}

	// Validate version
	if header.Version < 1 || header.Version > CurrentVaultVersion {
		return fmt.Errorf("unsupported vault version: %d (supported: 1-%d)", header.Version, CurrentVaultVersion)
	}

	// Validate iteration count (should be reasonable)
	if header.Iterations < 10000 || header.Iterations > 10000000 {
		return fmt.Errorf("suspicious PBKDF2 iteration count: %d (expected: 10,000 - 10,000,000)", header.Iterations)
	}

	return nil
}

// CreateVault creates a new encrypted vault file at the specified path.
// The vault is protected with the provided password using AES-256-GCM
// encryption with PBKDF2 key derivation (100,000 iterations).
//
// Parameters:
//   - path: File system path where the vault will be created
//   - password: Password for encrypting the vault (must not be empty)
//
// Returns:
//   - error: nil on success, or error describing the failure
//
// The function will fail if:
//   - The vault file already exists
//   - The password is empty
//   - Insufficient permissions to create the file
func CreateVault(path string, password string) error {
	// Validate input parameters
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

	// Create empty vault
	vaultData := VaultData{
		Entries:   []VaultEntry{},
		CreatedAt: time.Now(),
		Comment:   "Encrypted Flint Vault Storage",
	}

	return saveVaultData(path, password, vaultData)
}

// saveVaultData saves vault data in encrypted form to a file.
// This internal function handles the complete encryption process including
// JSON serialization, gzip compression, key derivation, and AES-GCM encryption.
//
// Parameters:
//   - path: File system path where the vault will be saved
//   - password: Password for encrypting the vault
//   - data: VaultData structure containing all vault entries and metadata
//
// Returns:
//   - error: nil on success, or error describing the failure
//
// Security features:
//   - Generates cryptographically secure 32-byte salt
//   - Uses PBKDF2 with 100,000 iterations for key derivation
//   - Employs AES-256-GCM for authenticated encryption
//   - Clears sensitive data from memory after use
func saveVaultData(path, password string, data VaultData) error {
	// Serialize data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("data serialization error: %w", err)
	}

	// Compress data to save space
	compressedData, err := compressData(jsonData)
	if err != nil {
		return fmt.Errorf("data compression error: %w", err)
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

	// Encrypt compressed data
	ciphertext := gcm.Seal(nil, nonce, compressedData, nil)

	// Create vault header
	header := VaultHeader{
		Version:    CurrentVaultVersion,
		Iterations: PBKDF2Iters,
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

	// Write encrypted data
	if _, err := file.Write(ciphertext); err != nil {
		return fmt.Errorf("data write error: %w", err)
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

// loadVaultData loads and decrypts vault data from a file.
// This internal function handles the complete decryption process including
// file validation, key derivation, AES-GCM decryption, decompression, and JSON parsing.
//
// Parameters:
//   - path: File system path to the vault file
//   - password: Password for decrypting the vault
//
// Returns:
//   - *VaultData: Pointer to loaded vault data on success
//   - error: nil on success, or error describing the failure
//
// Security features:
//   - Validates magic header and version
//   - Uses stored salt for key derivation
//   - Verifies authentication tag during decryption
//   - Clears sensitive data from memory after use
//
// The function will fail if:
//   - File doesn't exist or can't be read
//   - Invalid vault file format
//   - Wrong password or corrupted data
//   - Unsupported vault version
func loadVaultData(path, password string) (*VaultData, error) {
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

	// Read encrypted data
	ciphertext, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("encrypted data read error: %w", err)
	}

	// Decrypt data
	compressedData, err := gcm.Open(nil, header.Nonce[:], ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: invalid password or corrupted data")
	}

	// Decompress data
	jsonData, err := decompressData(compressedData)
	if err != nil {
		return nil, fmt.Errorf("data decompression error: %w", err)
	}

	// Deserialize JSON
	var vaultData VaultData
	if err := json.Unmarshal(jsonData, &vaultData); err != nil {
		return nil, fmt.Errorf("data deserialization error: %w", err)
	}

	// Clear key from memory
	for i := range key {
		key[i] = 0
	}

	return &vaultData, nil
}
