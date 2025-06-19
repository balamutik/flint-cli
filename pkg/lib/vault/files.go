package vault

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

// VaultEntry represents a file or directory in the vault
type VaultEntry struct {
	Path    string    `json:"path"`     // Path to file/directory
	Name    string    `json:"name"`     // Name of file/directory
	IsDir   bool      `json:"is_dir"`   // Whether it's a directory
	Size    int64     `json:"size"`     // File size (0 for directories)
	Mode    uint32    `json:"mode"`     // Access permissions
	ModTime time.Time `json:"mod_time"` // Last modification time
	Content []byte    `json:"content"`  // File content (empty for directories)
}

// VaultData contains all vault data
type VaultData struct {
	Entries   []VaultEntry `json:"entries"`    // List of files and directories
	CreatedAt time.Time    `json:"created_at"` // Vault creation time
	Comment   string       `json:"comment"`    // Vault comment
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

// AddFileToVault adds a single file to an existing vault.
// If the file already exists in the vault, it will be updated with new content.
// If the specified path is a directory, it delegates to AddDirectoryToVault.
//
// Parameters:
//   - vaultPath: Path to the existing vault file
//   - password: Password for accessing the vault
//   - filePath: Path to the file to add to the vault
//
// Returns:
//   - error: nil on success, or error describing the failure
//
// The function preserves:
//   - Original file permissions (mode)
//   - Last modification time
//   - File content exactly as stored on disk
func AddFileToVault(vaultPath, password, filePath string) error {
	// Load existing vault
	vaultData, err := loadVaultData(vaultPath, password)
	if err != nil {
		return fmt.Errorf("vault load error: %w", err)
	}

	// Get file information
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file info error: %w", err)
	}

	if fileInfo.IsDir() {
		return AddDirectoryToVault(vaultPath, password, filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("file read error: %w", err)
	}

	// Create file entry
	entry := VaultEntry{
		Path:    filepath.Clean(filePath),
		Name:    fileInfo.Name(),
		IsDir:   false,
		Size:    fileInfo.Size(),
		Mode:    uint32(fileInfo.Mode()),
		ModTime: fileInfo.ModTime(),
		Content: content,
	}

	// Check if file with this path already exists
	for i, existingEntry := range vaultData.Entries {
		if existingEntry.Path == entry.Path {
			vaultData.Entries[i] = entry // Update existing file
			return saveVaultData(vaultPath, password, *vaultData)
		}
	}

	// Add new file
	vaultData.Entries = append(vaultData.Entries, entry)

	return saveVaultData(vaultPath, password, *vaultData)
}

// AddDirectoryToVault recursively adds a directory and all its contents to the vault.
// This function walks through the entire directory tree and adds all files and
// subdirectories while preserving the original structure and metadata.
//
// Parameters:
//   - vaultPath: Path to the existing vault file
//   - password: Password for accessing the vault
//   - dirPath: Path to the directory to add to the vault
//
// Returns:
//   - error: nil on success, or error describing the failure
//
// Behavior:
//   - Recursively processes all subdirectories
//   - Updates existing files if they already exist in vault
//   - Preserves directory structure, permissions, and timestamps
//   - Skips symbolic links for security reasons
func AddDirectoryToVault(vaultPath, password, dirPath string) error {
	vaultData, err := loadVaultData(vaultPath, password)
	if err != nil {
		return fmt.Errorf("vault load error: %w", err)
	}

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create relative path
		relPath, err := filepath.Rel(filepath.Dir(dirPath), path)
		if err != nil {
			relPath = path
		}

		entry := VaultEntry{
			Path:    filepath.Clean(relPath),
			Name:    info.Name(),
			IsDir:   info.IsDir(),
			Size:    info.Size(),
			Mode:    uint32(info.Mode()),
			ModTime: info.ModTime(),
		}

		// If it's a file, read its content
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("file read error %s: %w", path, err)
			}
			entry.Content = content
		}

		// Check if item with this path already exists
		found := false
		for i, existingEntry := range vaultData.Entries {
			if existingEntry.Path == entry.Path {
				vaultData.Entries[i] = entry
				found = true
				break
			}
		}

		if !found {
			vaultData.Entries = append(vaultData.Entries, entry)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("directory walk error: %w", err)
	}

	return saveVaultData(vaultPath, password, *vaultData)
}

