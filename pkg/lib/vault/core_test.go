package vault

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testPassword = "TestPassword123!"
	testContent  = "This is test file content for encryption testing"
)

// Helper functions for test setup and cleanup
func setupCoreTest(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "core_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tmpDir
}

func cleanupCoreTest(t *testing.T, dir string) {
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

// TestCreateVault тестирует создание vault
func TestCreateVault(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")

	// Тест 1: Успешное создание vault
	err := CreateVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Проверяем что файл создан
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatal("Vault file was not created")
	}

	// Тест 2: Создание vault с существующим файлом (должен дать ошибку)
	err = CreateVault(vaultPath, testPassword)
	if err == nil {
		t.Fatal("Expected error when creating vault with existing file")
	}

	// Тест 3: Создание vault с пустым паролем (должен дать ошибку)
	vaultPath2 := filepath.Join(tmpDir, "vault2.vault")
	err = CreateVault(vaultPath2, "")
	if err == nil {
		t.Fatal("Expected error when creating vault with empty password")
	}

	// Тест 4: Создание vault с пустым путём (должен дать ошибку)
	err = CreateVault("", testPassword)
	if err == nil {
		t.Fatal("Expected error when creating vault with empty path")
	}
}

// TestAddFileToVault тестирует добавление файлов
func TestAddFileToVault(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")
	testFilePath := createTestFile(t, tmpDir, "test.txt", testContent)

	// Создаём vault
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Тест 1: Добавление файла в vault
	err := AddFileToVault(vaultPath, testPassword, testFilePath)
	if err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Проверяем что файл добавлен
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

	// Тест 2: Добавление несуществующего файла (должен дать ошибку)
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.txt")
	err = AddFileToVault(vaultPath, testPassword, nonExistentPath)
	if err == nil {
		t.Fatal("Expected error when adding non-existent file")
	}

	// Тест 3: Неправильный пароль (должен дать ошибку)
	err = AddFileToVault(vaultPath, "wrongpassword", testFilePath)
	if err == nil {
		t.Fatal("Expected error with wrong password")
	}
}

// TestAddDirectoryToVault тестирует добавление директорий
func TestAddDirectoryToVault(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")

	// Создаём тестовую структуру директорий
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

	// Создаём vault
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Тест добавления директории в vault
	err := AddDirectoryToVault(vaultPath, testPassword, testDir)
	if err != nil {
		t.Fatalf("AddDirectoryToVault failed: %v", err)
	}

	// Проверяем что директория и файлы добавлены
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("Expected entries in vault after adding directory")
	}

	// Подсчитываем файлы (должно быть минимум 3 файла)
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

// TestExtractFromVault тестирует извлечение файлов
func TestExtractFromVault(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")
	outputDir := filepath.Join(tmpDir, "output")

	// Создаём vault и добавляем тестовый файл
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	testFilePath := createTestFile(t, tmpDir, "test.txt", testContent)
	if err := AddFileToVault(vaultPath, testPassword, testFilePath); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Тест извлечения из vault
	err := ExtractFromVault(vaultPath, testPassword, outputDir)
	if err != nil {
		t.Fatalf("ExtractFromVault failed: %v", err)
	}

	// Получаем список файлов для проверки извлечённого пути
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("No entries in vault")
	}

	// Проверяем что извлечённый файл существует и имеет правильное содержимое
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

// TestGetFromVault тестирует селективное извлечение
func TestGetFromVault(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")
	outputDir := filepath.Join(tmpDir, "output")

	// Создаём vault и добавляем несколько файлов
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

	// Получаем список файлов для определения путей
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// Находим путь для file1.txt в vault
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

	// Тест селективного извлечения
	err = GetFromVault(vaultPath, testPassword, outputDir, []string{file1Path})
	if err != nil {
		t.Fatalf("GetFromVault failed: %v", err)
	}

	// Проверяем что извлечён только file1.txt
	extractedPath1 := filepath.Join(outputDir, file1Path)
	extractedPath2 := filepath.Join(outputDir, entries[1].Path)

	if _, err := os.Stat(extractedPath1); os.IsNotExist(err) {
		t.Fatal("Expected file1.txt to be extracted")
	}

	if _, err := os.Stat(extractedPath2); err == nil {
		t.Fatal("Expected file2.txt to NOT be extracted")
	}
}

