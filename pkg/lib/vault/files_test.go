package vault

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAddFileToVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Не удалось создать временный каталог: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.dat")
	password := "password123"

	// Создаем хранилище
	err = CreateVault(vaultPath, password)
	if err != nil {
		t.Fatalf("Не удалось создать хранилище: %v", err)
	}

	// Создаем тестовый файл
	testFilePath := filepath.Join(tmpDir, "test_file.txt")
	testContent := "Это тестовый файл с русским текстом! 🚀"
	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	// Добавляем файл в хранилище
	err = AddFileToVault(vaultPath, password, testFilePath)
	if err != nil {
		t.Fatalf("Не удалось добавить файл в хранилище: %v", err)
	}

	// Проверяем, что файл был добавлен
	data, err := loadVaultData(vaultPath, password)
	if err != nil {
		t.Fatalf("Не удалось загрузить хранилище: %v", err)
	}

	if len(data.Entries) != 1 {
		t.Fatalf("Ожидалось 1 запись, получили %d", len(data.Entries))
	}

	entry := data.Entries[0]
	if entry.Path != testFilePath {
		t.Errorf("Неверный путь: ожидался %s, получили %s", testFilePath, entry.Path)
	}

	if entry.Name != "test_file.txt" {
		t.Errorf("Неверное имя файла: ожидалось 'test_file.txt', получили '%s'", entry.Name)
	}

	if entry.IsDir {
		t.Errorf("Файл отмечен как каталог")
	}

	if string(entry.Content) != testContent {
		t.Errorf("Содержимое файла не совпадает:\nОжидалось: %s\nПолучили: %s",
			testContent, string(entry.Content))
	}

	if entry.Size != int64(len(testContent)) {
		t.Errorf("Размер файла не совпадает: ожидался %d, получили %d",
			len(testContent), entry.Size)
	}
}

func TestAddFileToVaultUpdateExisting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Не удалось создать временный каталог: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.dat")
	password := "password123"
	testFilePath := filepath.Join(tmpDir, "test_file.txt")

	// Создаем хранилище
	err = CreateVault(vaultPath, password)
	if err != nil {
		t.Fatalf("Не удалось создать хранилище: %v", err)
	}

	// Создаем и добавляем первую версию файла
	content1 := "Первая версия файла"
	err = os.WriteFile(testFilePath, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	err = AddFileToVault(vaultPath, password, testFilePath)
	if err != nil {
		t.Fatalf("Не удалось добавить файл в хранилище: %v", err)
	}

	// Обновляем файл и добавляем снова
	time.Sleep(100 * time.Millisecond) // Чтобы время модификации отличалось
	content2 := "Обновленная версия файла"
	err = os.WriteFile(testFilePath, []byte(content2), 0644)
	if err != nil {
		t.Fatalf("Не удалось обновить тестовый файл: %v", err)
	}

	err = AddFileToVault(vaultPath, password, testFilePath)
	if err != nil {
		t.Fatalf("Не удалось обновить файл в хранилище: %v", err)
	}

	// Проверяем, что в хранилище только одна запись с обновленным содержимым
	data, err := loadVaultData(vaultPath, password)
	if err != nil {
		t.Fatalf("Не удалось загрузить хранилище: %v", err)
	}

	if len(data.Entries) != 1 {
		t.Fatalf("Ожидалось 1 запись, получили %d", len(data.Entries))
	}

	entry := data.Entries[0]
	if string(entry.Content) != content2 {
		t.Errorf("Содержимое файла не обновилось:\nОжидалось: %s\nПолучили: %s",
			content2, string(entry.Content))
	}
}

func TestAddDirectoryToVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Не удалось создать временный каталог: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.dat")
	password := "password123"

	// Создаем хранилище
	err = CreateVault(vaultPath, password)
	if err != nil {
		t.Fatalf("Не удалось создать хранилище: %v", err)
	}

	// Создаем структуру каталогов и файлов
	testDir := filepath.Join(tmpDir, "test_directory")
	subDir := filepath.Join(testDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Не удалось создать тестовые каталоги: %v", err)
	}

	// Создаем файлы
	file1 := filepath.Join(testDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")

	err = os.WriteFile(file1, []byte("Содержимое файла 1"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать файл 1: %v", err)
	}

	err = os.WriteFile(file2, []byte("Содержимое файла 2"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать файл 2: %v", err)
	}

	// Добавляем каталог в хранилище
	err = AddDirectoryToVault(vaultPath, password, testDir)
	if err != nil {
		t.Fatalf("Не удалось добавить каталог в хранилище: %v", err)
	}

	// Проверяем содержимое хранилища
	data, err := loadVaultData(vaultPath, password)
	if err != nil {
		t.Fatalf("Не удалось загрузить хранилище: %v", err)
	}

	// Должно быть 4 записи: корневой каталог, подкаталог и 2 файла
	expectedEntries := 4
	if len(data.Entries) != expectedEntries {
		t.Fatalf("Ожидалось %d записей, получили %d", expectedEntries, len(data.Entries))
	}

	// Проверяем, что все ожидаемые пути присутствуют
	expectedPaths := map[string]bool{
		"test_directory":                  true,
		"test_directory/subdir":           true,
		"test_directory/file1.txt":        true,
		"test_directory/subdir/file2.txt": true,
	}

	for _, entry := range data.Entries {
		if !expectedPaths[entry.Path] {
			t.Errorf("Неожиданный путь в хранилище: %s", entry.Path)
		}
		delete(expectedPaths, entry.Path)
	}

	if len(expectedPaths) > 0 {
		t.Errorf("Отсутствуют ожидаемые пути: %v", expectedPaths)
	}
}

func TestExtractVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Не удалось создать временный каталог: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.dat")
	password := "password123"
	extractDir := filepath.Join(tmpDir, "extracted")

	// Создаем хранилище с тестовыми данными
	testData := VaultData{
		Entries: []VaultEntry{
			{
				Path:    "test_dir",
				Name:    "test_dir",
				IsDir:   true,
				Size:    0,
				Mode:    0755,
				ModTime: time.Now(),
				Content: nil,
			},
			{
				Path:    "test_dir/file.txt",
				Name:    "file.txt",
				IsDir:   false,
				Size:    13,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("test content!"),
			},
			{
				Path:    "root_file.txt",
				Name:    "root_file.txt",
				IsDir:   false,
				Size:    10,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("root file!"),
			},
		},
		CreatedAt: time.Now(),
		Comment:   "Test vault",
	}

	err = saveVaultData(vaultPath, password, testData)
	if err != nil {
		t.Fatalf("Не удалось сохранить тестовые данные: %v", err)
	}

	// Извлекаем все файлы
	err = ExtractVault(vaultPath, password, extractDir)
	if err != nil {
		t.Fatalf("Не удалось извлечь файлы: %v", err)
	}

	// Проверяем извлеченные файлы
	extractedFile1 := filepath.Join(extractDir, "test_dir/file.txt")
	extractedFile2 := filepath.Join(extractDir, "root_file.txt")

	// Проверяем, что каталог создан
	if info, err := os.Stat(filepath.Join(extractDir, "test_dir")); err != nil || !info.IsDir() {
		t.Errorf("Каталог test_dir не был создан")
	}

	// Проверяем содержимое файлов
	content1, err := os.ReadFile(extractedFile1)
	if err != nil {
		t.Fatalf("Не удалось прочитать извлеченный файл 1: %v", err)
	}
	if string(content1) != "test content!" {
		t.Errorf("Содержимое файла 1 не совпадает: ожидалось 'test content!', получили '%s'", string(content1))
	}

	content2, err := os.ReadFile(extractedFile2)
	if err != nil {
		t.Fatalf("Не удалось прочитать извлеченный файл 2: %v", err)
	}
	if string(content2) != "root file!" {
		t.Errorf("Содержимое файла 2 не совпадает: ожидалось 'root file!', получили '%s'", string(content2))
	}
}

func TestExtractSpecific(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Не удалось создать временный каталог: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.dat")
	password := "password123"

	// Создаем хранилище с тестовыми данными
	testData := VaultData{
		Entries: []VaultEntry{
			{
				Path:    "dir1",
				Name:    "dir1",
				IsDir:   true,
				Size:    0,
				Mode:    0755,
				ModTime: time.Now(),
				Content: nil,
			},
			{
				Path:    "dir1/file1.txt",
				Name:    "file1.txt",
				IsDir:   false,
				Size:    8,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("file1!!!"),
			},
			{
				Path:    "dir2",
				Name:    "dir2",
				IsDir:   true,
				Size:    0,
				Mode:    0755,
				ModTime: time.Now(),
				Content: nil,
			},
			{
				Path:    "dir2/file2.txt",
				Name:    "file2.txt",
				IsDir:   false,
				Size:    8,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("file2!!!"),
			},
			{
				Path:    "single_file.txt",
				Name:    "single_file.txt",
				IsDir:   false,
				Size:    11,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("single file"),
			},
		},
		CreatedAt: time.Now(),
		Comment:   "Test vault",
	}

	err = saveVaultData(vaultPath, password, testData)
	if err != nil {
		t.Fatalf("Не удалось сохранить тестовые данные: %v", err)
	}

	t.Run("ExtractSingleFile", func(t *testing.T) {
		extractDir := filepath.Join(tmpDir, "extract_single")

		err = ExtractSpecific(vaultPath, password, "single_file.txt", extractDir)
		if err != nil {
			t.Fatalf("Не удалось извлечь файл: %v", err)
		}

		// Проверяем, что файл извлечен
		extractedFile := filepath.Join(extractDir, "single_file.txt")
		content, err := os.ReadFile(extractedFile)
		if err != nil {
			t.Fatalf("Не удалось прочитать извлеченный файл: %v", err)
		}

		if string(content) != "single file" {
			t.Errorf("Содержимое не совпадает: ожидалось 'single file', получили '%s'", string(content))
		}
	})

	t.Run("ExtractDirectory", func(t *testing.T) {
		extractDir := filepath.Join(tmpDir, "extract_dir")

		err = ExtractSpecific(vaultPath, password, "dir1", extractDir)
		if err != nil {
			t.Fatalf("Не удалось извлечь каталог: %v", err)
		}

		// Проверяем, что файл из каталога извлечен
		extractedFile := filepath.Join(extractDir, "file1.txt")
		content, err := os.ReadFile(extractedFile)
		if err != nil {
			t.Fatalf("Не удалось прочитать извлеченный файл: %v", err)
		}

		if string(content) != "file1!!!" {
			t.Errorf("Содержимое не совпадает: ожидалось 'file1!!!', получили '%s'", string(content))
		}
	})

	t.Run("ExtractNonExistent", func(t *testing.T) {
		extractDir := filepath.Join(tmpDir, "extract_nonexistent")

		err = ExtractSpecific(vaultPath, password, "nonexistent_file.txt", extractDir)
		if err == nil {
			t.Errorf("Ожидалась ошибка при извлечении несуществующего файла")
		}

		expectedError := "файл или каталог 'nonexistent_file.txt' не найден в хранилище"
		if err.Error() != expectedError {
			t.Errorf("Ожидалась ошибка '%s', получили '%s'", expectedError, err.Error())
		}
	})
}

func TestRemoveFromVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Не удалось создать временный каталог: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.dat")
	password := "password123"

	// Создаем хранилище с тестовыми данными
	testData := VaultData{
		Entries: []VaultEntry{
			{
				Path:    "dir1",
				Name:    "dir1",
				IsDir:   true,
				Size:    0,
				Mode:    0755,
				ModTime: time.Now(),
				Content: nil,
			},
			{
				Path:    "dir1/file1.txt",
				Name:    "file1.txt",
				IsDir:   false,
				Size:    8,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("content1"),
			},
			{
				Path:    "dir1/file2.txt",
				Name:    "file2.txt",
				IsDir:   false,
				Size:    8,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("content2"),
			},
			{
				Path:    "standalone_file.txt",
				Name:    "standalone_file.txt",
				IsDir:   false,
				Size:    10,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("standalone"),
			},
		},
		CreatedAt: time.Now(),
		Comment:   "Test vault",
	}

	err = saveVaultData(vaultPath, password, testData)
	if err != nil {
		t.Fatalf("Не удалось сохранить тестовые данные: %v", err)
	}

	t.Run("RemoveFile", func(t *testing.T) {
		err = RemoveFromVault(vaultPath, password, "standalone_file.txt")
		if err != nil {
			t.Fatalf("Не удалось удалить файл: %v", err)
		}

		// Проверяем, что файл удален
		data, err := loadVaultData(vaultPath, password)
		if err != nil {
			t.Fatalf("Не удалось загрузить хранилище: %v", err)
		}

		for _, entry := range data.Entries {
			if entry.Path == "standalone_file.txt" {
				t.Errorf("Файл не был удален")
			}
		}

		// Проверяем, что остальные файлы остались
		if len(data.Entries) != 3 {
			t.Errorf("Ожидалось 3 записи после удаления, получили %d", len(data.Entries))
		}
	})

	t.Run("RemoveDirectory", func(t *testing.T) {
		err = RemoveFromVault(vaultPath, password, "dir1")
		if err != nil {
			t.Fatalf("Не удалось удалить каталог: %v", err)
		}

		// Проверяем, что каталог и все его содержимое удалено
		data, err := loadVaultData(vaultPath, password)
		if err != nil {
			t.Fatalf("Не удалось загрузить хранилище: %v", err)
		}

		for _, entry := range data.Entries {
			if entry.Path == "dir1" || entry.Path == "dir1/file1.txt" || entry.Path == "dir1/file2.txt" {
				t.Errorf("Элемент каталога не был удален: %s", entry.Path)
			}
		}

		// После удаления каталога со всем содержимым должно остаться 0 записей
		if len(data.Entries) != 0 {
			t.Errorf("Ожидалось 0 записей после удаления каталога, получили %d", len(data.Entries))
		}
	})

	t.Run("RemoveNonExistent", func(t *testing.T) {
		err = RemoveFromVault(vaultPath, password, "nonexistent_file.txt")
		if err == nil {
			t.Errorf("Ожидалась ошибка при удалении несуществующего файла")
		}

		expectedError := "файл или каталог 'nonexistent_file.txt' не найден в хранилище"
		if err.Error() != expectedError {
			t.Errorf("Ожидалась ошибка '%s', получили '%s'", expectedError, err.Error())
		}
	})
}

func TestGetFileFromVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Не удалось создать временный каталог: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.dat")
	password := "password123"

	// Создаем хранилище с тестовыми данными
	testContent := "Содержимое тестового файла 📁"
	testData := VaultData{
		Entries: []VaultEntry{
			{
				Path:    "test_file.txt",
				Name:    "test_file.txt",
				IsDir:   false,
				Size:    int64(len(testContent)),
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte(testContent),
			},
			{
				Path:    "test_dir",
				Name:    "test_dir",
				IsDir:   true,
				Size:    0,
				Mode:    0755,
				ModTime: time.Now(),
				Content: nil,
			},
		},
		CreatedAt: time.Now(),
		Comment:   "Test vault",
	}

	err = saveVaultData(vaultPath, password, testData)
	if err != nil {
		t.Fatalf("Не удалось сохранить тестовые данные: %v", err)
	}

	t.Run("GetExistingFile", func(t *testing.T) {
		content, err := GetFileFromVault(vaultPath, password, "test_file.txt")
		if err != nil {
			t.Fatalf("Не удалось получить файл из хранилища: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("Содержимое не совпадает:\nОжидалось: %s\nПолучили: %s",
				testContent, string(content))
		}
	})

	t.Run("GetNonExistentFile", func(t *testing.T) {
		_, err := GetFileFromVault(vaultPath, password, "nonexistent.txt")
		if err == nil {
			t.Errorf("Ожидалась ошибка при попытке получить несуществующий файл")
		}

		expectedError := "файл 'nonexistent.txt' не найден в хранилище"
		if err.Error() != expectedError {
			t.Errorf("Ожидалась ошибка '%s', получили '%s'", expectedError, err.Error())
		}
	})

	t.Run("GetDirectory", func(t *testing.T) {
		_, err := GetFileFromVault(vaultPath, password, "test_dir")
		if err == nil {
			t.Errorf("Ожидалась ошибка при попытке получить каталог как файл")
		}

		expectedError := "файл 'test_dir' не найден в хранилище"
		if err.Error() != expectedError {
			t.Errorf("Ожидалась ошибка '%s', получили '%s'", expectedError, err.Error())
		}
	})
}

