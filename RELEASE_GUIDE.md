# 🚀 Flint Vault Release Guide

Пошаговая инструкция для создания релизов проекта Flint Vault CLI.

## 📋 Подготовка к релизу

### 1. Убедитесь что всё готово
```bash
# Проверьте что все тесты проходят
go test ./...

# Протестируйте локальную сборку
./test-build.sh

# Убедитесь что все изменения зафиксированы
git status
git add .
git commit -m "Prepare for release v1.0.0"
git push origin main
```

### 2. Протестируйте кросс-компиляцию
```bash
# Тестовая сборка
./build.sh v1.0.0-test

# Проверьте что все платформы собрались
ls -la dist/

# Протестируйте один из бинарных файлов
cd build/linux-amd64 && ./flint-vault version
```

## 🏷️ Создание релиза

### Метод 1: Автоматический релиз (рекомендуется)

1. **Создайте и запушьте тег:**
```bash
# Создайте тег версии
git tag v1.0.0

# Отправьте тег на GitHub
git push origin v1.0.0
```

2. **GitHub Actions автоматически:**
   - Запустит тесты
   - Соберёт бинарные файлы для всех платформ
   - Создаст релиз на GitHub
   - Загрузит все файлы

3. **Проверьте релиз:**
   - Перейдите в раздел "Releases" вашего репозитория
   - Убедитесь что релиз создан корректно
   - Проверьте наличие всех файлов

### Метод 2: Ручной релиз

1. **Соберите релиз локально:**
```bash
# Очистите предыдущие сборки
rm -rf build/ dist/

# Соберите для всех платформ
./build.sh v1.0.0

# Проверьте результат
ls -la dist/
```

2. **Создайте релиз на GitHub:**
   - Перейдите в раздел "Releases" → "Create a new release"
   - Укажите тег: `v1.0.0`
   - Название: `Flint Vault v1.0.0`
   - Используйте шаблон из `RELEASE_TEMPLATE.md`

3. **Загрузите файлы:**
   - Перетащите все файлы из `dist/` в раздел "Attach binaries"
   - Обязательно включите `checksums.txt`

## 📦 Проверка релиза

### Тестирование скачанных файлов
```bash
# Скачайте архив для вашей платформы
wget https://github.com/yourusername/flint-vault-cli/releases/download/v1.0.0/flint-vault-v1.0.0-linux-amd64.tar.gz

# Проверьте контрольную сумму
wget https://github.com/yourusername/flint-vault-cli/releases/download/v1.0.0/checksums.txt
sha256sum -c checksums.txt

# Извлеките и протестируйте
tar -xzf flint-vault-v1.0.0-linux-amd64.tar.gz
cd flint-vault-v1.0.0-linux-amd64
./flint-vault version
./flint-vault --help
```

### Тестирование установки
```bash
# Протестируйте скрипт установки
sudo ./install.sh

# Проверьте что команда доступна глобально
which flint-vault
flint-vault version
```

## 🔧 Обновление документации

### После успешного релиза:

1. **Обновите README.md:**
```markdown
## Installation

### Download Binary (Recommended)
Download the latest release for your platform:
[Releases](https://github.com/yourusername/flint-vault-cli/releases)

### Quick Install (Linux/macOS)
```bash
# Replace with your actual GitHub username
wget https://github.com/yourusername/flint-vault-cli/releases/download/v1.0.0/flint-vault-v1.0.0-linux-amd64.tar.gz
tar -xzf flint-vault-v1.0.0-linux-amd64.tar.gz
cd flint-vault-v1.0.0-linux-amd64
sudo ./install.sh
```

2. **Создайте CHANGELOG.md** (если нужно):
```markdown
# Changelog

## [v1.0.0] - 2025-06-21
### Added
- Initial release
- AES-256-GCM encryption with PBKDF2 key derivation
- High-performance streaming operations
- Parallel processing support
- Cross-platform support (Linux, macOS, Windows)

### Security
- Military-grade encryption
- SHA-256 integrity verification
- Memory-safe operations
```

## 🎯 Структура версий

### Семантическое версионирование (SemVer)
- **Major (1.x.x)**: Существенные изменения, несовместимые с предыдущими версиями
- **Minor (x.1.x)**: Новые функции, обратно совместимые
- **Patch (x.x.1)**: Исправления ошибок, обратно совместимые

### Примеры:
- `v1.0.0` - Первый стабильный релиз
- `v1.1.0` - Добавлены новые команды
- `v1.0.1` - Исправлены ошибки
- `v2.0.0` - Изменён формат vault файлов

## 🛠️ Устранение неполадок

### GitHub Actions не запускается
- Проверьте что тег создан корректно: `git tag -l`
- Убедитесь что workflows файл существует: `.github/workflows/release.yml`
- Проверьте логи в разделе "Actions" на GitHub

### Сборка падает
- Запустите локально: `./build.sh v1.0.0`
- Проверьте что все тесты проходят: `go test ./...`
- Убедитесь что Go модули корректны: `go mod tidy && go mod verify`

### Файлы не загружаются
- Проверьте права доступа к репозиторию
- Убедитесь что `GITHUB_TOKEN` имеет права на создание релизов
- Попробуйте создать релиз вручную

## 📋 Чек-лист релиза

- [ ] Все тесты проходят
- [ ] Локальная сборка работает
- [ ] Кросс-компиляция успешна
- [ ] Тег создан и отправлен
- [ ] GitHub Actions завершился успешно
- [ ] Релиз создан на GitHub
- [ ] Все файлы присутствуют
- [ ] Контрольные суммы корректны
- [ ] Релиз протестирован на разных платформах
- [ ] Документация обновлена
- [ ] Объявление о релизе (если нужно)

---

**Готово к релизу! 🎉** 