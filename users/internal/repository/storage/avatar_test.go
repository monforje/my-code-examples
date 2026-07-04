package storage_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

	"users/internal/repository/storage"
)

func newTestStorage(t *testing.T) *storage.LocalAvatarStorage {
	t.Helper()
	dir := t.TempDir()
	return storage.NewLocalAvatarStorage(dir, "/uploads/avatars")
}

func TestSave_CreatesFile(t *testing.T) {
	s := newTestStorage(t)
	identityID := uuid.New()
	body := []byte("fake png data")

	key, url, err := s.Save(identityID, "avatar.png", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	wantKey := identityID.String() + "/avatar.png"
	if key != wantKey {
		t.Errorf("objectKey = %q, want %q", key, wantKey)
	}

	wantURL := "/uploads/avatars/" + wantKey
	if url != wantURL {
		t.Errorf("url = %q, want %q", url, wantURL)
	}
}

func TestSave_OverwritesExistingFile(t *testing.T) {
	s := newTestStorage(t)
	identityID := uuid.New()

	_, _, err := s.Save(identityID, "avatar.png", bytes.NewReader([]byte("old data")))
	if err != nil {
		t.Fatalf("Save() first error = %v", err)
	}

	_, _, err = s.Save(identityID, "avatar.png", bytes.NewReader([]byte("new data")))
	if err != nil {
		t.Fatalf("Save() second error = %v", err)
	}
}

func TestDelete_RemovesFile(t *testing.T) {
	s := newTestStorage(t)
	identityID := uuid.New()

	key, _, err := s.Save(identityID, "avatar.png", bytes.NewReader([]byte("data")))
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := s.Delete(key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Повторное удаление не должно упасть (файл уже не существует)
	if err := s.Delete(key); err != nil {
		t.Fatalf("Delete() second call error = %v", err)
	}
}

func TestDelete_RemovesEmptyDir(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewLocalAvatarStorage(dir, "/uploads/avatars")
	identityID := uuid.New()

	key, _, err := s.Save(identityID, "avatar.png", bytes.NewReader([]byte("data")))
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Директория пользователя существует до удаления
	userDir := filepath.Join(dir, filepath.Dir(key))
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		t.Fatal("user dir should exist before Delete()")
	}

	if err := s.Delete(key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// После удаления последнего файла директория пользователя тоже удаляется
	if _, err := os.Stat(userDir); !os.IsNotExist(err) {
		t.Errorf("user dir should be removed after Delete(), err = %v", err)
	}
}

func TestSave_DifferentIdentityID_CreatesSeparateDirs(t *testing.T) {
	s := newTestStorage(t)
	id1 := uuid.New()
	id2 := uuid.New()

	key1, _, err := s.Save(id1, "avatar.png", bytes.NewReader([]byte("user1")))
	if err != nil {
		t.Fatalf("Save() id1 error = %v", err)
	}

	key2, _, err := s.Save(id2, "avatar.png", bytes.NewReader([]byte("user2")))
	if err != nil {
		t.Fatalf("Save() id2 error = %v", err)
	}

	if key1 == key2 {
		t.Errorf("objectKeys should differ, both = %q", key1)
	}

	// Удаление одного не влияет на другой
	if err := s.Delete(key1); err != nil {
		t.Fatalf("Delete(key1) error = %v", err)
	}

	if err := s.Delete(key2); err != nil {
		t.Fatalf("Delete(key2) error = %v", err)
	}
}

func TestSave_NestedFilename(t *testing.T) {
	s := newTestStorage(t)
	identityID := uuid.New()

	key, url, err := s.Save(identityID, "sub/dir/avatar.png", bytes.NewReader([]byte("data")))
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if !strings.Contains(key, "sub/dir/avatar.png") {
		t.Errorf("objectKey should contain nested path, got = %q", key)
	}
	if !strings.HasSuffix(url, "sub/dir/avatar.png") {
		t.Errorf("url should end with nested path, got = %q", url)
	}

	_ = s.Delete(key)
}
