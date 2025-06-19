package vault

import (
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// Test constants
const (
	testPassword  = "TestPassword123!"
	testVaultPath = "test_vault.dat"
	testContent   = "This is test file content for encryption testing"
)

// Helper functions for test setup and cleanup
func setupTest(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tmpDir
}

func cleanupTest(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("Warning: failed to cleanup test dir %s: %v", dir, err)
	}
}

func createTestFile(t *testing.T, dir, filename, content string) string {
	filePath := filepath.Join(dir, filename)
	if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

// TestCreateVault tests basic vault creation
func TestCreateVault(t *testing.T) {
	tmpDir := setupTest(t)
	defer cleanupTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, testVaultPath)

	// Test successful vault creation
	err := CreateVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Check that vault file exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatal("Vault file was not created")
	}

	// Test creating vault with existing file (should fail)
	err = CreateVault(vaultPath, testPassword)
	if err == nil {
		t.Fatal("Expected error when creating vault with existing file")
	}

	// Test creating vault with empty password (should fail)
	vaultPath2 := filepath.Join(tmpDir, "vault2.dat")
	err = CreateVault(vaultPath2, "")
	if err == nil {
		t.Fatal("Expected error when creating vault with empty password")
	}
}

// TestAddFileToVault tests adding files to vault
func TestAddFileToVault(t *testing.T) {
	tmpDir := setupTest(t)
	defer cleanupTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, testVaultPath)
	testFilePath := createTestFile(t, tmpDir, "test.txt", testContent)

	// Create vault
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Test adding file to vault
	err := AddFileToVault(vaultPath, testPassword, testFilePath)
	if err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Verify file was added
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry in vault, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Name != "test.txt" {
		t.Fatalf("Expected file name 'test.txt', got '%s'", entry.Name)
	}

	if entry.Size != int64(len(testContent)) {
		t.Fatalf("Expected file size %d, got %d", len(testContent), entry.Size)
	}
}

// TestListVault tests vault listing functionality
func TestListVault(t *testing.T) {
	tmpDir := setupTest(t)
	defer cleanupTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, testVaultPath)

	// Create vault
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Test listing empty vault
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("Expected empty vault, got %d entries", len(entries))
	}

	// Add a file and test listing
	testFilePath := createTestFile(t, tmpDir, "test.txt", testContent)
	if err := AddFileToVault(vaultPath, testPassword, testFilePath); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	entries, err = ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	// Test listing with wrong password (should fail)
	_, err = ListVault(vaultPath, "wrongpassword")
	if err == nil {
		t.Fatal("Expected error when listing with wrong password")
	}
}

// TestExtractFromVault tests extraction functionality
func TestExtractFromVault(t *testing.T) {
	tmpDir := setupTest(t)
	defer cleanupTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, testVaultPath)
	outputDir := filepath.Join(tmpDir, "output")

	// Create vault and add test file
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	testFilePath := createTestFile(t, tmpDir, "test.txt", testContent)
	if err := AddFileToVault(vaultPath, testPassword, testFilePath); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Test extracting from vault
	err := ExtractFromVault(vaultPath, testPassword, outputDir)
	if err != nil {
		t.Fatalf("ExtractFromVault failed: %v", err)
	}

	// List files to debug
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("No entries in vault")
	}

	// Verify extracted file exists and has correct content
	// The file should be extracted with the same path as stored
	extractedPath := filepath.Join(outputDir, entries[0].Path)
	if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
		t.Fatalf("Extracted file does not exist at %s", extractedPath)
	}

	content, err := ioutil.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(content) != testContent {
		t.Fatalf("Content mismatch. Expected: '%s', got: '%s'", testContent, string(content))
	}
}

// TestGetFromVault tests selective extraction
func TestGetFromVault(t *testing.T) {
	tmpDir := setupTest(t)
	defer cleanupTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, testVaultPath)
	outputDir := filepath.Join(tmpDir, "output")

	// Create vault and add multiple files
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	file1 := createTestFile(t, tmpDir, "file1.txt", "Content 1")
	file2 := createTestFile(t, tmpDir, "file2.txt", "Content 2")

	if err := AddFileToVault(vaultPath, testPassword, file1); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}
	if err := AddFileToVault(vaultPath, testPassword, file2); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// List files to see what paths are stored
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// Find the path for file1.txt in the vault
	var file1Path string
	for _, entry := range entries {
		if entry.Name == "file1.txt" {
			file1Path = entry.Path
			break
		}
	}

	if file1Path == "" {
		t.Fatal("file1.txt not found in vault")
	}

	// Test selective extraction using the stored path
	err = GetFromVault(vaultPath, testPassword, outputDir, []string{file1Path})
	if err != nil {
		t.Fatalf("GetFromVault failed: %v", err)
	}

	// Verify only file1.txt was extracted
	extractedPath1 := filepath.Join(outputDir, file1Path)
	extractedPath2 := filepath.Join(outputDir, entries[1].Path)

	if _, err := os.Stat(extractedPath1); os.IsNotExist(err) {
		t.Fatal("Expected file1.txt to be extracted")
	}

	if _, err := os.Stat(extractedPath2); err == nil {
		t.Fatal("Expected file2.txt to NOT be extracted")
	}
}

