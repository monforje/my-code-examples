package handlers_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	httpserver "users/internal/http/gen"
	postgresrepo "users/internal/repository/postgres"
	service "users/internal/services"
	apperrors "users/pkg/errors"
)

const settingsOp = "users.profile.me.settings.update"

// ──────────────────────────────────────────────
// UsersProfileMeSettingsUpdate
// ──────────────────────────────────────────────

func TestSettingsUpdate_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, settingsOp)

	dn := "New Name"
	bio := "New bio"
	svc.EXPECT().UpdateSettings(gomock.Any(), &service.UpdateSettingsInput{
		DisplayName: &dn,
		Bio:         &bio,
	}).Return(&service.ProfileOutput{
		ID:          uuid.New().String(),
		IdentityID:  uuid.New().String(),
		Email:       "user@example.com",
		DisplayName: &dn,
		Bio:         bio,
		CreatedAt:   time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC),
	}, nil)

	resp := authPatchJSON(t, ts, "/profile/me/settings", map[string]string{
		"display_name": "New Name",
		"bio":          "New bio",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.ProfileResponse
	decodeJSON(t, resp, &body)
	if body.DisplayName == nil || *body.DisplayName != "New Name" {
		t.Errorf("display_name = %v, want %q", body.DisplayName, "New Name")
	}
	if body.Bio == nil || *body.Bio != "New bio" {
		t.Errorf("bio = %v, want %q", body.Bio, "New bio")
	}
}

func TestSettingsUpdate_PartialUpdate(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, settingsOp)

	dn := "Only Name"
	svc.EXPECT().UpdateSettings(gomock.Any(), &service.UpdateSettingsInput{
		DisplayName: &dn,
	}).Return(&service.ProfileOutput{
		ID:          uuid.New().String(),
		IdentityID:  uuid.New().String(),
		Email:       "user@example.com",
		DisplayName: &dn,
		Bio:         "",
		CreatedAt:   time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC),
	}, nil)

	resp := authPatchJSON(t, ts, "/profile/me/settings", map[string]string{
		"display_name": "Only Name",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.ProfileResponse
	decodeJSON(t, resp, &body)
	if body.DisplayName == nil || *body.DisplayName != "Only Name" {
		t.Errorf("display_name = %v, want %q", body.DisplayName, "Only Name")
	}
}

func TestSettingsUpdate_InvalidJSON(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, settingsOp)

	resp := authPatchJSON(t, ts, "/profile/me/settings", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INVALIDJSON {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INVALIDJSON)
	}
}

func TestSettingsUpdate_EmptyDisplayName(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, settingsOp)

	resp := authPatchJSON(t, ts, "/profile/me/settings", map[string]string{
		"display_name": "",
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.VALIDATIONERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.VALIDATIONERROR)
	}
}

func TestSettingsUpdate_DisplayNameTooLong(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, settingsOp)

	longName := strings.Repeat("a", 51)
	resp := authPatchJSON(t, ts, "/profile/me/settings", map[string]string{
		"display_name": longName,
	})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.VALIDATIONERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.VALIDATIONERROR)
	}
}

func TestSettingsUpdate_ProfileNotFound(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, settingsOp)

	svc.EXPECT().UpdateSettings(gomock.Any(), gomock.Any()).
		Return(nil, apperrors.New("UsersService.UpdateSettings", postgresrepo.ErrUserProfileNotFound))

	resp := authPatchJSON(t, ts, "/profile/me/settings", map[string]string{
		"display_name": "Test",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.NOTFOUND {
		t.Errorf("code = %q, want %q", body.Code, httpserver.NOTFOUND)
	}
}

func TestSettingsUpdate_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, settingsOp)

	svc.EXPECT().UpdateSettings(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("db down"))

	resp := authPatchJSON(t, ts, "/profile/me/settings", map[string]string{
		"display_name": "Test",
	})
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INTERNALERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INTERNALERROR)
	}
}