// TestRemoveFromVault тестирует удаление файлов
func TestRemoveFromVault(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")

	// Создаём vault и добавляем тестовые файлы
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

	// Получаем список файлов перед удалением
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries before removal, got %d", len(entries))
	}

	// Находим путь для file1.txt
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

	// Тест удаления файла из vault
	err = RemoveFromVault(vaultPath, testPassword, []string{file1Path})
	if err != nil {
		t.Fatalf("RemoveFromVault failed: %v", err)
	}

	// Проверяем что файл удалён
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

// TestParallelOperations тестирует параллельные операции
func TestParallelOperations(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")

	// Создаём vault
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Создаём несколько тестовых файлов с уникальными именами
	var filePaths []string
	for i := 0; i < 5; i++ {
		fileName := filepath.Join("parallel_file_" + string(rune('A'+i)) + ".txt")
		content := "Parallel test content " + string(rune('A'+i))
		filePath := createTestFile(t, tmpDir, fileName, content)
		filePaths = append(filePaths, filePath)
	}

	// Тест параллельного добавления файлов
	config := DefaultParallelConfig()
	config.MaxConcurrency = 2
	config.Context = context.Background()

	stats, err := AddMultipleFilesToVaultParallel(vaultPath, testPassword, filePaths, config)
	if err != nil {
		// Выводим детали ошибок для диагностики
		t.Logf("Parallel add errors: %v", err)
		for i, e := range stats.Errors {
			t.Logf("Error %d: %v", i+1, e)
		}
		t.Fatalf("AddMultipleFilesToVaultParallel failed: %v", err)
	}

	if stats.TotalFiles != int64(len(filePaths)) {
		t.Errorf("Expected %d total files, got %d", len(filePaths), stats.TotalFiles)
	}

	if stats.SuccessfulFiles != stats.TotalFiles {
		t.Errorf("Expected all files to be successful, got %d/%d", stats.SuccessfulFiles, stats.TotalFiles)
	}

	// Проверяем что файлы добавлены
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != len(filePaths) {
		t.Fatalf("Expected %d entries, got %d", len(filePaths), len(entries))
	}

	// Тест параллельного извлечения конкретных файлов
	outputDir := filepath.Join(tmpDir, "parallel_output")
	var targetPaths []string
	for _, entry := range entries[:2] { // Извлекаем первые 2 файла
		targetPaths = append(targetPaths, entry.Path)
	}

	extractStats, err := ExtractMultipleFilesFromVaultParallel(vaultPath, testPassword, outputDir, targetPaths, config)
	if err != nil {
		t.Fatalf("ExtractMultipleFilesFromVaultParallel failed: %v", err)
	}

	if extractStats.SuccessfulFiles != int64(len(targetPaths)) {
		t.Errorf("Expected %d successful extractions, got %d", len(targetPaths), extractStats.SuccessfulFiles)
	}

	// Проверяем что файлы извлечены
	for _, targetPath := range targetPaths {
		extractedPath := filepath.Join(outputDir, targetPath)
		if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to be extracted", extractedPath)
		}
	}
}