// TestRemoveFromVault tests file removal
func TestRemoveFromVault(t *testing.T) {
	tmpDir := setupTest(t)
	defer cleanupTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, testVaultPath)

	// Create vault and add test files
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	file1 := createTestFile(t, tmpDir, "file1.txt", "Content 1")
	file2 := createTestFile(t, tmpDir, "file2.txt", "Content 2")

	if err := AddFileToVault(vaultPath, testPassword, file1); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}
	if err := AddFileToVault(vaultPath, testPassword, file2); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// List files before removal to get correct paths
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries before removal, got %d", len(entries))
	}

	// Find the path for file1.txt
	var file1Path string
	for _, entry := range entries {
		if entry.Name == "file1.txt" {
			file1Path = entry.Path
			break
		}
	}

	if file1Path == "" {
		t.Fatal("file1.txt not found in vault")
	}

	// Test removing file from vault using the stored path
	err = RemoveFromVault(vaultPath, testPassword, []string{file1Path})
	if err != nil {
		t.Fatalf("RemoveFromVault failed: %v", err)
	}

	// Verify file was removed
	entries, err = ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry after removal, got %d", len(entries))
	}

	if entries[0].Name != "file2.txt" {
		t.Fatalf("Expected remaining file to be 'file2.txt', got '%s'", entries[0].Name)
	}
}

// TestVaultIntegrity tests SHA-256 integrity verification
func TestVaultIntegrity(t *testing.T) {
	tmpDir := setupTest(t)
	defer cleanupTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, testVaultPath)

	// Create vault and add test file
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	testFilePath := createTestFile(t, tmpDir, "test.txt", testContent)
	if err := AddFileToVault(vaultPath, testPassword, testFilePath); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Get file entry to check hash
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	expectedHash := sha256.Sum256([]byte(testContent))

	if entry.SHA256Hash != expectedHash {
		t.Fatalf("Hash mismatch. Expected: %x, got: %x", expectedHash, entry.SHA256Hash)
	}
}

// TestVaultCompression tests compression functionality
func TestVaultCompression(t *testing.T) {
	tmpDir := setupTest(t)
	defer cleanupTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, testVaultPath)

	// Create vault
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Create a highly compressible file (repeated content)
	compressibleContent := ""
	for i := 0; i < 1000; i++ {
		compressibleContent += "This is repeated content that should compress well. "
	}

	testFilePath := createTestFile(t, tmpDir, "compressible.txt", compressibleContent)
	if err := AddFileToVault(vaultPath, testPassword, testFilePath); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Check that compressed size is smaller than original size
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.CompressedSize >= entry.Size {
		t.Fatalf("Expected compressed size (%d) to be smaller than original size (%d)",
			entry.CompressedSize, entry.Size)
	}

	compressionRatio := float64(entry.CompressedSize) / float64(entry.Size)
	if compressionRatio > 0.5 {
		t.Fatalf("Expected compression ratio < 0.5, got %f", compressionRatio)
	}
}

// TestPasswordSecurity tests password validation
func TestPasswordSecurity(t *testing.T) {
	tmpDir := setupTest(t)
	defer cleanupTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, testVaultPath)

	// Create vault with one password
	password1 := "Password123!"
	if err := CreateVault(vaultPath, password1); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Test that wrong password fails
	password2 := "DifferentPassword456!"
	_, err := ListVault(vaultPath, password2)
	if err == nil {
		t.Fatal("Expected error when using wrong password")
	}

	// Test that correct password works
	entries, err := ListVault(vaultPath, password1)
	if err != nil {
		t.Fatalf("ListVault with correct password failed: %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("Expected empty vault, got %d entries", len(entries))
	}
}

// TestDirectoryAddition tests adding directories to vault
func TestDirectoryAddition(t *testing.T) {
	tmpDir := setupTest(t)
	defer cleanupTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, testVaultPath)

	// Create test directory structure
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	subDir := filepath.Join(testDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	createTestFile(t, testDir, "file1.txt", "Content 1")
	createTestFile(t, testDir, "file2.txt", "Content 2")
	createTestFile(t, subDir, "file3.txt", "Content 3")

	// Create vault
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Test adding directory to vault
	err := AddDirectoryToVault(vaultPath, testPassword, testDir)
	if err != nil {
		t.Fatalf("AddDirectoryToVault failed: %v", err)
	}

	// Verify directory and files were added
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("Expected entries in vault after adding directory")
	}

	// Count files (should have at least 3 files)
	var fileCount int
	for _, entry := range entries {
		if !entry.IsDir {
			fileCount++
		}
	}

	if fileCount < 3 {
		t.Fatalf("Expected at least 3 files in vault, got %d", fileCount)
	}
}

// Benchmark tests
func BenchmarkCreateVault(b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "vault_bench_*")
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vaultPath := filepath.Join(tmpDir, "bench_vault_"+string(rune(i))+".dat")
		if err := CreateVault(vaultPath, testPassword); err != nil {
			b.Fatalf("CreateVault failed: %v", err)
		}
	}
}

func BenchmarkAddFile(b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "vault_bench_*")
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "bench_vault.dat")
	CreateVault(vaultPath, testPassword)

	testFilePath := filepath.Join(tmpDir, "bench_file.txt")
	ioutil.WriteFile(testFilePath, []byte(testContent), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := AddFileToVault(vaultPath, testPassword, testFilePath); err != nil {
			b.Fatalf("AddFileToVault failed: %v", err)
		}
	}
}