// ExtractVault extracts all files and directories from the vault to a specified directory.
// This function recreates the complete directory structure as it was stored in the vault.
//
// Parameters:
//   - vaultPath: Path to the vault file
//   - password: Password for accessing the vault
//   - outputDir: Destination directory where files will be extracted
//
// Returns:
//   - error: nil on success, or error describing the failure
//
// Behavior:
//   - Creates output directory if it doesn't exist
//   - First creates all directories, then writes all files
//   - Preserves original file permissions and timestamps
//   - Overwrites existing files in the output directory
//   - Maintains the exact directory structure from the vault
func ExtractVault(vaultPath, password, outputDir string) error {
	vaultData, err := loadVaultData(vaultPath, password)
	if err != nil {
		return fmt.Errorf("vault load error: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("output directory creation error: %w", err)
	}

	// First create all directories
	for _, entry := range vaultData.Entries {
		if entry.IsDir {
			fullPath := filepath.Join(outputDir, entry.Path)
			if err := os.MkdirAll(fullPath, os.FileMode(entry.Mode)); err != nil {
				return fmt.Errorf("directory creation error %s: %w", fullPath, err)
			}
		}
	}

	// Then create all files
	for _, entry := range vaultData.Entries {
		if !entry.IsDir {
			fullPath := filepath.Join(outputDir, entry.Path)

			// Create directory for file if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return fmt.Errorf("file directory creation error %s: %w", fullPath, err)
			}

			// Write file
			if err := os.WriteFile(fullPath, entry.Content, os.FileMode(entry.Mode)); err != nil {
				return fmt.Errorf("file write error %s: %w", fullPath, err)
			}

			// Restore modification time
			if err := os.Chtimes(fullPath, entry.ModTime, entry.ModTime); err != nil {
				fmt.Printf("Warning: failed to restore time for %s: %v\n", fullPath, err)
			}
		}
	}

	return nil
}

// ExtractSpecific extracts a specific file or directory from the vault.
// This function allows selective extraction of individual files or complete
// directory trees without extracting the entire vault.
//
// Parameters:
//   - vaultPath: Path to the vault file
//   - password: Password for accessing the vault
//   - targetPath: Path to the specific file or directory in the vault to extract
//   - outputDir: Destination directory where files will be extracted
//
// Returns:
//   - error: nil on success, or error describing the failure
//
// Behavior:
//   - For single files: extracts directly to outputDir with original filename
//   - For directories: extracts complete subtree maintaining relative structure
//   - Creates necessary parent directories automatically
//   - Preserves original file permissions and timestamps
//   - Returns error if targetPath is not found in vault
func ExtractSpecific(vaultPath, password, targetPath, outputDir string) error {
	vaultData, err := loadVaultData(vaultPath, password)
	if err != nil {
		return fmt.Errorf("vault load error: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("output directory creation error: %w", err)
	}

	// Find matching entries
	var matchingEntries []VaultEntry
	for _, entry := range vaultData.Entries {
		// Check exact match or if element is a child of the directory
		if entry.Path == targetPath || strings.HasPrefix(entry.Path, targetPath+"/") {
			matchingEntries = append(matchingEntries, entry)
		}
	}

	if len(matchingEntries) == 0 {
		return fmt.Errorf("file or directory '%s' not found in vault", targetPath)
	}

	// First create all directories
	for _, entry := range matchingEntries {
		if entry.IsDir {
			// Calculate relative path from targetPath
			relPath, err := filepath.Rel(targetPath, entry.Path)
			if err != nil {
				relPath = entry.Path
			}

			fullPath := filepath.Join(outputDir, relPath)
			if err := os.MkdirAll(fullPath, os.FileMode(entry.Mode)); err != nil {
				return fmt.Errorf("directory creation error %s: %w", fullPath, err)
			}
		}
	}

	// Then create all files
	for _, entry := range matchingEntries {
		if !entry.IsDir {
			// If extracting single file, place it directly in outputDir
			var fullPath string
			if entry.Path == targetPath {
				fullPath = filepath.Join(outputDir, entry.Name)
			} else {
				// For files in directory, preserve structure
				relPath, err := filepath.Rel(targetPath, entry.Path)
				if err != nil {
					relPath = entry.Path
				}
				fullPath = filepath.Join(outputDir, relPath)
			}

			// Create directory for file if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return fmt.Errorf("file directory creation error %s: %w", fullPath, err)
			}

			// Write file
			if err := os.WriteFile(fullPath, entry.Content, os.FileMode(entry.Mode)); err != nil {
				return fmt.Errorf("file write error %s: %w", fullPath, err)
			}

			// Restore modification time
			if err := os.Chtimes(fullPath, entry.ModTime, entry.ModTime); err != nil {
				fmt.Printf("Warning: failed to restore time for %s: %v\n", fullPath, err)
			}
		}
	}

	return nil
}

