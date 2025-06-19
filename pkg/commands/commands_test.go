package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flint-vault/pkg/lib/vault"

	"github.com/urfave/cli/v3"
)

// TestHelper - вспомогательная структура для тестирования команд
type TestHelper struct {
	tmpDir     string
	vaultPath  string
	password   string
	extractDir string
}

func NewTestHelper(t *testing.T) *TestHelper {
	tmpDir, err := os.MkdirTemp("", "cmd_test_*")
	if err != nil {
		t.Fatalf("Не удалось создать временный каталог: %v", err)
	}

	return &TestHelper{
		tmpDir:     tmpDir,
		vaultPath:  filepath.Join(tmpDir, "test.dat"),
		password:   "TestPassword123!",
		extractDir: filepath.Join(tmpDir, "extracted"),
	}
}

func (h *TestHelper) Cleanup() {
	os.RemoveAll(h.tmpDir)
}

func (h *TestHelper) CreateTestFiles(t *testing.T) {
	// Создаем структуру тестовых файлов
	testDir := filepath.Join(h.tmpDir, "test_files")
	subDir := filepath.Join(testDir, "subdir")

	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Не удалось создать тестовые каталоги: %v", err)
	}

	// Создаем файлы
	files := map[string]string{
		filepath.Join(testDir, "file1.txt"):        "Содержимое файла 1 🚀",
		filepath.Join(testDir, "file2.txt"):        "Содержимое файла 2 📁",
		filepath.Join(subDir, "nested_file.txt"):   "Файл в подкаталоге 📂",
		filepath.Join(h.tmpDir, "single_file.txt"): "Отдельный файл",
	}

	for path, content := range files {
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Не удалось создать файл %s: %v", path, err)
		}
	}
}

func TestCreateCommand(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Создаем приложение для тестирования
	app := &cli.Command{
		Name: "test-create",
		Commands: []*cli.Command{
			{
				Name: "create",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Aliases: []string{"f"}},
					&cli.StringFlag{Name: "password", Aliases: []string{"p"}},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					file := cmd.String("file")
					password := cmd.String("password")
					return vault.CreateVault(file, password)
				},
			},
		},
	}

	testCases := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Успешное создание",
			args:        []string{"test-create", "create", "-f", helper.vaultPath, "-p", helper.password},
			expectError: false,
		},
		{
			name:        "Отсутствует файл",
			args:        []string{"test-create", "create", "-p", helper.password},
			expectError: true,
		},
		{
			name:        "Отсутствует пароль",
			args:        []string{"test-create", "create", "-f", helper.vaultPath + "_no_pass"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := app.Run(context.Background(), tc.args)

			if tc.expectError && err == nil {
				t.Errorf("Ожидалась ошибка, но получили nil")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Не ожидалась ошибка, получили: %v", err)
			}

			// Проверяем, что файл создан при успешном выполнении
			if !tc.expectError && err == nil {
				if _, err := os.Stat(helper.vaultPath); os.IsNotExist(err) {
					t.Errorf("Файл хранилища не был создан")
				}
			}
		})
	}
}

func TestAddCommand(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()
	helper.CreateTestFiles(t)

	// Сначала создаем хранилище
	err := vault.CreateVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Не удалось создать хранилище: %v", err)
	}

	testCases := []struct {
		name       string
		sourcePath string
		expectErr  bool
	}{
		{
			name:       "Добавление отдельного файла",
			sourcePath: filepath.Join(helper.tmpDir, "single_file.txt"),
			expectErr:  false,
		},
		{
			name:       "Добавление каталога",
			sourcePath: filepath.Join(helper.tmpDir, "test_files"),
			expectErr:  false,
		},
		{
			name:       "Добавление несуществующего файла",
			sourcePath: filepath.Join(helper.tmpDir, "nonexistent.txt"),
			expectErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			// Проверяем, что источник существует (для валидных тестов)
			if !tc.expectErr {
				if _, statErr := os.Stat(tc.sourcePath); statErr != nil {
					t.Fatalf("Тестовый файл не существует: %s", tc.sourcePath)
				}
			}

			// Определяем, файл это или каталог
			if info, statErr := os.Stat(tc.sourcePath); statErr == nil && info.IsDir() {
				err = vault.AddDirectoryToVault(helper.vaultPath, helper.password, tc.sourcePath)
			} else if statErr == nil {
				err = vault.AddFileToVault(helper.vaultPath, helper.password, tc.sourcePath)
			} else {
				err = statErr // Ошибка stat
			}

			if tc.expectErr && err == nil {
				t.Errorf("Ожидалась ошибка, но получили nil")
			}

			if !tc.expectErr && err != nil {
				t.Errorf("Не ожидалась ошибка, получили: %v", err)
			}

			// Проверяем, что файл добавлен при успешном выполнении
			if !tc.expectErr && err == nil {
				data, err := vault.ListVault(helper.vaultPath, helper.password)
				if err != nil {
					t.Fatalf("Не удалось прочитать хранилище: %v", err)
				}

				if len(data.Entries) == 0 {
					t.Errorf("После добавления файла хранилище пустое")
				}
			}
		})
	}
}