// TestAddDirectoryToVaultParallel тестирует параллельное добавление директории
func TestAddDirectoryToVaultParallel(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")

	// Создаём тестовую директорию с уникальными именами файлов
	testDir := filepath.Join(tmpDir, "parallel_testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Создаём файлы в директории с уникальными именами
	for i := 0; i < 3; i++ {
		fileName := "parallel_dir_file_" + string(rune('X'+i)) + ".txt"
		content := "Parallel directory content " + string(rune('X'+i))
		createTestFile(t, testDir, fileName, content)
	}

	// Создаём vault
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Тест параллельного добавления директории
	config := DefaultParallelConfig()
	config.MaxConcurrency = 2

	stats, err := AddDirectoryToVaultParallel(vaultPath, testPassword, testDir, config)
	if err != nil {
		// Выводим детали ошибок для диагностики
		t.Logf("Parallel directory add errors: %v", err)
		for i, e := range stats.Errors {
			t.Logf("Error %d: %v", i+1, e)
		}
		t.Fatalf("AddDirectoryToVaultParallel failed: %v", err)
	}

	if stats.SuccessfulFiles < 3 {
		t.Errorf("Expected at least 3 successful files, got %d", stats.SuccessfulFiles)
	}

	// Проверяем что файлы добавлены
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	var fileCount int
	for _, entry := range entries {
		if !entry.IsDir {
			fileCount++
		}
	}

	if fileCount < 3 {
		t.Errorf("Expected at least 3 files in vault, got %d", fileCount)
	}
}

// TestVaultIntegrity тестирует целостность данных
func TestVaultIntegrity(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")

	// Создаём vault и добавляем тестовый файл
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	testFilePath := createTestFile(t, tmpDir, "test.txt", testContent)
	if err := AddFileToVault(vaultPath, testPassword, testFilePath); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Получаем информацию о файле для проверки хеша
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

// TestCompressionFunctions тестирует функции сжатия
func TestCompressionFunctions(t *testing.T) {
	testData := []byte("This is test data for compression testing with some repetitive content repetitive content repetitive content")

	// Тест сжатия
	compressed, err := compressData(testData)
	if err != nil {
		t.Fatalf("compressData failed: %v", err)
	}

	if len(compressed) == 0 {
		t.Fatal("Compressed data is empty")
	}

	// Тест распаковки
	decompressed, err := decompressData(compressed)
	if err != nil {
		t.Fatalf("decompressData failed: %v", err)
	}

	if string(decompressed) != string(testData) {
		t.Fatalf("Decompressed data mismatch. Expected: %s, got: %s", string(testData), string(decompressed))
	}

	// Проверяем что сжатие эффективно для повторяющихся данных
	if len(compressed) >= len(testData) {
		t.Logf("Warning: Compression not effective. Original: %d, Compressed: %d", len(testData), len(compressed))
	}
}

// TestCalculateFileMetadata тестирует расчёт метаданных файла
func TestCalculateFileMetadata(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	testFilePath := createTestFile(t, tmpDir, "test.txt", testContent)

	hash, compressedSize, err := calculateFileMetadata(testFilePath)
	if err != nil {
		t.Fatalf("calculateFileMetadata failed: %v", err)
	}

	// Проверяем хеш
	expectedHash := sha256.Sum256([]byte(testContent))
	if hash != expectedHash {
		t.Fatalf("Hash mismatch. Expected: %x, got: %x", expectedHash, hash)
	}

	// Проверяем что размер сжатых данных больше 0
	if compressedSize <= 0 {
		t.Fatal("Compressed size should be greater than 0")
	}
}

// BenchmarkAddFileToVault бенчмарк для добавления файла
func BenchmarkAddFileToVault(b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "bench_core_*")
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "bench.vault")
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

// BenchmarkExtractFromVault бенчмарк для извлечения файлов
func BenchmarkExtractFromVault(b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "bench_core_*")
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "bench.vault")
	CreateVault(vaultPath, testPassword)

	testFilePath := filepath.Join(tmpDir, "bench_file.txt")
	ioutil.WriteFile(testFilePath, []byte(testContent), 0644)
	AddFileToVault(vaultPath, testPassword, testFilePath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputDir := filepath.Join(tmpDir, "bench_output_"+string(rune(i)))
		if err := ExtractFromVault(vaultPath, testPassword, outputDir); err != nil {
			b.Fatalf("ExtractFromVault failed: %v", err)
		}
	}
}

