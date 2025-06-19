package vault

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// Test constants for info tests
const (
	infoTestPassword = "InfoTestPassword123!"
	infoTestContent  = "Test content for info testing"
)

// Setup helper for info tests
func setupInfoTest(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "info_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tmpDir
}

// Cleanup helper for info tests
func cleanupInfoTest(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("Warning: failed to cleanup test dir %s: %v", dir, err)
	}
}

// TestIsFlintVault tests the IsFlintVault function
func TestIsFlintVault(t *testing.T) {
	tmpDir := setupInfoTest(t)
	defer cleanupInfoTest(t, tmpDir)

	// Test with valid vault file
	vaultPath := filepath.Join(tmpDir, "test.vault")
	if err := CreateVault(vaultPath, infoTestPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	isVault, err := IsFlintVault(vaultPath)
	if err != nil {
		t.Fatalf("IsFlintVault failed: %v", err)
	}
	if !isVault {
		t.Fatal("Valid vault file was not recognized as Flint Vault")
	}

	// Test with non-vault file
	textFilePath := filepath.Join(tmpDir, "test.txt")
	if err := ioutil.WriteFile(textFilePath, []byte("This is not a vault file"), 0644); err != nil {
		t.Fatalf("Failed to create test text file: %v", err)
	}

	isVault, err = IsFlintVault(textFilePath)
	if err != nil {
		t.Fatalf("IsFlintVault failed on text file: %v", err)
	}
	if isVault {
		t.Fatal("Text file was incorrectly identified as Flint Vault")
	}

	// Test with non-existent file
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.vault")
	isVault, err = IsFlintVault(nonExistentPath)
	if err == nil {
		t.Fatal("Expected error when checking non-existent file")
	}
	if isVault {
		t.Fatal("Non-existent file should not be identified as vault")
	}

	// Test with empty file
	emptyFilePath := filepath.Join(tmpDir, "empty.vault")
	if err := ioutil.WriteFile(emptyFilePath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	isVault, err = IsFlintVault(emptyFilePath)
	if err != nil {
		t.Fatalf("IsFlintVault failed on empty file: %v", err)
	}
	if isVault {
		t.Fatal("Empty file was incorrectly identified as Flint Vault")
	}
}

// TestGetVaultInfo tests the GetVaultInfo function
func TestGetVaultInfo(t *testing.T) {
	tmpDir := setupInfoTest(t)
	defer cleanupInfoTest(t, tmpDir)

	// Test with valid vault file
	vaultPath := filepath.Join(tmpDir, "test.vault")
	if err := CreateVault(vaultPath, infoTestPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	info, err := GetVaultInfo(vaultPath)
	if err != nil {
		t.Fatalf("GetVaultInfo failed: %v", err)
	}

	// Verify vault info
	if !info.IsFlintVault {
		t.Fatal("Valid vault not recognized as Flint Vault")
	}

	if info.Version != CurrentVaultVersion {
		t.Fatalf("Expected version %d, got %d", CurrentVaultVersion, info.Version)
	}

	if info.Iterations != PBKDF2Iters {
		t.Fatalf("Expected %d iterations, got %d", PBKDF2Iters, info.Iterations)
	}

	if info.FileSize <= 0 {
		t.Fatal("Expected positive file size")
	}

	if info.FilePath != vaultPath {
		t.Fatalf("Expected file path %s, got %s", vaultPath, info.FilePath)
	}

	// Test with non-vault file
	textFilePath := filepath.Join(tmpDir, "test.txt")
	testContent := "This is a test text file"
	if err := ioutil.WriteFile(textFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test text file: %v", err)
	}

	info, err = GetVaultInfo(textFilePath)
	if err != nil {
		t.Fatalf("GetVaultInfo failed on text file: %v", err)
	}

	if info.IsFlintVault {
		t.Fatal("Text file incorrectly identified as Flint Vault")
	}

	if info.Version != 0 {
		t.Fatalf("Expected version 0 for non-vault, got %d", info.Version)
	}

	if info.Iterations != 0 {
		t.Fatalf("Expected 0 iterations for non-vault, got %d", info.Iterations)
	}

	if info.FileSize != int64(len(testContent)) {
		t.Fatalf("Expected file size %d, got %d", len(testContent), info.FileSize)
	}

	// Test with non-existent file
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.vault")
	_, err = GetVaultInfo(nonExistentPath)
	if err == nil {
		t.Fatal("Expected error when getting info for non-existent file")
	}
}

// TestValidateVaultFile tests the ValidateVaultFile function
func TestValidateVaultFile(t *testing.T) {
	tmpDir := setupInfoTest(t)
	defer cleanupInfoTest(t, tmpDir)

	// Test with valid vault file
	vaultPath := filepath.Join(tmpDir, "test.vault")
	if err := CreateVault(vaultPath, infoTestPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	err := ValidateVaultFile(vaultPath)
	if err != nil {
		t.Fatalf("ValidateVaultFile failed on valid vault: %v", err)
	}

	// Test with non-existent file
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.vault")
	err = ValidateVaultFile(nonExistentPath)
	if err == nil {
		t.Fatal("Expected error when validating non-existent file")
	}

	// Test with file that's too small
	smallFilePath := filepath.Join(tmpDir, "small.vault")
	if err := ioutil.WriteFile(smallFilePath, []byte("small"), 0644); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	err = ValidateVaultFile(smallFilePath)
	if err == nil {
		t.Fatal("Expected error when validating file that's too small")
	}

	// Test with file that has wrong magic header
	wrongMagicPath := filepath.Join(tmpDir, "wrong_magic.vault")
	wrongMagicData := make([]byte, 100) // Large enough to pass size check
	copy(wrongMagicData, []byte("WRONGMAG"))
	if err := ioutil.WriteFile(wrongMagicPath, wrongMagicData, 0644); err != nil {
		t.Fatalf("Failed to create wrong magic file: %v", err)
	}

	err = ValidateVaultFile(wrongMagicPath)
	if err == nil {
		t.Fatal("Expected error when validating file with wrong magic header")
	}
}

// TestVaultInfoIntegration tests full integration of info functions with vault operations
func TestVaultInfoIntegration(t *testing.T) {
	tmpDir := setupInfoTest(t)
	defer cleanupInfoTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "integration.vault")

	// Create vault
	if err := CreateVault(vaultPath, infoTestPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Test info functions on new vault
	isVault, err := IsFlintVault(vaultPath)
	if err != nil {
		t.Fatalf("IsFlintVault failed: %v", err)
	}
	if !isVault {
		t.Fatal("Newly created vault not recognized")
	}

	info, err := GetVaultInfo(vaultPath)
	if err != nil {
		t.Fatalf("GetVaultInfo failed: %v", err)
	}
	initialSize := info.FileSize

	if err := ValidateVaultFile(vaultPath); err != nil {
		t.Fatalf("ValidateVaultFile failed: %v", err)
	}

	// Add a file to vault
	testFilePath := filepath.Join(tmpDir, "test.txt")
	if err := ioutil.WriteFile(testFilePath, []byte(infoTestContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := AddFileToVault(vaultPath, infoTestPassword, testFilePath); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Test info functions after adding file
	isVault, err = IsFlintVault(vaultPath)
	if err != nil {
		t.Fatalf("IsFlintVault failed after adding file: %v", err)
	}
	if !isVault {
		t.Fatal("Vault not recognized after adding file")
	}

	info, err = GetVaultInfo(vaultPath)
	if err != nil {
		t.Fatalf("GetVaultInfo failed after adding file: %v", err)
	}

	// File size should have increased
	if info.FileSize <= initialSize {
		t.Fatalf("Expected file size to increase after adding file. Initial: %d, After: %d",
			initialSize, info.FileSize)
	}

	if err := ValidateVaultFile(vaultPath); err != nil {
		t.Fatalf("ValidateVaultFile failed after adding file: %v", err)
	}
}

// TestMultipleVaultVersions tests info functions with different vault characteristics
func TestMultipleVaultVersions(t *testing.T) {
	tmpDir := setupInfoTest(t)
	defer cleanupInfoTest(t, tmpDir)

	// Test with various vault configurations
	testCases := []struct {
		name     string
		password string
	}{
		{"simple", "simple123"},
		{"complex", "Complex!Password@123#$%"},
		{"unicode", "пароль测试🔒"},
		{"long", "ThisIsAVeryLongPasswordThatShouldStillWorkCorrectlyWithTheVaultSystem123456789"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vaultPath := filepath.Join(tmpDir, tc.name+".vault")

			// Create vault with specific password
			if err := CreateVault(vaultPath, tc.password); err != nil {
				t.Fatalf("CreateVault failed for %s: %v", tc.name, err)
			}

			// Test all info functions
			isVault, err := IsFlintVault(vaultPath)
			if err != nil {
				t.Fatalf("IsFlintVault failed for %s: %v", tc.name, err)
			}
			if !isVault {
				t.Fatalf("Vault %s not recognized", tc.name)
			}

			info, err := GetVaultInfo(vaultPath)
			if err != nil {
				t.Fatalf("GetVaultInfo failed for %s: %v", tc.name, err)
			}

			if !info.IsFlintVault {
				t.Fatalf("Vault %s not identified in info", tc.name)
			}

			if err := ValidateVaultFile(vaultPath); err != nil {
				t.Fatalf("ValidateVaultFile failed for %s: %v", tc.name, err)
			}
		})
	}
}

// BenchmarkIsFlintVault benchmarks the IsFlintVault function
func BenchmarkIsFlintVault(b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "info_bench_*")
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "bench.vault")
	CreateVault(vaultPath, infoTestPassword)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := IsFlintVault(vaultPath)
		if err != nil {
			b.Fatalf("IsFlintVault failed: %v", err)
		}
	}
}

// BenchmarkGetVaultInfo benchmarks the GetVaultInfo function
func BenchmarkGetVaultInfo(b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "info_bench_*")
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "bench.vault")
	CreateVault(vaultPath, infoTestPassword)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetVaultInfo(vaultPath)
		if err != nil {
			b.Fatalf("GetVaultInfo failed: %v", err)
		}
	}
}