func TestCompressDecompressData(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "Простой текст",
			data: []byte("Hello, world!"),
		},
		{
			name: "Пустые данные",
			data: []byte(""),
		},
		{
			name: "Русский текст",
			data: []byte("Привет, мир! 🌍"),
		},
		{
			name: "Большие данные",
			data: bytes.Repeat([]byte("Test data "), 1000),
		},
		{
			name: "Бинарные данные",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Сжимаем данные
			compressed, err := compressData(tc.data)
			if err != nil {
				t.Fatalf("Не удалось сжать данные: %v", err)
			}

			// Разжимаем данные
			decompressed, err := decompressData(compressed)
			if err != nil {
				t.Fatalf("Не удалось разжать данные: %v", err)
			}

			// Сравниваем результат
			if !bytes.Equal(tc.data, decompressed) {
				t.Errorf("Данные не совпадают после сжатия/разжатия:\nОригинал: %v\nВосстановлено: %v",
					tc.data, decompressed)
			}

			// Проверяем, что сжатие действительно работает для больших данных
			if len(tc.data) > 100 && len(compressed) >= len(tc.data) {
				t.Logf("Предупреждение: сжатие не уменьшило размер данных (было: %d, стало: %d)",
					len(tc.data), len(compressed))
			}
		})
	}
}

