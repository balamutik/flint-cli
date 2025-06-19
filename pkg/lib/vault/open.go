package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

// OpenVault открывает и расшифровывает существующее хранилище
func OpenVault(path string, password string) ([]byte, error) {
	// Валидация входных параметров
	if len(password) == 0 {
		return nil, fmt.Errorf("пароль не может быть пустым")
	}

	if path == "" {
		return nil, fmt.Errorf("путь к файлу не может быть пустым")
	}

	// Открываем файл хранилища
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	// Читаем заголовок
	var header VaultHeader
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("ошибка чтения заголовка: %w", err)
	}

	// Проверяем магический заголовок
	if string(header.Magic[:]) != VaultMagic {
		return nil, fmt.Errorf("неверный формат файла хранилища")
	}

	// Проверяем версию
	if header.Version != 1 {
		return nil, fmt.Errorf("неподдерживаемая версия хранилища: %d", header.Version)
	}

	// Выводим ключ из пароля
	key := pbkdf2.Key([]byte(password), header.Salt[:], int(header.Iterations), KeyLength, sha256.New)

	// Очищаем пароль из памяти
	passwordBytes := []byte(password)
	for i := range passwordBytes {
		passwordBytes[i] = 0
	}

	// Создаем AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания AES cipher: %w", err)
	}

	// Создаем GCM для расшифровки
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM: %w", err)
	}

	// Читаем зашифрованные данные
	ciphertext, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения зашифрованных данных: %w", err)
	}

	// Расшифровываем данные
	plaintext, err := gcm.Open(nil, header.Nonce[:], ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки: неверный пароль или поврежденные данные")
	}

	// Очищаем ключ из памяти
	for i := range key {
		key[i] = 0
	}

	return plaintext, nil
}
