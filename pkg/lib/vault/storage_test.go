package vault

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// TestIsFlintVault тестирует определение Flint Vault файлов
func TestIsFlintVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Тест 1: Валидный vault файл
	vaultPath := filepath.Join(tmpDir, "test.vault")
	err = CreateVault(vaultPath, "test123")
	if err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	isVault, err := IsFlintVault(vaultPath)
	if err != nil {
		t.Fatalf("IsFlintVault failed: %v", err)
	}
	if !isVault {
		t.Error("Expected true for valid vault file")
	}

	// Тест 2: Несуществующий файл
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.vault")
	_, err = IsFlintVault(nonExistentPath)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Тест 3: Обычный текстовый файл
	txtPath := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(txtPath, []byte("This is not a vault file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	isVault, err = IsFlintVault(txtPath)
	if err != nil {
		t.Fatalf("IsFlintVault failed: %v", err)
	}
	if isVault {
		t.Error("Expected false for text file")
	}

	// Тест 4: Файл с неправильным magic header
	fakePath := filepath.Join(tmpDir, "fake.vault")
	fakeHeader := VaultHeader{
		Version:       CurrentVaultVersion,
		Iterations:    PBKDF2Iters,
		DirectorySize: 100,
	}
	copy(fakeHeader.Magic[:], "FAKE001") // Неправильный magic

	file, err := os.Create(fakePath)
	if err != nil {
		t.Fatalf("Failed to create fake file: %v", err)
	}

	err = binary.Write(file, binary.LittleEndian, fakeHeader)
	file.Close()
	if err != nil {
		t.Fatalf("Failed to write fake header: %v", err)
	}

	isVault, err = IsFlintVault(fakePath)
	if err != nil {
		t.Fatalf("IsFlintVault failed: %v", err)
	}
	if isVault {
		t.Error("Expected false for file with wrong magic header")
	}

	// Тест 5: Слишком маленький файл
	smallPath := filepath.Join(tmpDir, "small.vault")
	err = os.WriteFile(smallPath, []byte("small"), 0644)
	if err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	isVault, err = IsFlintVault(smallPath)
	if err != nil {
		t.Fatalf("IsFlintVault failed: %v", err)
	}
	if isVault {
		t.Error("Expected false for too small file")
	}
}

// TestGetVaultInfo тестирует получение информации о vault файле
func TestGetVaultInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Тест 1: Валидный vault файл
	vaultPath := filepath.Join(tmpDir, "test.vault")
	err = CreateVault(vaultPath, "test123")
	if err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	info, err := GetVaultInfo(vaultPath)
	if err != nil {
		t.Fatalf("GetVaultInfo failed: %v", err)
	}

	if !info.IsFlintVault {
		t.Error("Expected IsFlintVault to be true")
	}
	if info.Version != CurrentVaultVersion {
		t.Errorf("Expected version %d, got %d", CurrentVaultVersion, info.Version)
	}
	if info.Iterations != PBKDF2Iters {
		t.Errorf("Expected iterations %d, got %d", PBKDF2Iters, info.Iterations)
	}
	if info.FileSize <= 0 {
		t.Error("Expected positive file size")
	}
	if info.FilePath != vaultPath {
		t.Errorf("Expected path %s, got %s", vaultPath, info.FilePath)
	}

	// Тест 2: Несуществующий файл
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.vault")
	_, err = GetVaultInfo(nonExistentPath)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Тест 3: Обычный файл
	txtPath := filepath.Join(tmpDir, "test.txt")
	testContent := "This is not a vault file"
	err = os.WriteFile(txtPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	info, err = GetVaultInfo(txtPath)
	if err != nil {
		t.Fatalf("GetVaultInfo failed: %v", err)
	}

	if info.IsFlintVault {
		t.Error("Expected IsFlintVault to be false for text file")
	}
	if info.Version != 0 {
		t.Error("Expected version 0 for non-vault file")
	}
	if info.Iterations != 0 {
		t.Error("Expected iterations 0 for non-vault file")
	}
	if info.FileSize != int64(len(testContent)) {
		t.Errorf("Expected file size %d, got %d", len(testContent), info.FileSize)
	}
}

// TestValidateVaultFile тестирует валидацию vault файлов
func TestValidateVaultFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Тест 1: Валидный vault файл
	vaultPath := filepath.Join(tmpDir, "test.vault")
	err = CreateVault(vaultPath, "test123")
	if err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	err = ValidateVaultFile(vaultPath)
	if err != nil {
		t.Errorf("ValidateVaultFile failed for valid vault: %v", err)
	}

	// Тест 2: Несуществующий файл
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.vault")
	err = ValidateVaultFile(nonExistentPath)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Тест 3: Слишком маленький файл
	smallPath := filepath.Join(tmpDir, "small.vault")
	err = os.WriteFile(smallPath, []byte("small"), 0644)
	if err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	err = ValidateVaultFile(smallPath)
	if err == nil {
		t.Error("Expected error for too small file")
	}

	// Тест 4: Файл с неправильным magic header
	wrongMagicPath := filepath.Join(tmpDir, "wrong_magic.vault")
	header := VaultHeader{
		Version:       CurrentVaultVersion,
		Iterations:    PBKDF2Iters,
		DirectorySize: 100,
	}
	copy(header.Magic[:], "WRONG01")

	file, err := os.Create(wrongMagicPath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	err = binary.Write(file, binary.LittleEndian, header)
	if err != nil {
		file.Close()
		t.Fatalf("Failed to write header: %v", err)
	}

	// Добавляем минимальные данные для прохождения проверки размера
	padding := make([]byte, 16)
	_, err = file.Write(padding)
	file.Close()
	if err != nil {
		t.Fatalf("Failed to write padding: %v", err)
	}

	err = ValidateVaultFile(wrongMagicPath)
	if err == nil {
		t.Error("Expected error for wrong magic header")
	}

	// Тест 5: Неподдерживаемая версия
	wrongVersionPath := filepath.Join(tmpDir, "wrong_version.vault")
	header = VaultHeader{
		Version:       999, // Неподдерживаемая версия
		Iterations:    PBKDF2Iters,
		DirectorySize: 100,
	}
	copy(header.Magic[:], VaultMagic)

	file, err = os.Create(wrongVersionPath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	err = binary.Write(file, binary.LittleEndian, header)
	if err != nil {
		file.Close()
		t.Fatalf("Failed to write header: %v", err)
	}

	_, err = file.Write(padding)
	file.Close()
	if err != nil {
		t.Fatalf("Failed to write padding: %v", err)
	}

	err = ValidateVaultFile(wrongVersionPath)
	if err == nil {
		t.Error("Expected error for unsupported version")
	}

	// Тест 6: Подозрительное количество итераций
	suspiciousIterPath := filepath.Join(tmpDir, "suspicious_iter.vault")
	header = VaultHeader{
		Version:       CurrentVaultVersion,
		Iterations:    100, // Слишком мало итераций
		DirectorySize: 100,
	}
	copy(header.Magic[:], VaultMagic)

	file, err = os.Create(suspiciousIterPath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	err = binary.Write(file, binary.LittleEndian, header)
	if err != nil {
		file.Close()
		t.Fatalf("Failed to write header: %v", err)
	}

	_, err = file.Write(padding)
	file.Close()
	if err != nil {
		t.Fatalf("Failed to write padding: %v", err)
	}

	err = ValidateVaultFile(suspiciousIterPath)
	if err == nil {
		t.Error("Expected error for suspicious iteration count")
	}
}

// BenchmarkIsFlintVault бенчмарк для проверки производительности
func BenchmarkIsFlintVault(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "bench.vault")
	err = CreateVault(vaultPath, "benchmark123")
	if err != nil {
		b.Fatalf("Failed to create vault: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := IsFlintVault(vaultPath)
		if err != nil {
			b.Fatalf("IsFlintVault failed: %v", err)
		}
	}
}

// BenchmarkGetVaultInfo бенчмарк для получения информации
func BenchmarkGetVaultInfo(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "bench.vault")
	err = CreateVault(vaultPath, "benchmark123")
	if err != nil {
		b.Fatalf("Failed to create vault: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetVaultInfo(vaultPath)
		if err != nil {
			b.Fatalf("GetVaultInfo failed: %v", err)
		}
	}
}

// BenchmarkValidateVaultFile бенчмарк для валидации
func BenchmarkValidateVaultFile(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "bench.vault")
	err = CreateVault(vaultPath, "benchmark123")
	if err != nil {
		b.Fatalf("Failed to create vault: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ValidateVaultFile(vaultPath)
		if err != nil {
			b.Fatalf("ValidateVaultFile failed: %v", err)
		}
	}
}