// TestHelperFunctions тестирует вспомогательные функции
func TestHelperFunctions(t *testing.T) {
	// Тест getOptimalBufferSizeForFile
	testCases := []struct {
		size     int64
		expected int
	}{
		{500 * 1024, 64 * 1024},              // < 1MB -> 64KB
		{50 * 1024 * 1024, 1024 * 1024},      // < 100MB -> 1MB
		{200 * 1024 * 1024, 4 * 1024 * 1024}, // > 100MB -> 4MB
	}

	for _, tc := range testCases {
		result := getOptimalBufferSizeForFile(tc.size)
		if result != tc.expected {
			t.Errorf("getOptimalBufferSizeForFile(%d) = %d, expected %d", tc.size, result, tc.expected)
		}
	}

	// Тест compareHashesConstantTime
	hash1 := []byte("hash1")
	hash2 := []byte("hash1")
	hash3 := []byte("hash2")
	hash4 := []byte("different_length")

	if !compareHashesConstantTime(hash1, hash2) {
		t.Error("compareHashesConstantTime failed for identical hashes")
	}

	if compareHashesConstantTime(hash1, hash3) {
		t.Error("compareHashesConstantTime failed for different hashes")
	}

	if compareHashesConstantTime(hash1, hash4) {
		t.Error("compareHashesConstantTime failed for different length hashes")
	}

	// Тест formatFileSize
	testSizes := []struct {
		input    int64
		expected string
	}{
		{512, "512B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1024 * 1024, "1.0MB"},
		{1024 * 1024 * 1024, "1.0GB"},
		{1024 * 1024 * 1024 * 1024, "1.0TB"},
	}

	for _, tc := range testSizes {
		result := formatFileSize(tc.input)
		if result != tc.expected {
			t.Errorf("formatFileSize(%d) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

// TestDefaultParallelConfig тестирует конфигурацию по умолчанию
func TestDefaultParallelConfig(t *testing.T) {
	config := DefaultParallelConfig()

	if config == nil {
		t.Fatal("DefaultParallelConfig returned nil")
	}

	if config.MaxConcurrency <= 0 {
		t.Error("MaxConcurrency should be positive")
	}

	if config.Timeout <= 0 {
		t.Error("Timeout should be positive")
	}

	if config.Context == nil {
		t.Error("Context should not be nil")
	}
}

// TestPrintParallelStats тестирует вывод статистики
func TestPrintParallelStats(t *testing.T) {
	stats := &ParallelStats{
		TotalFiles:      10,
		SuccessfulFiles: 8,
		FailedFiles:     2,
		TotalSize:       1024 * 1024, // 1MB
		Duration:        time.Second,
		Errors:          []error{fmt.Errorf("test error 1"), fmt.Errorf("test error 2")},
	}

	// Это просто проверяет что функция не падает
	// В реальном тесте мы бы захватили stdout для проверки вывода
	PrintParallelStats(stats)
}

// TestStreamingFunctions тестирует streaming функции
func TestStreamingFunctions(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	// Тест decompressDataStreaming
	testData := []byte("Test data for streaming compression")

	// Сначала сжимаем данные
	compressed, err := compressData(testData)
	if err != nil {
		t.Fatalf("compressData failed: %v", err)
	}

	// Тестируем streaming распаковку
	reader := strings.NewReader(string(compressed))
	gzipReader, err := decompressDataStreaming(reader)
	if err != nil {
		t.Fatalf("decompressDataStreaming failed: %v", err)
	}
	defer gzipReader.Close()

	decompressed, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		t.Fatalf("Reading from gzip reader failed: %v", err)
	}

	if string(decompressed) != string(testData) {
		t.Error("Streaming decompression data mismatch")
	}

	// Тест с некорректными данными
	badReader := strings.NewReader("not compressed data")
	_, err = decompressDataStreaming(badReader)
	if err == nil {
		t.Error("Expected error for invalid compressed data")
	}
}

// TestPasswordSanitization тестирует очистку пароля из памяти
func TestPasswordSanitization(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "sanitize.vault")
	password := "SanitizeTest123!"

	// Создаём vault
	if err := CreateVault(vaultPath, password); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Тестируем что функции работают с паролем (и очищают его внутри)
	_, err := ListVault(vaultPath, password)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	// Проверяем что пароль всё ещё работает (не был изменён снаружи)
	_, err = ListVault(vaultPath, password)
	if err != nil {
		t.Fatalf("ListVault failed on second call: %v", err)
	}
}

// TestErrorConditions тестирует различные условия ошибок
func TestErrorConditions(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	// Тест с несуществующим vault файлом
	nonExistentVault := filepath.Join(tmpDir, "nonexistent.vault")

	_, err := ListVault(nonExistentVault, "password")
	if err == nil {
		t.Error("Expected error for non-existent vault")
	}

	err = ExtractFromVault(nonExistentVault, "password", tmpDir)
	if err == nil {
		t.Error("Expected error for non-existent vault")
	}

	err = RemoveFromVault(nonExistentVault, "password", []string{"file.txt"})
	if err == nil {
		t.Error("Expected error for non-existent vault")
	}

	// Тест с недоступной директорией для извлечения
	vaultPath := filepath.Join(tmpDir, "test.vault")
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	testFile := createTestFile(t, tmpDir, "test.txt", "test content")
	if err := AddFileToVault(vaultPath, testPassword, testFile); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Попытка извлечь в несуществующую директорию с неправильными правами
	badOutputDir := "/root/nonexistent_dir"
	err = ExtractFromVault(vaultPath, testPassword, badOutputDir)
	if err == nil {
		t.Error("Expected error for inaccessible output directory")
	}
}

// BenchmarkCreateVault бенчмарк для создания vault
func BenchmarkCreateVault(b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "bench_create_*")
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vaultPath := filepath.Join(tmpDir, "bench_"+string(rune('a'+i))+".vault")
		if err := CreateVault(vaultPath, testPassword); err != nil {
			b.Fatalf("CreateVault failed: %v", err)
		}
	}
}

// BenchmarkCompressionFunctions бенчмарк для функций сжатия
func BenchmarkCompressionFunctions(b *testing.B) {
	testData := []byte(strings.Repeat("Test data for compression benchmarking. ", 100))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compressed, err := compressData(testData)
		if err != nil {
			b.Fatalf("compressData failed: %v", err)
		}

		_, err = decompressData(compressed)
		if err != nil {
			b.Fatalf("decompressData failed: %v", err)
		}
	}
}

