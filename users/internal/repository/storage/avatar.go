// Package storage реализует локальное файловое хранилище аватаров пользователей.
//
// Структура на диске:
//
//	<baseDir>/
//	  <identity_id>/
//	    <filename>          — файл аватара
//
//publicBase — префикс URL для формирования публичной ссылки на аватар.
//Например, если baseDir = "./uploads/avatars", а publicBase = "/uploads/avatars",
//то файл uploads/avatars/abc123/avatar.png будет доступен по URL /uploads/avatars/abc123/avatar.png.
package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// LocalAvatarStorage хранит аватары на локальной файловой системе.
type LocalAvatarStorage struct {
	baseDir    string // базовая директория на диске, куда пишутся файлы
	publicBase string // публичный префикс URL для отдачи файлов через HTTP
}

// NewLocalAvatarStorage создаёт хранилище аватаров.
//   - baseDir — путь к директории на диске (например "./uploads/avatars")
//   - publicBase — URL-префикс (например "/uploads/avatars")
func NewLocalAvatarStorage(baseDir, publicBase string) *LocalAvatarStorage {
	return &LocalAvatarStorage{
		baseDir:    baseDir,
		publicBase: publicBase,
	}
}

// Save сохраняет файл аватара на диск.
//
// Файл записывается по пути <baseDir>/<identityID>/<filename>.
// Возвращает objectKey (относительный путь от baseDir) и url (публичный URL).
//
// Перед записью создаётся директория пользователя, если она не существует.
// Если файл с таким именем уже существует — он перезаписывается.
func (s *LocalAvatarStorage) Save(identityID uuid.UUID, filename string, r io.Reader) (objectKey string, url string, err error) {
	// Формируем относительный путь (objectKey) и полный путь на диске
	key := filepath.Join(identityID.String(), filename)
	path := filepath.Join(s.baseDir, key)

	// Создаём все промежуточные директории (включая директорию пользователя)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", "", fmt.Errorf("create avatar dir: %w", err)
	}

	// Создаём файл и записываем содержимое из переданного io.Reader
	f, err := os.Create(path)
	if err != nil {
		return "", "", fmt.Errorf("create avatar file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", "", fmt.Errorf("write avatar file: %w", err)
	}

	// Возвращаем objectKey и публичный URL: <publicBase>/<identity_id>/<filename>
	return key, s.publicBase + "/" + key, nil
}

// Delete удаляет файл аватара по objectKey.
//
// Если файл не существует — ошибка игнорируется.
// После удаления файла, если директория пользователя пуста — она тоже удаляется.
func (s *LocalAvatarStorage) Delete(objectKey string) error {
	path := filepath.Join(s.baseDir, objectKey)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete avatar file: %w", err)
	}

	// Пытаемся удалить пустую директорию пользователя
	dir := filepath.Dir(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	if len(entries) == 0 {
		_ = os.Remove(dir)
	}

	return nil
}