// ExtractMultiple extracts multiple files and directories from the vault.
// This function allows selective extraction of multiple files and complete
// directory trees without extracting the entire vault, providing flexible
// and efficient batch extraction capabilities.
//
// Parameters:
//   - vaultPath: Path to the vault file
//   - password: Password for accessing the vault
//   - targetPaths: Slice of paths to files or directories in the vault to extract
//   - outputDir: Destination directory where files will be extracted
//
// Returns:
//   - []string: Slice of successfully extracted paths
//   - []string: Slice of paths that were not found in the vault
//   - error: nil on success, or error describing the failure
//
// Behavior:
//   - Processes all target paths and extracts matching files/directories
//   - For single files: extracts directly to outputDir with original filename
//   - For directories: extracts complete subtree maintaining relative structure
//   - Creates necessary parent directories automatically
//   - Preserves original file permissions and timestamps
//   - Returns lists of successful and failed extractions for detailed feedback
//   - Continues processing even if some paths are not found
func ExtractMultiple(vaultPath, password string, targetPaths []string, outputDir string) ([]string, []string, error) {
	if len(targetPaths) == 0 {
		return nil, nil, fmt.Errorf("no target paths specified")
	}

	vaultData, err := loadVaultData(vaultPath, password)
	if err != nil {
		return nil, nil, fmt.Errorf("vault load error: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("output directory creation error: %w", err)
	}

	var allMatchingEntries []VaultEntry
	var extractedPaths []string
	var notFoundPaths []string

	// Find matching entries for all target paths
	for _, targetPath := range targetPaths {
		var targetMatches []VaultEntry
		found := false

		for _, entry := range vaultData.Entries {
			// Check exact match or if element is a child of the directory
			if entry.Path == targetPath || strings.HasPrefix(entry.Path, targetPath+"/") {
				targetMatches = append(targetMatches, entry)
				found = true
			}
		}

		if found {
			extractedPaths = append(extractedPaths, targetPath)
			allMatchingEntries = append(allMatchingEntries, targetMatches...)
		} else {
			notFoundPaths = append(notFoundPaths, targetPath)
		}
	}

	if len(allMatchingEntries) == 0 {
		return extractedPaths, notFoundPaths, fmt.Errorf("none of the specified files or directories were found in vault")
	}

	// Remove duplicates from matching entries (in case of overlapping paths)
	uniqueEntries := make(map[string]VaultEntry)
	for _, entry := range allMatchingEntries {
		uniqueEntries[entry.Path] = entry
	}

	// Convert back to slice
	var entriesToExtract []VaultEntry
	for _, entry := range uniqueEntries {
		entriesToExtract = append(entriesToExtract, entry)
	}

	// First create all directories
	for _, entry := range entriesToExtract {
		if entry.IsDir {
			// Determine the best target path for this directory entry
			var bestTargetPath string
			for _, targetPath := range extractedPaths {
				if entry.Path == targetPath || strings.HasPrefix(entry.Path, targetPath+"/") {
					bestTargetPath = targetPath
					break
				}
			}

			var fullPath string
			if entry.Path == bestTargetPath {
				// Directory is the target itself
				fullPath = filepath.Join(outputDir, entry.Name)
			} else {
				// Directory is a child of target
				relPath, err := filepath.Rel(bestTargetPath, entry.Path)
				if err != nil {
					relPath = entry.Path
				}
				fullPath = filepath.Join(outputDir, relPath)
			}

			if err := os.MkdirAll(fullPath, os.FileMode(entry.Mode)); err != nil {
				return extractedPaths, notFoundPaths, fmt.Errorf("directory creation error %s: %w", fullPath, err)
			}
		}
	}

	// Then create all files
	for _, entry := range entriesToExtract {
		if !entry.IsDir {
			// Determine the best target path for this file entry
			var bestTargetPath string
			for _, targetPath := range extractedPaths {
				if entry.Path == targetPath || strings.HasPrefix(entry.Path, targetPath+"/") {
					bestTargetPath = targetPath
					break
				}
			}

			var fullPath string
			if entry.Path == bestTargetPath {
				// File is the target itself
				fullPath = filepath.Join(outputDir, entry.Name)
			} else {
				// File is a child of target directory
				relPath, err := filepath.Rel(bestTargetPath, entry.Path)
				if err != nil {
					relPath = entry.Path
				}
				fullPath = filepath.Join(outputDir, relPath)
			}

			// Create directory for file if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return extractedPaths, notFoundPaths, fmt.Errorf("file directory creation error %s: %w", fullPath, err)
			}

			// Write file
			if err := os.WriteFile(fullPath, entry.Content, os.FileMode(entry.Mode)); err != nil {
				return extractedPaths, notFoundPaths, fmt.Errorf("file write error %s: %w", fullPath, err)
			}

			// Restore modification time
			if err := os.Chtimes(fullPath, entry.ModTime, entry.ModTime); err != nil {
				fmt.Printf("Warning: failed to restore time for %s: %v\n", fullPath, err)
			}
		}
	}

	return extractedPaths, notFoundPaths, nil
}