func TestInvalidCompressedData(t *testing.T) {
	// Тестируем разжатие некорректных данных
	invalidData := []byte("invalid compressed data")

	_, err := decompressData(invalidData)
	if err == nil {
		t.Errorf("Ожидалась ошибка при разжатии некорректных данных")
	}
}

func TestListVault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Не удалось создать временный каталог: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.dat")
	password := "password123"

	// Создаем хранилище
	err = CreateVault(vaultPath, password)
	if err != nil {
		t.Fatalf("Не удалось создать хранилище: %v", err)
	}

	// Тестируем список пустого хранилища
	data, err := ListVault(vaultPath, password)
	if err != nil {
		t.Fatalf("Не удалось получить список пустого хранилища: %v", err)
	}

	if len(data.Entries) != 0 {
		t.Errorf("Пустое хранилище должно содержать 0 записей, получили %d", len(data.Entries))
	}

	if data.Comment != "Зашифрованное хранилище Flint Vault" {
		t.Errorf("Неверный комментарий: %s", data.Comment)
	}
}

// Benchmark тесты для проверки производительности
func BenchmarkCreateVault(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "vault_bench_*")
	if err != nil {
		b.Fatalf("Не удалось создать временный каталог: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	password := "BenchmarkPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vaultPath := filepath.Join(tmpDir, fmt.Sprintf("bench_%d.dat", i))
		err := CreateVault(vaultPath, password)
		if err != nil {
			b.Fatalf("Не удалось создать хранилище: %v", err)
		}
	}
}

func BenchmarkCompressDecompress(b *testing.B) {
	data := bytes.Repeat([]byte("Test data for compression benchmark "), 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compressed, err := compressData(data)
		if err != nil {
			b.Fatalf("Ошибка сжатия: %v", err)
		}

		_, err = decompressData(compressed)
		if err != nil {
			b.Fatalf("Ошибка разжатия: %v", err)
		}
	}
}