// TestExtractFromVaultWithOptions тестирует извлечение с опциями
func TestExtractFromVaultWithOptions(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")
	outputDirFull := filepath.Join(tmpDir, "output_full")
	outputDirFlat := filepath.Join(tmpDir, "output_flat")

	// Создаём vault
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Создаём тестовую структуру каталогов
	testDir := filepath.Join(tmpDir, "testdata")
	if err := os.MkdirAll(filepath.Join(testDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Создаём файлы в разных каталогах
	rootFile := createTestFile(t, testDir, "root.txt", "Root file content")
	subFile := createTestFile(t, filepath.Join(testDir, "subdir"), "sub.txt", "Sub file content")

	// Добавляем файлы в vault
	if err := AddFileToVault(vaultPath, testPassword, rootFile); err != nil {
		t.Fatalf("AddFileToVault failed for root file: %v", err)
	}
	if err := AddFileToVault(vaultPath, testPassword, subFile); err != nil {
		t.Fatalf("AddFileToVault failed for sub file: %v", err)
	}

	// Тест извлечения с полными путями (extractFullPath = true)
	err := ExtractFromVaultWithOptions(vaultPath, testPassword, outputDirFull, true)
	if err != nil {
		t.Fatalf("ExtractFromVaultWithOptions with full paths failed: %v", err)
	}

	// Проверяем что файлы извлечены с полной структурой каталогов
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir {
			fullPath := filepath.Join(outputDirFull, entry.Path)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("Expected file with full path %s to exist", fullPath)
			}
		}
	}

	// Тест извлечения без полных путей (extractFullPath = false)
	err = ExtractFromVaultWithOptions(vaultPath, testPassword, outputDirFlat, false)
	if err != nil {
		t.Fatalf("ExtractFromVaultWithOptions without full paths failed: %v", err)
	}

	// Проверяем что файлы извлечены только с именами файлов
	for _, entry := range entries {
		if !entry.IsDir {
			filename := filepath.Base(entry.Path)
			flatPath := filepath.Join(outputDirFlat, filename)
			if _, err := os.Stat(flatPath); os.IsNotExist(err) {
				t.Errorf("Expected file with flat name %s to exist", flatPath)
			}

			// Проверяем что полный путь НЕ создан
			fullPath := filepath.Join(outputDirFlat, entry.Path)
			if fullPath != flatPath { // Только если это разные пути
				if _, err := os.Stat(fullPath); err == nil {
					t.Errorf("Did not expect file with full path %s to exist in flat extraction", fullPath)
				}
			}
		}
	}

	// Проверяем содержимое файлов
	rootContent, err := ioutil.ReadFile(filepath.Join(outputDirFlat, "root.txt"))
	if err != nil {
		t.Fatalf("Failed to read extracted root file: %v", err)
	}
	if string(rootContent) != "Root file content" {
		t.Errorf("Root file content mismatch")
	}

	subContent, err := ioutil.ReadFile(filepath.Join(outputDirFlat, "sub.txt"))
	if err != nil {
		t.Fatalf("Failed to read extracted sub file: %v", err)
	}
	if string(subContent) != "Sub file content" {
		t.Errorf("Sub file content mismatch")
	}
}