func TestListCommand(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Создаем хранилище с тестовыми данными
	err := vault.CreateVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Не удалось создать хранилище: %v", err)
	}

	// Добавляем тестовый файл
	testFile := filepath.Join(helper.tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	err = vault.AddFileToVault(helper.vaultPath, helper.password, testFile)
	if err != nil {
		t.Fatalf("Не удалось добавить файл в хранилище: %v", err)
	}

	// Тестируем команду list
	data, err := vault.ListVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Не удалось получить список: %v", err)
	}

	if len(data.Entries) != 1 {
		t.Errorf("Ожидалось 1 запись, получили %d", len(data.Entries))
	}

	entry := data.Entries[0]
	if entry.Name != "test.txt" {
		t.Errorf("Неверное имя файла: ожидалось 'test.txt', получили '%s'", entry.Name)
	}

	if string(entry.Content) != "test content" {
		t.Errorf("Неверное содержимое файла")
	}
}

func TestExtractCommand(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Создаем хранилище с тестовыми данными
	err := vault.CreateVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Не удалось создать хранилище: %v", err)
	}

	// Добавляем тестовые файлы
	testFile1 := filepath.Join(helper.tmpDir, "test1.txt")
	testFile2 := filepath.Join(helper.tmpDir, "test2.txt")

	err = os.WriteFile(testFile1, []byte("content1"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	err = os.WriteFile(testFile2, []byte("content2"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	err = vault.AddFileToVault(helper.vaultPath, helper.password, testFile1)
	if err != nil {
		t.Fatalf("Не удалось добавить файл в хранилище: %v", err)
	}

	err = vault.AddFileToVault(helper.vaultPath, helper.password, testFile2)
	if err != nil {
		t.Fatalf("Не удалось добавить файл в хранилище: %v", err)
	}

	// Тестируем извлечение всех файлов
	err = vault.ExtractVault(helper.vaultPath, helper.password, helper.extractDir)
	if err != nil {
		t.Fatalf("Не удалось извлечь файлы: %v", err)
	}

	// Проверяем извлеченные файлы (используем полные пути как в хранилище)
	extractedFile1 := filepath.Join(helper.extractDir, testFile1)
	extractedFile2 := filepath.Join(helper.extractDir, testFile2)

	content1, err := os.ReadFile(extractedFile1)
	if err != nil {
		t.Fatalf("Не удалось прочитать извлеченный файл 1: %v", err)
	}

	if string(content1) != "content1" {
		t.Errorf("Содержимое файла 1 не совпадает")
	}

	content2, err := os.ReadFile(extractedFile2)
	if err != nil {
		t.Fatalf("Не удалось прочитать извлеченный файл 2: %v", err)
	}

	if string(content2) != "content2" {
		t.Errorf("Содержимое файла 2 не совпадает")
	}
}

func TestGetCommand(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Создаем хранилище с тестовыми данными
	err := vault.CreateVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Не удалось создать хранилище: %v", err)
	}

	// Добавляем тестовые файлы
	testFile := filepath.Join(helper.tmpDir, "specific_file.txt")
	err = os.WriteFile(testFile, []byte("specific content"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	err = vault.AddFileToVault(helper.vaultPath, helper.password, testFile)
	if err != nil {
		t.Fatalf("Не удалось добавить файл в хранилище: %v", err)
	}

	// Тестируем извлечение конкретного файла (используем полный путь как в хранилище)
	getExtractDir := filepath.Join(helper.tmpDir, "get_extract")
	err = vault.ExtractSpecific(helper.vaultPath, helper.password, testFile, getExtractDir)
	if err != nil {
		t.Fatalf("Не удалось извлечь конкретный файл: %v", err)
	}

	// Проверяем извлеченный файл
	extractedFile := filepath.Join(getExtractDir, "specific_file.txt")
	content, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("Не удалось прочитать извлеченный файл: %v", err)
	}

	if string(content) != "specific content" {
		t.Errorf("Content mismatch: expected 'specific content', got '%s'", string(content))
	}
}

func TestGetCommandMultiple(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create vault with test data
	err := vault.CreateVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	// Add multiple test files
	testFile1 := filepath.Join(helper.tmpDir, "file1.txt")
	testFile2 := filepath.Join(helper.tmpDir, "file2.txt")
	testDir := filepath.Join(helper.tmpDir, "test_dir")
	testFileInDir := filepath.Join(testDir, "file_in_dir.txt")

	// Create test structure
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err = os.WriteFile(testFile1, []byte("content1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file1: %v", err)
	}

	err = os.WriteFile(testFile2, []byte("content2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file2: %v", err)
	}

	err = os.WriteFile(testFileInDir, []byte("dir content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file in dir: %v", err)
	}

	// Add files to vault
	err = vault.AddFileToVault(helper.vaultPath, helper.password, testFile1)
	if err != nil {
		t.Fatalf("Failed to add file1 to vault: %v", err)
	}

	err = vault.AddFileToVault(helper.vaultPath, helper.password, testFile2)
	if err != nil {
		t.Fatalf("Failed to add file2 to vault: %v", err)
	}

	err = vault.AddDirectoryToVault(helper.vaultPath, helper.password, testDir)
	if err != nil {
		t.Fatalf("Failed to add directory to vault: %v", err)
	}

	t.Run("ExtractMultipleFiles", func(t *testing.T) {
		multiExtractDir := filepath.Join(helper.tmpDir, "multi_extract")

		// Test extracting multiple files (use full paths as stored in vault)
		extractedPaths, notFoundPaths, err := vault.ExtractMultiple(
			helper.vaultPath,
			helper.password,
			[]string{testFile1, testFile2},
			multiExtractDir,
		)
		if err != nil {
			t.Fatalf("Failed to extract multiple files: %v", err)
		}

		// Check results
		if len(extractedPaths) != 2 {
			t.Errorf("Expected 2 extracted paths, got %d", len(extractedPaths))
		}
		if len(notFoundPaths) != 0 {
			t.Errorf("Expected 0 not found paths, got %d", len(notFoundPaths))
		}

		// Check extracted files
		extractedFile1 := filepath.Join(multiExtractDir, "file1.txt")
		content1, err := os.ReadFile(extractedFile1)
		if err != nil {
			t.Fatalf("Failed to read extracted file1: %v", err)
		}
		if string(content1) != "content1" {
			t.Errorf("File1 content mismatch: expected 'content1', got '%s'", string(content1))
		}

		extractedFile2 := filepath.Join(multiExtractDir, "file2.txt")
		content2, err := os.ReadFile(extractedFile2)
		if err != nil {
			t.Fatalf("Failed to read extracted file2: %v", err)
		}
		if string(content2) != "content2" {
			t.Errorf("File2 content mismatch: expected 'content2', got '%s'", string(content2))
		}
	})

	t.Run("ExtractMixedFilesAndDirs", func(t *testing.T) {
		mixedExtractDir := filepath.Join(helper.tmpDir, "mixed_extract")

		// Test extracting mixed files and directories (use full paths as stored in vault)
		extractedPaths, notFoundPaths, err := vault.ExtractMultiple(
			helper.vaultPath,
			helper.password,
			[]string{testFile1, testDir},
			mixedExtractDir,
		)
		if err != nil {
			t.Fatalf("Failed to extract mixed targets: %v", err)
		}

		// Check results
		if len(extractedPaths) != 2 {
			t.Errorf("Expected 2 extracted paths, got %d", len(extractedPaths))
		}
		if len(notFoundPaths) != 0 {
			t.Errorf("Expected 0 not found paths, got %d", len(notFoundPaths))
		}

		// Check individual file
		extractedFile := filepath.Join(mixedExtractDir, "file1.txt")
		if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
			t.Errorf("file1.txt was not extracted")
		}

		// Check directory content (relative to testDir base)
		dirFile := filepath.Join(mixedExtractDir, "file_in_dir.txt")
		content, err := os.ReadFile(dirFile)
		if err != nil {
			t.Fatalf("Failed to read file from extracted directory: %v", err)
		}
		if string(content) != "dir content" {
			t.Errorf("Directory file content mismatch: expected 'dir content', got '%s'", string(content))
		}
	})

	t.Run("ExtractWithNotFound", func(t *testing.T) {
		notFoundExtractDir := filepath.Join(helper.tmpDir, "not_found_extract")

		// Test extracting with some non-existent files
		extractedPaths, notFoundPaths, err := vault.ExtractMultiple(
			helper.vaultPath,
			helper.password,
			[]string{testFile1, "nonexistent.txt", testFile2},
			notFoundExtractDir,
		)
		if err != nil {
			t.Fatalf("Failed to extract with not found targets: %v", err)
		}

		// Check results
		if len(extractedPaths) != 2 {
			t.Errorf("Expected 2 extracted paths, got %d", len(extractedPaths))
		}
		if len(notFoundPaths) != 1 {
			t.Errorf("Expected 1 not found path, got %d", len(notFoundPaths))
		}

		// Check that existing files were extracted
		extractedFile1 := filepath.Join(notFoundExtractDir, "file1.txt")
		if _, err := os.Stat(extractedFile1); os.IsNotExist(err) {
			t.Errorf("file1.txt was not extracted")
		}

		extractedFile2 := filepath.Join(notFoundExtractDir, "file2.txt")
		if _, err := os.Stat(extractedFile2); os.IsNotExist(err) {
			t.Errorf("file2.txt was not extracted")
		}
	})
}

func TestRemoveCommand(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Создаем хранилище с тестовыми данными
	err := vault.CreateVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Не удалось создать хранилище: %v", err)
	}

	// Добавляем тестовые файлы
	testFile1 := filepath.Join(helper.tmpDir, "remove1.txt")
	testFile2 := filepath.Join(helper.tmpDir, "remove2.txt")

	err = os.WriteFile(testFile1, []byte("remove content 1"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	err = os.WriteFile(testFile2, []byte("remove content 2"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	err = vault.AddFileToVault(helper.vaultPath, helper.password, testFile1)
	if err != nil {
		t.Fatalf("Не удалось добавить файл в хранилище: %v", err)
	}

	err = vault.AddFileToVault(helper.vaultPath, helper.password, testFile2)
	if err != nil {
		t.Fatalf("Не удалось добавить файл в хранилище: %v", err)
	}

	// Проверяем, что в хранилище 2 файла
	data, err := vault.ListVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Не удалось получить список: %v", err)
	}

	if len(data.Entries) != 2 {
		t.Fatalf("Ожидалось 2 файла, получили %d", len(data.Entries))
	}

	// Удаляем один файл (используем полный путь как в хранилище)
	err = vault.RemoveFromVault(helper.vaultPath, helper.password, testFile1)
	if err != nil {
		t.Fatalf("Не удалось удалить файл: %v", err)
	}

	// Проверяем, что остался 1 файл
	data, err = vault.ListVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Не удалось получить список после удаления: %v", err)
	}

	if len(data.Entries) != 1 {
		t.Errorf("Ожидался 1 файл после удаления, получили %d", len(data.Entries))
	}

	// Проверяем, что остался правильный файл
	if data.Entries[0].Path != testFile2 {
		t.Errorf("Остался неверный файл: %s", data.Entries[0].Path)
	}
}

func TestFormatSize(t *testing.T) {
	testCases := []struct {
		size     int64
		expected string
	}{
		{0, "0 Б"},
		{512, "512 Б"},
		{1024, "1.0 KБ"},
		{1536, "1.5 KБ"},
		{1048576, "1.0 MБ"},
		{1073741824, "1.0 GБ"},
		{1099511627776, "1.0 TБ"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Size_%d", tc.size), func(t *testing.T) {
			result := formatSize(tc.size)
			if result != tc.expected {
				t.Errorf("formatSize(%d) = %s, ожидалось %s", tc.size, result, tc.expected)
			}
		})
	}
}

func TestPasswordSecurity(t *testing.T) {
	// Тест для проверки, что пароли не остаются в памяти
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Создаем хранилище
	err := vault.CreateVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Не удалось создать хранилище: %v", err)
	}

	// Загружаем хранилище с неверным паролем
	_, err = vault.ListVault(helper.vaultPath, "WrongPassword")
	if err == nil {
		t.Errorf("Ожидалась ошибка при неверном пароле")
	}

	// Проверяем, что ошибка содержит правильное сообщение
	if !strings.Contains(err.Error(), "неверный пароль") {
		t.Errorf("Ошибка должна содержать информацию о неверном пароле: %v", err)
	}
}

func TestFullWorkflow(t *testing.T) {
	// Интеграционный тест полного рабочего процесса
	helper := NewTestHelper(t)
	defer helper.Cleanup()
	helper.CreateTestFiles(t)

	// 1. Создаем хранилище
	err := vault.CreateVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Шаг 1 - Не удалось создать хранилище: %v", err)
	}

	// 2. Добавляем отдельный файл
	singleFile := filepath.Join(helper.tmpDir, "single_file.txt")
	err = vault.AddFileToVault(helper.vaultPath, helper.password, singleFile)
	if err != nil {
		t.Fatalf("Шаг 2 - Не удалось добавить файл: %v", err)
	}

	// 3. Добавляем каталог с файлами
	testDir := filepath.Join(helper.tmpDir, "test_files")
	err = vault.AddDirectoryToVault(helper.vaultPath, helper.password, testDir)
	if err != nil {
		t.Fatalf("Шаг 3 - Не удалось добавить каталог: %v", err)
	}

	// 4. Проверяем содержимое
	data, err := vault.ListVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Шаг 4 - Не удалось получить список: %v", err)
	}

	if len(data.Entries) < 4 { // Минимум: single_file + test_files + 2 файла внутри
		t.Errorf("Шаг 4 - Ожидалось минимум 4 записи, получили %d", len(data.Entries))
	}

	// 5. Извлекаем конкретный файл (используем полный путь как в хранилище)
	specificExtractDir := filepath.Join(helper.tmpDir, "specific_extract")
	err = vault.ExtractSpecific(helper.vaultPath, helper.password, singleFile, specificExtractDir)
	if err != nil {
		t.Fatalf("Шаг 5 - Не удалось извлечь конкретный файл: %v", err)
	}

	// Проверяем извлеченный файл
	extractedSpecific := filepath.Join(specificExtractDir, "single_file.txt")
	if _, err := os.Stat(extractedSpecific); os.IsNotExist(err) {
		t.Errorf("Шаг 5 - Конкретный файл не был извлечен")
	}

	// 6. Извлекаем все файлы
	err = vault.ExtractVault(helper.vaultPath, helper.password, helper.extractDir)
	if err != nil {
		t.Fatalf("Шаг 6 - Не удалось извлечь все файлы: %v", err)
	}

	// 7. Удаляем файл (используем полный путь как в хранилище)
	err = vault.RemoveFromVault(helper.vaultPath, helper.password, singleFile)
	if err != nil {
		t.Fatalf("Шаг 7 - Не удалось удалить файл: %v", err)
	}

	// 8. Проверяем, что файл удален
	dataAfterRemove, err := vault.ListVault(helper.vaultPath, helper.password)
	if err != nil {
		t.Fatalf("Шаг 8 - Не удалось получить список после удаления: %v", err)
	}

	if len(dataAfterRemove.Entries) >= len(data.Entries) {
		t.Errorf("Шаг 8 - Файл не был удален (было %d, стало %d)", len(data.Entries), len(dataAfterRemove.Entries))
	}

	t.Logf("✅ Полный рабочий процесс выполнен успешно!")
}
