package vault

import (
	"encoding/binary"
	"fmt"
	"os"
)

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