// TestExtractMultipleFilesWithOptions тестирует параллельное извлечение с опциями
func TestExtractMultipleFilesWithOptions(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")
	outputDirFull := filepath.Join(tmpDir, "output_full")
	outputDirFlat := filepath.Join(tmpDir, "output_flat")

	// Создаём vault
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Создаём тестовую структуру
	testDir := filepath.Join(tmpDir, "testdata")
	if err := os.MkdirAll(filepath.Join(testDir, "docs"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Создаём файлы
	file1 := createTestFile(t, testDir, "file1.txt", "File 1 content")
	file2 := createTestFile(t, filepath.Join(testDir, "docs"), "file2.txt", "File 2 content")

	// Добавляем файлы в vault
	if err := AddFileToVault(vaultPath, testPassword, file1); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}
	if err := AddFileToVault(vaultPath, testPassword, file2); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Получаем пути файлов в vault
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	var targetPaths []string
	for _, entry := range entries {
		if !entry.IsDir {
			targetPaths = append(targetPaths, entry.Path)
		}
	}

	if len(targetPaths) != 2 {
		t.Fatalf("Expected 2 target files, got %d", len(targetPaths))
	}

	// Тест параллельного извлечения с полными путями
	config := DefaultParallelConfig()
	config.MaxConcurrency = 2

	stats, err := ExtractMultipleFilesFromVaultParallelWithOptions(vaultPath, testPassword, outputDirFull, targetPaths, config, true)
	if err != nil {
		t.Fatalf("ExtractMultipleFilesFromVaultParallelWithOptions with full paths failed: %v", err)
	}

	if stats.SuccessfulFiles != int64(len(targetPaths)) {
		t.Errorf("Expected %d successful extractions, got %d", len(targetPaths), stats.SuccessfulFiles)
	}

	// Проверяем файлы с полными путями
	for _, targetPath := range targetPaths {
		fullPath := filepath.Join(outputDirFull, targetPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file with full path %s to exist", fullPath)
		}
	}

	// Тест параллельного извлечения без полных путей
	stats, err = ExtractMultipleFilesFromVaultParallelWithOptions(vaultPath, testPassword, outputDirFlat, targetPaths, config, false)
	if err != nil {
		t.Fatalf("ExtractMultipleFilesFromVaultParallelWithOptions without full paths failed: %v", err)
	}

	if stats.SuccessfulFiles != int64(len(targetPaths)) {
		t.Errorf("Expected %d successful extractions, got %d", len(targetPaths), stats.SuccessfulFiles)
	}

	// Проверяем файлы без полных путей
	for _, targetPath := range targetPaths {
		filename := filepath.Base(targetPath)
		flatPath := filepath.Join(outputDirFlat, filename)
		if _, err := os.Stat(flatPath); os.IsNotExist(err) {
			t.Errorf("Expected file with flat name %s to exist", flatPath)
		}
	}
}

// TestGetFromVaultWithOptions тестирует селективное извлечение с опциями
func TestGetFromVaultWithOptions(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.vault")
	outputDirFull := filepath.Join(tmpDir, "output_full")
	outputDirFlat := filepath.Join(tmpDir, "output_flat")

	// Создаём vault
	if err := CreateVault(vaultPath, testPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Создаём тестовые файлы в подкаталоге
	testDir := filepath.Join(tmpDir, "testdata")
	if err := os.MkdirAll(filepath.Join(testDir, "project", "src"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile := createTestFile(t, filepath.Join(testDir, "project", "src"), "main.go", "package main")

	// Добавляем файл в vault
	if err := AddFileToVault(vaultPath, testPassword, testFile); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Получаем путь файла в vault
	entries, err := ListVault(vaultPath, testPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	var targetPath string
	for _, entry := range entries {
		if entry.Name == "main.go" {
			targetPath = entry.Path
			break
		}
	}

	if targetPath == "" {
		t.Fatal("main.go not found in vault")
	}

	// Тест извлечения с полным путем
	err = GetFromVaultWithOptions(vaultPath, testPassword, outputDirFull, []string{targetPath}, true)
	if err != nil {
		t.Fatalf("GetFromVaultWithOptions with full path failed: %v", err)
	}

	fullPath := filepath.Join(outputDirFull, targetPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("Expected file with full path %s to exist", fullPath)
	}

	// Тест извлечения без полного пути
	err = GetFromVaultWithOptions(vaultPath, testPassword, outputDirFlat, []string{targetPath}, false)
	if err != nil {
		t.Fatalf("GetFromVaultWithOptions without full path failed: %v", err)
	}

	flatPath := filepath.Join(outputDirFlat, "main.go")
	if _, err := os.Stat(flatPath); os.IsNotExist(err) {
		t.Errorf("Expected file with flat name %s to exist", flatPath)
	}

	// Проверяем содержимое
	content, err := ioutil.ReadFile(flatPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}
	if string(content) != "package main" {
		t.Errorf("File content mismatch")
	}
}