// TestExtractMultiple тестирует извлечение нескольких файлов и каталогов
func TestExtractMultiple(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vaultPath := filepath.Join(tmpDir, "test.dat")
	password := "password123"

	// Create test vault with multiple files and directories
	testData := VaultData{
		Entries: []VaultEntry{
			{
				Path:    "dir1",
				Name:    "dir1",
				IsDir:   true,
				Size:    0,
				Mode:    0755,
				ModTime: time.Now(),
				Content: nil,
			},
			{
				Path:    "dir1/file1.txt",
				Name:    "file1.txt",
				IsDir:   false,
				Size:    8,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("file1!!!"),
			},
			{
				Path:    "dir1/file2.txt",
				Name:    "file2.txt",
				IsDir:   false,
				Size:    8,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("file2!!!"),
			},
			{
				Path:    "dir2",
				Name:    "dir2",
				IsDir:   true,
				Size:    0,
				Mode:    0755,
				ModTime: time.Now(),
				Content: nil,
			},
			{
				Path:    "dir2/file3.txt",
				Name:    "file3.txt",
				IsDir:   false,
				Size:    8,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("file3!!!"),
			},
			{
				Path:    "single_file.txt",
				Name:    "single_file.txt",
				IsDir:   false,
				Size:    11,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("single file"),
			},
			{
				Path:    "another_file.txt",
				Name:    "another_file.txt",
				IsDir:   false,
				Size:    12,
				Mode:    0644,
				ModTime: time.Now(),
				Content: []byte("another file"),
			},
		},
		CreatedAt: time.Now(),
		Comment:   "Test vault for multiple extraction",
	}

	err = saveVaultData(vaultPath, password, testData)
	if err != nil {
		t.Fatalf("Failed to save test data: %v", err)
	}

	t.Run("ExtractMultipleFiles", func(t *testing.T) {
		extractDir := filepath.Join(tmpDir, "extract_multiple_files")
		targetPaths := []string{"single_file.txt", "another_file.txt"}

		extractedPaths, notFoundPaths, err := ExtractMultiple(vaultPath, password, targetPaths, extractDir)
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
		file1 := filepath.Join(extractDir, "single_file.txt")
		content1, err := os.ReadFile(file1)
		if err != nil {
			t.Fatalf("Failed to read extracted file1: %v", err)
		}
		if string(content1) != "single file" {
			t.Errorf("File1 content mismatch: expected 'single file', got '%s'", string(content1))
		}

		file2 := filepath.Join(extractDir, "another_file.txt")
		content2, err := os.ReadFile(file2)
		if err != nil {
			t.Fatalf("Failed to read extracted file2: %v", err)
		}
		if string(content2) != "another file" {
			t.Errorf("File2 content mismatch: expected 'another file', got '%s'", string(content2))
		}
	})

	t.Run("ExtractMultipleDirectories", func(t *testing.T) {
		extractDir := filepath.Join(tmpDir, "extract_multiple_dirs")
		targetPaths := []string{"dir1", "dir2"}

		extractedPaths, notFoundPaths, err := ExtractMultiple(vaultPath, password, targetPaths, extractDir)
		if err != nil {
			t.Fatalf("Failed to extract multiple directories: %v", err)
		}

		// Check results
		if len(extractedPaths) != 2 {
			t.Errorf("Expected 2 extracted paths, got %d", len(extractedPaths))
		}
		if len(notFoundPaths) != 0 {
			t.Errorf("Expected 0 not found paths, got %d", len(notFoundPaths))
		}

		// Check extracted directories and files
		dir1File1 := filepath.Join(extractDir, "file1.txt")
		content1, err := os.ReadFile(dir1File1)
		if err != nil {
			t.Fatalf("Failed to read dir1/file1.txt: %v", err)
		}
		if string(content1) != "file1!!!" {
			t.Errorf("Dir1/file1 content mismatch: expected 'file1!!!', got '%s'", string(content1))
		}

		dir2File3 := filepath.Join(extractDir, "file3.txt")
		content3, err := os.ReadFile(dir2File3)
		if err != nil {
			t.Fatalf("Failed to read dir2/file3.txt: %v", err)
		}
		if string(content3) != "file3!!!" {
			t.Errorf("Dir2/file3 content mismatch: expected 'file3!!!', got '%s'", string(content3))
		}
	})

	t.Run("ExtractMixed", func(t *testing.T) {
		extractDir := filepath.Join(tmpDir, "extract_mixed")
		targetPaths := []string{"dir1", "single_file.txt", "another_file.txt"}

		extractedPaths, notFoundPaths, err := ExtractMultiple(vaultPath, password, targetPaths, extractDir)
		if err != nil {
			t.Fatalf("Failed to extract mixed targets: %v", err)
		}

		// Check results
		if len(extractedPaths) != 3 {
			t.Errorf("Expected 3 extracted paths, got %d", len(extractedPaths))
		}
		if len(notFoundPaths) != 0 {
			t.Errorf("Expected 0 not found paths, got %d", len(notFoundPaths))
		}

		// Check directory content
		dir1File1 := filepath.Join(extractDir, "file1.txt")
		if _, err := os.Stat(dir1File1); os.IsNotExist(err) {
			t.Errorf("Dir1/file1.txt was not extracted")
		}

		// Check individual files
		singleFile := filepath.Join(extractDir, "single_file.txt")
		if _, err := os.Stat(singleFile); os.IsNotExist(err) {
			t.Errorf("single_file.txt was not extracted")
		}

		anotherFile := filepath.Join(extractDir, "another_file.txt")
		if _, err := os.Stat(anotherFile); os.IsNotExist(err) {
			t.Errorf("another_file.txt was not extracted")
		}
	})

	t.Run("ExtractWithNotFound", func(t *testing.T) {
		extractDir := filepath.Join(tmpDir, "extract_with_not_found")
		targetPaths := []string{"single_file.txt", "nonexistent.txt", "dir1", "nonexistent_dir"}

		extractedPaths, notFoundPaths, err := ExtractMultiple(vaultPath, password, targetPaths, extractDir)
		if err != nil {
			t.Fatalf("Failed to extract with not found items: %v", err)
		}

		// Check results
		if len(extractedPaths) != 2 {
			t.Errorf("Expected 2 extracted paths, got %d", len(extractedPaths))
		}
		if len(notFoundPaths) != 2 {
			t.Errorf("Expected 2 not found paths, got %d", len(notFoundPaths))
		}

		// Check that existing items were extracted
		singleFile := filepath.Join(extractDir, "single_file.txt")
		if _, err := os.Stat(singleFile); os.IsNotExist(err) {
			t.Errorf("single_file.txt was not extracted")
		}

		dir1File1 := filepath.Join(extractDir, "file1.txt")
		if _, err := os.Stat(dir1File1); os.IsNotExist(err) {
			t.Errorf("dir1/file1.txt was not extracted")
		}
	})

	t.Run("ExtractEmptyTargets", func(t *testing.T) {
		extractDir := filepath.Join(tmpDir, "extract_empty")
		targetPaths := []string{}

		_, _, err := ExtractMultiple(vaultPath, password, targetPaths, extractDir)
		if err == nil {
			t.Errorf("Expected error for empty target paths")
		}
		if !strings.Contains(err.Error(), "no target paths specified") {
			t.Errorf("Expected 'no target paths specified' error, got: %v", err)
		}
	})

	t.Run("ExtractAllNotFound", func(t *testing.T) {
		extractDir := filepath.Join(tmpDir, "extract_all_not_found")
		targetPaths := []string{"nonexistent1.txt", "nonexistent2.txt"}

		extractedPaths, notFoundPaths, err := ExtractMultiple(vaultPath, password, targetPaths, extractDir)
		if err == nil {
			t.Errorf("Expected error when all targets not found")
		}
		if !strings.Contains(err.Error(), "none of the specified files or directories were found") {
			t.Errorf("Expected 'none found' error, got: %v", err)
		}

		if len(extractedPaths) != 0 {
			t.Errorf("Expected 0 extracted paths, got %d", len(extractedPaths))
		}
		if len(notFoundPaths) != 2 {
			t.Errorf("Expected 2 not found paths, got %d", len(notFoundPaths))
		}
	})
}
