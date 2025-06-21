package vault

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test constants
const (
	integrationTestPassword = "IntegrationTest123!"
)

// Test variables
var (
	largeFileContent = "Large content for performance testing. " + strings.Repeat("Data ", 1000)
)

// ========================
// INTEGRATION TESTS
// ========================

// TestCompleteWorkflow тестирует полный рабочий процесс
func TestCompleteWorkflow(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "integration.vault")

	// 1. Создание vault
	if err := CreateVault(vaultPath, integrationTestPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// 2. Проверка пустого vault
	entries, err := ListVault(vaultPath, integrationTestPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("Expected empty vault, got %d entries", len(entries))
	}

	// 3. Добавление файлов
	file1 := createTestFile(t, tmpDir, "document.txt", "Important document content")
	file2 := createTestFile(t, tmpDir, "config.json", `{"setting": "value"}`)

	if err := AddFileToVault(vaultPath, integrationTestPassword, file1); err != nil {
		t.Fatalf("AddFileToVault failed for file1: %v", err)
	}
	if err := AddFileToVault(vaultPath, integrationTestPassword, file2); err != nil {
		t.Fatalf("AddFileToVault failed for file2: %v", err)
	}

	// 4. Проверка списка файлов
	entries, err = ListVault(vaultPath, integrationTestPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// 5. Селективное извлечение
	outputDir1 := filepath.Join(tmpDir, "selective_output")
	if err := GetFromVault(vaultPath, integrationTestPassword, outputDir1, []string{entries[0].Path}); err != nil {
		t.Fatalf("GetFromVault failed: %v", err)
	}

	// 6. Полное извлечение
	outputDir2 := filepath.Join(tmpDir, "full_output")
	if err := ExtractFromVault(vaultPath, integrationTestPassword, outputDir2); err != nil {
		t.Fatalf("ExtractFromVault failed: %v", err)
	}

	// 7. Удаление одного файла
	if err := RemoveFromVault(vaultPath, integrationTestPassword, []string{entries[0].Path}); err != nil {
		t.Fatalf("RemoveFromVault failed: %v", err)
	}

	// 8. Проверка что файл удалён
	entries, err = ListVault(vaultPath, integrationTestPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry after removal, got %d", len(entries))
	}
}

// ========================
// SECURITY TESTS
// ========================

// TestPasswordSecurity тестирует безопасность паролей
func TestPasswordSecurity(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "security.vault")
	testFile := createTestFile(t, tmpDir, "secret.txt", "Top secret information")

	// Тест 1: Создание с сильным паролем
	strongPassword := "VeryStrongPassword123!@#$%^&*()"
	if err := CreateVault(vaultPath, strongPassword); err != nil {
		t.Fatalf("CreateVault failed with strong password: %v", err)
	}

	if err := AddFileToVault(vaultPath, strongPassword, testFile); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Тест 2: Неправильные пароли должны давать ошибку
	wrongPasswords := []string{
		"wrongpassword",
		"VeryStrongPassword123!@#$%^&*(",
		"VeryStrongPassword123!@#$%^&*()X",
		"",
		"verystrongpassword123!@#$%^&*()",
	}

	for i, wrongPass := range wrongPasswords {
		_, err := ListVault(vaultPath, wrongPass)
		if err == nil {
			t.Errorf("Test %d: Expected error with wrong password '%s'", i+1, wrongPass)
		}
	}

	// Тест 3: Правильный пароль должен работать
	entries, err := ListVault(vaultPath, strongPassword)
	if err != nil {
		t.Fatalf("ListVault failed with correct password: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
}

// TestPasswordComplexity тестирует различные типы паролей
func TestPasswordComplexity(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	testPasswords := []struct {
		name     string
		password string
		valid    bool
	}{
		{"Empty password", "", false},
		{"Simple password", "123", true},
		{"Complex password", "MyVeryComplexPassword!@#123", true},
		{"Unicode password", "пароль123", true},
		{"Very long password", strings.Repeat("a", 1000), true},
		{"Special chars", "!@#$%^&*()_+-={}[]|\\:;\"'<>?,./ ", true},
	}

	for i, test := range testPasswords {
		vaultPath := filepath.Join(tmpDir, "vault_"+string(rune('a'+i))+".vault")

		err := CreateVault(vaultPath, test.password)

		if test.valid && err != nil {
			t.Errorf("Test '%s': Expected success but got error: %v", test.name, err)
		} else if !test.valid && err == nil {
			t.Errorf("Test '%s': Expected error but got success", test.name)
		}
	}
}

// ========================
// PERFORMANCE TESTS
// ========================

// TestLargeFileHandling тестирует работу с большими файлами
func TestLargeFileHandling(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "large.vault")

	// Создаём файл размером ~200KB
	largeContent := strings.Repeat("Large file content with repetitive data. ", 5000)
	largeFile := createTestFile(t, tmpDir, "large.txt", largeContent)

	// Создаём vault
	if err := CreateVault(vaultPath, integrationTestPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Измеряем время добавления
	start := time.Now()
	if err := AddFileToVault(vaultPath, integrationTestPassword, largeFile); err != nil {
		t.Fatalf("AddFileToVault failed for large file: %v", err)
	}
	addDuration := time.Since(start)

	// Измеряем время извлечения
	outputDir := filepath.Join(tmpDir, "large_output")
	start = time.Now()
	if err := ExtractFromVault(vaultPath, integrationTestPassword, outputDir); err != nil {
		t.Fatalf("ExtractFromVault failed for large file: %v", err)
	}
	extractDuration := time.Since(start)

	// Проверяем целостность
	entries, err := ListVault(vaultPath, integrationTestPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	extractedPath := filepath.Join(outputDir, entries[0].Path)
	extractedContent, err := ioutil.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(extractedContent) != largeContent {
		t.Fatal("Large file content mismatch after extraction")
	}

	// Проверяем производительность (должно быть разумно быстро)
	t.Logf("Large file processing times: Add=%v, Extract=%v", addDuration, extractDuration)

	if addDuration > 5*time.Second {
		t.Logf("Warning: Adding large file took %v (might be slow)", addDuration)
	}
	if extractDuration > 5*time.Second {
		t.Logf("Warning: Extracting large file took %v (might be slow)", extractDuration)
	}
}

// ========================
// EDGE CASE TESTS
// ========================

// TestSpecialCharacterFiles тестирует файлы со специальными символами
func TestSpecialCharacterFiles(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "special.vault")

	if err := CreateVault(vaultPath, integrationTestPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Тестируем файлы с различными именами (избегаем символы, которые могут быть проблемными в путях)
	specialFiles := []struct {
		name    string
		content string
	}{
		{"file with spaces.txt", "Content with spaces"},
		{"file_with_underscores.txt", "Content with underscores"},
		{"file-with-dashes.txt", "Content with dashes"},
		{"file.multiple.dots.txt", "Content with dots"},
		{"числа123.txt", "Content with numbers"},
	}

	for _, file := range specialFiles {
		filePath := createTestFile(t, tmpDir, file.name, file.content)

		if err := AddFileToVault(vaultPath, integrationTestPassword, filePath); err != nil {
			t.Errorf("Failed to add file '%s': %v", file.name, err)
			continue
		}
	}

	// Проверяем что все файлы добавлены
	entries, err := ListVault(vaultPath, integrationTestPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != len(specialFiles) {
		t.Fatalf("Expected %d entries, got %d", len(specialFiles), len(entries))
	}

	// Извлекаем и проверяем содержимое
	outputDir := filepath.Join(tmpDir, "special_output")
	if err := ExtractFromVault(vaultPath, integrationTestPassword, outputDir); err != nil {
		t.Fatalf("ExtractFromVault failed: %v", err)
	}

	for _, entry := range entries {
		extractedPath := filepath.Join(outputDir, entry.Path)
		if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
			t.Errorf("Extracted file does not exist: %s", extractedPath)
		}
	}
}

// TestEmptyFiles тестирует работу с пустыми файлами
func TestEmptyFiles(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "empty.vault")
	emptyFile := createTestFile(t, tmpDir, "empty.txt", "")

	if err := CreateVault(vaultPath, integrationTestPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Добавляем пустой файл
	if err := AddFileToVault(vaultPath, integrationTestPassword, emptyFile); err != nil {
		t.Fatalf("AddFileToVault failed for empty file: %v", err)
	}

	// Проверяем что файл добавлен
	entries, err := ListVault(vaultPath, integrationTestPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Size != 0 {
		t.Errorf("Expected size 0 for empty file, got %d", entries[0].Size)
	}

	// Извлекаем и проверяем
	outputDir := filepath.Join(tmpDir, "empty_output")
	if err := ExtractFromVault(vaultPath, integrationTestPassword, outputDir); err != nil {
		t.Fatalf("ExtractFromVault failed: %v", err)
	}

	extractedPath := filepath.Join(outputDir, entries[0].Path)
	content, err := ioutil.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted empty file: %v", err)
	}

	if len(content) != 0 {
		t.Errorf("Expected empty content, got %d bytes", len(content))
	}
}

// ========================
// COMPRESSION TESTS
// ========================

// TestCompressionEfficiency тестирует эффективность сжатия
func TestCompressionEfficiency(t *testing.T) {
	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "compression.vault")

	if err := CreateVault(vaultPath, integrationTestPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Тест 1: Хорошо сжимаемые данные
	compressibleContent := strings.Repeat("AAAAAAAAAA", 1000) // 10KB повторяющихся данных
	compressibleFile := createTestFile(t, tmpDir, "compressible.txt", compressibleContent)

	if err := AddFileToVault(vaultPath, integrationTestPassword, compressibleFile); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Тест 2: Плохо сжимаемые данные (псевдо-случайные)
	randomContent := ""
	for i := 0; i < 1000; i++ {
		randomContent += string(rune(33 + (i*7)%94)) // Псевдо-случайные ASCII символы
	}
	randomFile := createTestFile(t, tmpDir, "random.txt", randomContent)

	if err := AddFileToVault(vaultPath, integrationTestPassword, randomFile); err != nil {
		t.Fatalf("AddFileToVault failed: %v", err)
	}

	// Проверяем результаты сжатия
	entries, err := ListVault(vaultPath, integrationTestPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	for _, entry := range entries {
		compressionRatio := float64(entry.CompressedSize) / float64(entry.Size)
		t.Logf("File %s: Original=%d, Compressed=%d, Ratio=%.2f",
			entry.Name, entry.Size, entry.CompressedSize, compressionRatio)

		if entry.Name == "compressible.txt" && compressionRatio > 0.1 {
			t.Logf("Warning: Highly compressible data has ratio %.2f (expected < 0.1)", compressionRatio)
		}
	}
}

// ========================
// STRESS TESTS
// ========================

// TestManySmallFiles тестирует много маленьких файлов
func TestManySmallFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tmpDir := setupCoreTest(t)
	defer cleanupCoreTest(t, tmpDir)

	vaultPath := filepath.Join(tmpDir, "many_files.vault")

	if err := CreateVault(vaultPath, integrationTestPassword); err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Создаём много маленьких файлов
	fileCount := 50
	for i := 0; i < fileCount; i++ {
		fileName := filepath.Join("file_" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)) + ".txt")
		content := "Small file content " + string(rune('0'+i%10))
		filePath := createTestFile(t, tmpDir, fileName, content)

		if err := AddFileToVault(vaultPath, integrationTestPassword, filePath); err != nil {
			t.Fatalf("AddFileToVault failed for file %d: %v", i, err)
		}
	}

	// Проверяем что все файлы добавлены
	entries, err := ListVault(vaultPath, integrationTestPassword)
	if err != nil {
		t.Fatalf("ListVault failed: %v", err)
	}

	if len(entries) != fileCount {
		t.Fatalf("Expected %d entries, got %d", fileCount, len(entries))
	}

	// Извлекаем все файлы
	outputDir := filepath.Join(tmpDir, "many_output")
	if err := ExtractFromVault(vaultPath, integrationTestPassword, outputDir); err != nil {
		t.Fatalf("ExtractFromVault failed: %v", err)
	}

	// Проверяем что все файлы извлечены
	for _, entry := range entries {
		extractedPath := filepath.Join(outputDir, entry.Path)
		if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
			t.Errorf("File not extracted: %s", extractedPath)
		}
	}
}