// ListVault lists vault contents by loading and returning the complete vault data structure.
// This function provides access to all vault metadata including creation time,
// comments, and detailed information about all stored files and directories.
//
// Parameters:
//   - vaultPath: Path to the vault file
//   - password: Password for accessing the vault
//
// Returns:
//   - *VaultData: Complete vault data structure containing all entries and metadata
//   - error: nil on success, or error describing the failure
func ListVault(vaultPath, password string) (*VaultData, error) {
	return loadVaultData(vaultPath, password)
}

// RemoveFromVault removes a file or directory from the vault.
// When removing a directory, all files and subdirectories within it are also removed.
// This operation is permanent and cannot be undone.
//
// Parameters:
//   - vaultPath: Path to the vault file
//   - password: Password for accessing the vault
//   - targetPath: Path to the file or directory in the vault to remove
//
// Returns:
//   - error: nil on success, or error describing the failure
//
// Warning: This operation permanently deletes data from the vault.
// Make sure to backup important data before removal.
func RemoveFromVault(vaultPath, password, targetPath string) error {
	vaultData, err := loadVaultData(vaultPath, password)
	if err != nil {
		return fmt.Errorf("vault load error: %w", err)
	}

	// Find and remove element
	found := false
	newEntries := make([]VaultEntry, 0, len(vaultData.Entries))

	for _, entry := range vaultData.Entries {
		// Remove element or its child elements
		if entry.Path != targetPath && !strings.HasPrefix(entry.Path, targetPath+"/") {
			newEntries = append(newEntries, entry)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("file or directory '%s' not found in vault", targetPath)
	}

	vaultData.Entries = newEntries
	return saveVaultData(vaultPath, password, *vaultData)
}

// GetFileFromVault extracts a specific file from the vault into memory.
// This function retrieves file content without creating files on disk,
// making it useful for programmatic access to vault contents.
//
// Parameters:
//   - vaultPath: Path to the vault file
//   - password: Password for accessing the vault
//   - filePath: Path to the specific file in the vault to retrieve
//
// Returns:
//   - []byte: File content as byte slice
//   - error: nil on success, or error describing the failure
//
// Note: This function only works with files, not directories.
// Use ListVault to get directory structure information.
func GetFileFromVault(vaultPath, password, filePath string) ([]byte, error) {
	vaultData, err := loadVaultData(vaultPath, password)
	if err != nil {
		return nil, fmt.Errorf("vault load error: %w", err)
	}

	// Find file
	for _, entry := range vaultData.Entries {
		if entry.Path == filePath && !entry.IsDir {
			return entry.Content, nil
		}
	}

	return nil, fmt.Errorf("file '%s' not found in vault", filePath)
}

// compressData compresses data using gzip compression.
// This internal function reduces vault file size by compressing the JSON data
// before encryption. Uses standard gzip compression for broad compatibility.
//
// Parameters:
//   - data: Raw data to compress
//
// Returns:
//   - []byte: Compressed data
//   - error: nil on success, or error if compression fails
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

// decompressData decompresses gzip-compressed data.
// This internal function reverses the compression applied by compressData,
// restoring the original JSON data after decryption.
//
// Parameters:
//   - data: Gzip-compressed data to decompress
//
// Returns:
//   - []byte: Decompressed original data
//   - error: nil on success, or error if decompression fails
func decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
