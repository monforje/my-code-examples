package e2e_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"testing"

	e2ehelpers "users/tests/e2e/helpers"

	"github.com/google/uuid"
)

func generateTestJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return buf.Bytes()
}

func generateTestPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func TestAvatarUpdate_Success_JPEG(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	createTestProfileWithEmail(t, identityID, "avatar-jpeg@example.com")

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	content := generateTestJPEG(t, 100, 100)
	resp := e2ehelpers.UploadAvatar(t, client, token, "/profile/me/avatar", "avatar", "avatar.jpeg", content)
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["avatar_url"] == nil || body["avatar_url"] == "" {
		t.Fatalf("avatar_url is empty")
	}
}

func TestAvatarUpdate_Success_PNG(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	createTestProfileWithEmail(t, identityID, "avatar-png@example.com")

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	content := generateTestPNG(t, 100, 100)
	resp := e2ehelpers.UploadAvatar(t, client, token, "/profile/me/avatar", "avatar", "avatar.png", content)
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["avatar_url"] == nil || body["avatar_url"] == "" {
		t.Fatalf("avatar_url is empty")
	}
}

func TestAvatarUpdate_ReplacesOldAvatar(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	createTestProfileWithEmail(t, identityID, "avatar-replace@example.com")

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)

	content := generateTestJPEG(t, 100, 100)
	resp := e2ehelpers.UploadAvatar(t, client, token, "/profile/me/avatar", "avatar", "first.jpeg", content)
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	first := e2ehelpers.Decode[map[string]any](t, resp)

	content2 := generateTestPNG(t, 200, 200)
	resp = e2ehelpers.UploadAvatar(t, client, token, "/profile/me/avatar", "avatar", "second.png", content2)
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)
	second := e2ehelpers.Decode[map[string]any](t, resp)

	if first["avatar_url"] == second["avatar_url"] {
		t.Fatalf("avatar_url should change after replace: %v", first["avatar_url"])
	}
}

func TestAvatarUpdate_TooLarge(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	createTestProfileWithEmail(t, identityID, "avatar-large@example.com")

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	content := make([]byte, 6*1024*1024)
	for i := range content {
		content[i] = 0xFF
	}
	resp := e2ehelpers.UploadAvatar(t, client, token, "/profile/me/avatar", "avatar", "large.jpeg", content)
	e2ehelpers.ExpectStatus(t, resp, http.StatusRequestEntityTooLarge)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["code"] != "AVATAR_TOO_LARGE" {
		t.Fatalf("code = %v, want AVATAR_TOO_LARGE", body["code"])
	}
}

func TestAvatarUpdate_InvalidFormat(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	createTestProfileWithEmail(t, identityID, "avatar-invalid@example.com")

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	resp := e2ehelpers.UploadAvatar(t, client, token, "/profile/me/avatar", "avatar", "avatar.txt", []byte("not an image"))
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnprocessableEntity)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["code"] != "INVALID_AVATAR_FORMAT" {
		t.Fatalf("code = %v, want INVALID_AVATAR_FORMAT", body["code"])
	}
}

func TestAvatarUpdate_NoFile(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	createTestProfileWithEmail(t, identityID, "avatar-nofile@example.com")

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	resp := e2ehelpers.UploadAvatar(t, client, token, "/profile/me/avatar", "avatar", "", nil)
	e2ehelpers.ExpectStatus(t, resp, http.StatusBadRequest)
}

func TestAvatarUpdate_ProfileNotFound(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	content := generateTestJPEG(t, 100, 100)
	resp := e2ehelpers.UploadAvatar(t, client, token, "/profile/me/avatar", "avatar", "avatar.jpeg", content)
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}

func TestAvatarDelete_Success(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	createTestProfileWithEmail(t, identityID, "avatar-del@example.com")

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	content := generateTestJPEG(t, 100, 100)
	resp := e2ehelpers.UploadAvatar(t, client, token, "/profile/me/avatar", "avatar", "avatar.jpeg", content)
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	resp = e2ehelpers.DeleteAuth(t, client, token, "/profile/me/avatar")
	e2ehelpers.ExpectStatus(t, resp, http.StatusOK)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["avatar_url"] != nil {
		t.Fatalf("avatar_url should be nil after delete: %v", body["avatar_url"])
	}
}

func TestAvatarDelete_NotFound(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New()
	createTestProfileWithEmail(t, identityID, "avatar-del-notfound@example.com")

	token := e2ehelpers.GenerateToken(t, tokenManager, identityID)
	resp := e2ehelpers.DeleteAuth(t, client, token, "/profile/me/avatar")
	e2ehelpers.ExpectStatus(t, resp, http.StatusNotFound)
}

func TestAvatarDelete_MissingToken(t *testing.T) {
	resetE2E(t)

	resp := e2ehelpers.DeleteAuth(t, client, "", "/profile/me/avatar")
	e2ehelpers.ExpectStatus(t, resp, http.StatusUnauthorized)

	body := e2ehelpers.Decode[map[string]any](t, resp)
	if body["code"] != "MISSING_AUTH_TOKEN" {
		t.Fatalf("code = %v, want MISSING_AUTH_TOKEN", body["code"])
	}
}