// ========================
// BENCHMARKS
// ========================

// BenchmarkCompleteWorkflow бенчмарк полного рабочего процесса
func BenchmarkCompleteWorkflow(b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "bench_workflow_*")
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vaultPath := filepath.Join(tmpDir, "bench_"+string(rune('a'+i))+".vault")
		testFile := filepath.Join(tmpDir, "bench_file_"+string(rune('a'+i))+".txt")

		// Создаём тестовый файл
		ioutil.WriteFile(testFile, []byte(largeFileContent), 0644)

		// Создаём vault
		CreateVault(vaultPath, integrationTestPassword)

		// Добавляем файл
		AddFileToVault(vaultPath, integrationTestPassword, testFile)

		// Извлекаем файл
		outputDir := filepath.Join(tmpDir, "bench_output_"+string(rune('a'+i)))
		ExtractFromVault(vaultPath, integrationTestPassword, outputDir)
	}
}

// BenchmarkMultipleFileOperations бенчмарк множественных операций
func BenchmarkMultipleFileOperations(b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "bench_multi_*")
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "bench_multi.vault")
	CreateVault(vaultPath, integrationTestPassword)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tmpDir, "bench_file_"+string(rune('a'+i%26))+".txt")
		ioutil.WriteFile(testFile, []byte("Benchmark content "+string(rune('a'+i%26))), 0644)

		AddFileToVault(vaultPath, integrationTestPassword, testFile)
	}
}
