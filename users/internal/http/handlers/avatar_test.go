package handlers_test

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	httpserver "users/internal/http/gen"
	service "users/internal/services"
	apperrors "users/pkg/errors"
)

const (
	avatarUpdateOp = "users.profile.me.avatar.update"
	avatarDeleteOp = "users.profile.me.avatar.delete"
)

// ──────────────────────────────────────────────
// helpers
// ──────────────────────────────────────────────

// createTempImage создаёт временный файл указанного размера с нужным расширением.
func createTempImage(t *testing.T, size int, ext string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "avatar"+ext)
	data := make([]byte, size)
	// Минимальный заголовок для валидного формата.
	switch ext {
	case ".jpeg", ".jpg":
		data[0], data[1], data[2] = 0xFF, 0xD8, 0xFF // JPEG SOI
	case ".png":
		data[0], data[1], data[2], data[3] = 0x89, 0x50, 0x4E, 0x47 // PNG signature
	case ".webp":
		copy(data, "RIFF") // RIFF header для WebP
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

// authMultipartPUT отправляет multipart/form-data PUT с Bearer token и реальным содержимым файла.
func authMultipartPUT(t *testing.T, ts *httptest.Server, path, fieldName, filePath string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if filePath != "" {
		srcFile, err := os.Open(filePath)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		defer srcFile.Close()

		part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		if _, err := io.Copy(part, srcFile); err != nil {
			t.Fatalf("io.Copy: %v", err)
		}
	} else {
		_ = writer.WriteField(fieldName, "value")
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPut, ts.URL+path, &buf)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-access-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}

// authDelete отправляет DELETE запрос с Bearer token без тела.
func authDelete(t *testing.T, ts *httptest.Server, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, ts.URL+path, nil)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Authorization", "Bearer test-access-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}

// ──────────────────────────────────────────────
// UsersProfileMeAvatarUpdate
// ──────────────────────────────────────────────

func TestAvatarUpdate_Success_JPEG(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarUpdateOp)

	filePath := createTempImage(t, 100, ".jpeg")
	svc.EXPECT().UpdateAvatar(gomock.Any(), gomock.Any()).
		Return(&service.UpdateAvatarOutput{
			AvatarURL: "/uploads/avatars/new.jpeg",
			UpdatedAt: time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC),
		}, nil)

	resp := authMultipartPUT(t, ts, "/profile/me/avatar", "avatar", filePath)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.UpdateAvatarResponse
	decodeJSON(t, resp, &body)
	if body.AvatarUrl != "/uploads/avatars/new.jpeg" {
		t.Errorf("avatar_url = %q, want /uploads/avatars/new.jpeg", body.AvatarUrl)
	}
}

func TestAvatarUpdate_Success_PNG(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarUpdateOp)

	filePath := createTempImage(t, 100, ".png")
	svc.EXPECT().UpdateAvatar(gomock.Any(), gomock.Any()).
		Return(&service.UpdateAvatarOutput{
			AvatarURL: "/uploads/avatars/new.png",
			UpdatedAt: time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC),
		}, nil)

	resp := authMultipartPUT(t, ts, "/profile/me/avatar", "avatar", filePath)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.UpdateAvatarResponse
	decodeJSON(t, resp, &body)
	if body.AvatarUrl != "/uploads/avatars/new.png" {
		t.Errorf("avatar_url = %q, want /uploads/avatars/new.png", body.AvatarUrl)
	}
}

func TestAvatarUpdate_Success_WebP(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarUpdateOp)

	filePath := createTempImage(t, 100, ".webp")
	svc.EXPECT().UpdateAvatar(gomock.Any(), gomock.Any()).
		Return(&service.UpdateAvatarOutput{
			AvatarURL: "/uploads/avatars/new.webp",
			UpdatedAt: time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC),
		}, nil)

	resp := authMultipartPUT(t, ts, "/profile/me/avatar", "avatar", filePath)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.UpdateAvatarResponse
	decodeJSON(t, resp, &body)
	if body.AvatarUrl != "/uploads/avatars/new.webp" {
		t.Errorf("avatar_url = %q, want /uploads/avatars/new.webp", body.AvatarUrl)
	}
}

func TestAvatarUpdate_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarUpdateOp)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("avatar", "avatar.jpeg")
	part.Write([]byte("data"))
	writer.Close()

	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/profile/me/avatar", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestAvatarUpdate_NoFile(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarUpdateOp)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("avatar", "value")
	writer.Close()

	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/profile/me/avatar", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-access-token")
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAvatarUpdate_TooLarge(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarUpdateOp)

	filePath := createTempImage(t, 6*1024*1024, ".jpeg")

	resp := authMultipartPUT(t, ts, "/profile/me/avatar", "avatar", filePath)
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusRequestEntityTooLarge)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.AVATARTOOLARGE {
		t.Errorf("code = %q, want %q", body.Code, httpserver.AVATARTOOLARGE)
	}
}

func TestAvatarUpdate_InvalidFormat(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarUpdateOp)

	dir := t.TempDir()
	txtPath := filepath.Join(dir, "avatar.txt")
	os.WriteFile(txtPath, []byte("hello"), 0644)

	resp := authMultipartPUT(t, ts, "/profile/me/avatar", "avatar", txtPath)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INVALIDAVATARFORMAT {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INVALIDAVATARFORMAT)
	}
}

func TestAvatarUpdate_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarUpdateOp)

	filePath := createTempImage(t, 100, ".jpeg")
	svc.EXPECT().UpdateAvatar(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("disk full"))

	resp := authMultipartPUT(t, ts, "/profile/me/avatar", "avatar", filePath)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INTERNALERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INTERNALERROR)
	}
}

// ──────────────────────────────────────────────
// UsersProfileMeAvatarDelete
// ──────────────────────────────────────────────

func TestAvatarDelete_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarDeleteOp)

	svc.EXPECT().DeleteAvatar(gomock.Any()).
		Return(&service.DeleteAvatarOutput{
			AvatarURL: nil,
			UpdatedAt: time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC),
		}, nil)

	resp := authDelete(t, ts, "/profile/me/avatar")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.DeleteAvatarResponse
	decodeJSON(t, resp, &body)
	if body.AvatarUrl != nil {
		t.Errorf("avatar_url = %v, want nil", body.AvatarUrl)
	}
}

func TestAvatarDelete_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarDeleteOp)

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/profile/me/avatar", nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestAvatarDelete_NotFound(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarDeleteOp)

	svc.EXPECT().DeleteAvatar(gomock.Any()).
		Return(nil, apperrors.New("UsersService.DeleteAvatar", service.ErrAvatarNotFound))

	resp := authDelete(t, ts, "/profile/me/avatar")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.NOTFOUND {
		t.Errorf("code = %q, want %q", body.Code, httpserver.NOTFOUND)
	}
}

func TestAvatarDelete_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, avatarDeleteOp)

	svc.EXPECT().DeleteAvatar(gomock.Any()).
		Return(nil, errors.New("db down"))

	resp := authDelete(t, ts, "/profile/me/avatar")
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INTERNALERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INTERNALERROR)
	}
}
