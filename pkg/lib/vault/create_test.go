package vault

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name        string
		vaultPath   string
		password    string
		expectError bool
	}{
		{
			name:        "Valid vault creation",
			vaultPath:   filepath.Join(tmpDir, "test1.dat"),
			password:    "validPassword123",
			expectError: false,
		},
		{
			name:        "Empty password",
			vaultPath:   filepath.Join(tmpDir, "test2.dat"),
			password:    "",
			expectError: true,
		},
		{
			name:        "Empty path",
			vaultPath:   "",
			password:    "validPassword123",
			expectError: true,
		},
		{
			name:        "Short password",
			vaultPath:   filepath.Join(tmpDir, "test3.dat"),
			password:    "short",
			expectError: false,
		},
		{
			name:        "Very long password",
			vaultPath:   filepath.Join(tmpDir, "test4.dat"),
			password:    "this_is_a_very_long_password_that_should_still_work_fine_even_though_it_is_quite_lengthy",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := CreateVault(tc.vaultPath, tc.password)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Check that file was created
				if _, err := os.Stat(tc.vaultPath); os.IsNotExist(err) {
					t.Errorf("Vault file was not created")
				}
			}
		})
	}
}

func TestCreateVaultFileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "existing.dat")
	password := "testPassword123"

	// Create vault first time
	err = CreateVault(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create initial vault: %v", err)
	}

	// Try to create again - should fail
	err = CreateVault(vaultPath, password)
	if err == nil {
		t.Errorf("Expected error when creating vault with existing file")
	}
}

func TestLoadVaultDataInvalidPassword(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.dat")
	correctPassword := "correctPassword123"
	wrongPassword := "wrongPassword456"

	// Create vault
	err = CreateVault(vaultPath, correctPassword)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	// Try to load with wrong password
	_, err = loadVaultData(vaultPath, wrongPassword)
	if err == nil {
		t.Errorf("Expected error with wrong password")
	}

	// Try to load with correct password
	_, err = loadVaultData(vaultPath, correctPassword)
	if err != nil {
		t.Errorf("Unexpected error with correct password: %v", err)
	}
}

func TestLoadVaultDataCorruptedFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	corruptedPath := filepath.Join(tmpDir, "corrupted.dat")

	// Create a file with invalid content
	err = os.WriteFile(corruptedPath, []byte("This is not a valid vault file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Try to load corrupted file
	_, err = loadVaultData(corruptedPath, "anypassword")
	if err == nil {
		t.Errorf("Expected error when loading corrupted file")
	}
}

func TestLoadVaultDataNonExistentFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nonExistentPath := filepath.Join(tmpDir, "nonexistent.dat")

	// Try to load non-existent file
	_, err = loadVaultData(nonExistentPath, "anypassword")
	if err == nil {
		t.Errorf("Expected error when loading non-existent file")
	}
}

func TestSaveLoadCyclePreservesData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "cycle_test.dat")
	password := "cycleTestPassword123"

	// Create test data
	originalData := VaultData{
		Entries: []VaultEntry{
			{
				Path:    "test_file.txt",
				Name:    "test_file.txt",
				IsDir:   false,
				Size:    100,
				Mode:    0644,
				ModTime: time.Now().Truncate(time.Second), // Truncate for comparison
				Content: []byte("test content"),
			},
		},
		CreatedAt: time.Now().Truncate(time.Second),
		Comment:   "Test vault for cycle testing",
	}

	// Save data
	err = saveVaultData(vaultPath, password, originalData)
	if err != nil {
		t.Fatalf("Failed to save vault data: %v", err)
	}

	// Load data back
	loadedData, err := loadVaultData(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to load vault data: %v", err)
	}

	// Compare data
	if len(loadedData.Entries) != len(originalData.Entries) {
		t.Errorf("Entry count mismatch: expected %d, got %d", len(originalData.Entries), len(loadedData.Entries))
	}

	if loadedData.Comment != originalData.Comment {
		t.Errorf("Comment mismatch: expected '%s', got '%s'", originalData.Comment, loadedData.Comment)
	}

	if len(originalData.Entries) > 0 && len(loadedData.Entries) > 0 {
		orig := originalData.Entries[0]
		loaded := loadedData.Entries[0]

		if orig.Path != loaded.Path {
			t.Errorf("Path mismatch: expected '%s', got '%s'", orig.Path, loaded.Path)
		}

		if string(orig.Content) != string(loaded.Content) {
			t.Errorf("Content mismatch: expected '%s', got '%s'", string(orig.Content), string(loaded.Content))
		}
	}
}

func TestVaultHeaderValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "header_test.dat")
	password := "headerTestPassword123"

	// Create vault
	err = CreateVault(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	// Open file and read header manually
	file, err := os.Open(vaultPath)
	if err != nil {
		t.Fatalf("Failed to open vault file: %v", err)
	}
	defer file.Close()

	var header VaultHeader
	err = readVaultHeader(file, &header)
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}

	// Check magic header
	if string(header.Magic[:]) != VaultMagic {
		t.Errorf("Magic header mismatch: expected '%s', got '%s'", VaultMagic, string(header.Magic[:]))
	}

	// Check version
	if header.Version != CurrentVaultVersion {
		t.Errorf("Version mismatch: expected %d, got %d", CurrentVaultVersion, header.Version)
	}

	// Check iteration count
	if header.Iterations != PBKDF2Iters {
		t.Errorf("Iterations mismatch: expected %d, got %d", PBKDF2Iters, header.Iterations)
	}
}

// Test new identification functions
func TestIsFlintVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid vault
	vaultPath := filepath.Join(tmpDir, "valid_vault.dat")
	password := "testPassword123"
	err = CreateVault(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	// Create non-vault file
	nonVaultPath := filepath.Join(tmpDir, "not_vault.txt")
	err = os.WriteFile(nonVaultPath, []byte("This is just a text file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create non-vault file: %v", err)
	}

	// Test valid vault
	isVault, err := IsFlintVault(vaultPath)
	if err != nil {
		t.Errorf("Unexpected error checking valid vault: %v", err)
	}
	if !isVault {
		t.Errorf("Valid vault not recognized as Flint Vault")
	}

	// Test non-vault file
	isVault, err = IsFlintVault(nonVaultPath)
	if err != nil {
		t.Errorf("Unexpected error checking non-vault file: %v", err)
	}
	if isVault {
		t.Errorf("Non-vault file incorrectly identified as Flint Vault")
	}

	// Test non-existent file
	_, err = IsFlintVault(filepath.Join(tmpDir, "nonexistent.dat"))
	if err == nil {
		t.Errorf("Expected error for non-existent file")
	}
}

func TestGetVaultInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid vault
	vaultPath := filepath.Join(tmpDir, "info_test.dat")
	password := "testPassword123"
	err = CreateVault(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	// Get vault info
	info, err := GetVaultInfo(vaultPath)
	if err != nil {
		t.Fatalf("Failed to get vault info: %v", err)
	}

	// Check results
	if !info.IsFlintVault {
		t.Errorf("Vault not identified as Flint Vault")
	}

	if info.Version != CurrentVaultVersion {
		t.Errorf("Version mismatch: expected %d, got %d", CurrentVaultVersion, info.Version)
	}

	if info.Iterations != PBKDF2Iters {
		t.Errorf("Iterations mismatch: expected %d, got %d", PBKDF2Iters, info.Iterations)
	}

	if info.FileSize <= 0 {
		t.Errorf("Invalid file size: %d", info.FileSize)
	}

	if info.FilePath != vaultPath {
		t.Errorf("File path mismatch: expected '%s', got '%s'", vaultPath, info.FilePath)
	}

	// Test non-vault file
	nonVaultPath := filepath.Join(tmpDir, "not_vault.txt")
	err = os.WriteFile(nonVaultPath, []byte("Not a vault"), 0644)
	if err != nil {
		t.Fatalf("Failed to create non-vault file: %v", err)
	}

	info, err = GetVaultInfo(nonVaultPath)
	if err != nil {
		t.Fatalf("Failed to get info for non-vault file: %v", err)
	}

	if info.IsFlintVault {
		t.Errorf("Non-vault file incorrectly identified as Flint Vault")
	}
}

func TestValidateVaultFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid vault
	validVaultPath := filepath.Join(tmpDir, "valid.dat")
	password := "testPassword123"
	err = CreateVault(validVaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	// Test valid vault
	err = ValidateVaultFile(validVaultPath)
	if err != nil {
		t.Errorf("Valid vault failed validation: %v", err)
	}

	// Test non-existent file
	err = ValidateVaultFile(filepath.Join(tmpDir, "nonexistent.dat"))
	if err == nil {
		t.Errorf("Expected error for non-existent file")
	}

	// Test file too small
	smallFile := filepath.Join(tmpDir, "small.dat")
	err = os.WriteFile(smallFile, []byte("small"), 0644)
	if err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	err = ValidateVaultFile(smallFile)
	if err == nil {
		t.Errorf("Expected error for file too small")
	}

	// Test file with wrong magic
	wrongMagicFile := filepath.Join(tmpDir, "wrong_magic.dat")
	wrongData := make([]byte, 100)
	copy(wrongData, []byte("WRONG001"))
	err = os.WriteFile(wrongMagicFile, wrongData, 0644)
	if err != nil {
		t.Fatalf("Failed to create wrong magic file: %v", err)
	}

	err = ValidateVaultFile(wrongMagicFile)
	if err == nil {
		t.Errorf("Expected error for wrong magic header")
	}
}

// Helper function to read vault header (for testing)
func readVaultHeader(file *os.File, header *VaultHeader) error {
	// Use the same binary.Read approach as in the main code
	return binary.Read(file, binary.LittleEndian, header)
}
