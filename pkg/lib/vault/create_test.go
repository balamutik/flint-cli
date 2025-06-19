package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateVault(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name        string
		path        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid vault creation",
			path:        filepath.Join(tmpDir, "test_vault.dat"),
			password:    "SecurePassword123!",
			expectError: false,
		},
		{
			name:        "Empty password",
			path:        filepath.Join(tmpDir, "empty_pass.dat"),
			password:    "",
			expectError: true,
			errorMsg:    "password cannot be empty",
		},
		{
			name:        "Empty path",
			path:        "",
			password:    "password123",
			expectError: true,
			errorMsg:    "file path cannot be empty",
		},
		{
			name:        "Short password",
			path:        filepath.Join(tmpDir, "short_pass.dat"),
			password:    "123",
			expectError: false, // Short password allowed but not recommended
		},
		{
			name:        "Very long password",
			path:        filepath.Join(tmpDir, "long_pass.dat"),
			password:    "ThisIsAVeryLongPasswordThatShouldStillWorkFineBecauseWeUseProperKeyDerivation123!@#",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := CreateVault(tc.path, tc.password)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tc.errorMsg != "" && err.Error() != tc.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tc.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check that file was created
			if _, err := os.Stat(tc.path); os.IsNotExist(err) {
				t.Errorf("Vault file was not created: %s", tc.path)
			}

			// Check that we can load the vault
			data, err := loadVaultData(tc.path, tc.password)
			if err != nil {
				t.Errorf("Failed to load created vault: %v", err)
				return
			}

			// Check contents of new vault
			if len(data.Entries) != 0 {
				t.Errorf("New vault should be empty but contains %d items", len(data.Entries))
			}

			if data.Comment != "Encrypted Flint Vault Storage" {
				t.Errorf("Invalid comment in vault: %s", data.Comment)
			}

			// Check that creation time is reasonable (within last minute)
			if time.Since(data.CreatedAt) > time.Minute {
				t.Errorf("Vault creation time seems incorrect: %v", data.CreatedAt)
			}
		})
	}
}

func TestCreateVaultFileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "existing.dat")
	password := "password123"

	// Create first vault
	err = CreateVault(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create first vault: %v", err)
	}

	// Try to create vault with same name
	err = CreateVault(vaultPath, password)
	if err == nil {
		t.Errorf("Expected error when trying to overwrite existing file")
	}

	expectedError := "vault file already exists"
	if err != nil && err.Error() != expectedError+": "+vaultPath {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestLoadVaultDataInvalidPassword(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.dat")
	correctPassword := "CorrectPassword123!"
	wrongPassword := "WrongPassword"

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

	expectedError := "decryption failed: invalid password or corrupted data"
	if err != nil && err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestLoadVaultDataCorruptedFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create corrupted file
	corruptedPath := filepath.Join(tmpDir, "corrupted.dat")
	err = os.WriteFile(corruptedPath, []byte("corrupted data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Try to load corrupted file
	_, err = loadVaultData(corruptedPath, "password")
	if err == nil {
		t.Errorf("Expected error when loading corrupted file")
	}
}

func TestLoadVaultDataNonExistentFile(t *testing.T) {
	// Try to load non-existent file
	_, err := loadVaultData("/nonexistent/path/vault.dat", "password")
	if err == nil {
		t.Errorf("Expected error when loading non-existent file")
	}
}

func TestSaveLoadCyclePreservesData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "cycle_test.dat")
	password := "TestPassword123!"

	// Create test data
	originalData := VaultData{
		Entries: []VaultEntry{
			{
				Path:    "test_file.txt",
				Name:    "test_file.txt",
				IsDir:   false,
				Size:    12,
				Mode:    0644,
				ModTime: time.Now().Truncate(time.Second), // Remove nanoseconds for exact comparison
				Content: []byte("test content"),
			},
			{
				Path:    "test_dir",
				Name:    "test_dir",
				IsDir:   true,
				Size:    0,
				Mode:    0755,
				ModTime: time.Now().Truncate(time.Second),
				Content: nil,
			},
		},
		CreatedAt: time.Now().Truncate(time.Second),
		Comment:   "Test vault data",
	}

	// Save data
	err = saveVaultData(vaultPath, password, originalData)
	if err != nil {
		t.Fatalf("Failed to save data: %v", err)
	}

	// Load data
	loadedData, err := loadVaultData(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	// Compare data
	if len(loadedData.Entries) != len(originalData.Entries) {
		t.Errorf("Entry count mismatch: expected %d, got %d",
			len(originalData.Entries), len(loadedData.Entries))
	}

	for i, entry := range loadedData.Entries {
		orig := originalData.Entries[i]
		if entry.Path != orig.Path ||
			entry.Name != orig.Name ||
			entry.IsDir != orig.IsDir ||
			entry.Size != orig.Size ||
			entry.Mode != orig.Mode ||
			!entry.ModTime.Equal(orig.ModTime) {
			t.Errorf("Entry %d mismatch:\nExpected: %+v\nGot: %+v", i, orig, entry)
		}

		if !equalBytes(entry.Content, orig.Content) {
			t.Errorf("Entry %d content mismatch", i)
		}
	}

	if !loadedData.CreatedAt.Equal(originalData.CreatedAt) {
		t.Errorf("Creation time mismatch: expected %v, got %v",
			originalData.CreatedAt, loadedData.CreatedAt)
	}

	if loadedData.Comment != originalData.Comment {
		t.Errorf("Comment mismatch: expected '%s', got '%s'",
			originalData.Comment, loadedData.Comment)
	}
}

func TestVaultHeaderValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "header_test.dat")
	password := "password123"

	// Create valid vault
	err = CreateVault(vaultPath, password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	// Read file and corrupt magic header
	data, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Corrupt magic header (first 8 bytes)
	corruptedData := make([]byte, len(data))
	copy(corruptedData, data)
	copy(corruptedData[:8], []byte("BADMAGI!"))

	corruptedPath := filepath.Join(tmpDir, "bad_magic.dat")
	err = os.WriteFile(corruptedPath, corruptedData, 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}

	// Try to load file with invalid magic header
	_, err = loadVaultData(corruptedPath, password)
	if err == nil {
		t.Errorf("Expected error with invalid magic header")
	}

	expectedError := "invalid vault file format"
	if err != nil && err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// Helper function for comparing byte arrays
func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
