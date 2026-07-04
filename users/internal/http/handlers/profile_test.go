package handlers_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	httpserver "users/internal/http/gen"
	postgresrepo "users/internal/repository/postgres"
	service "users/internal/services"
	apperrors "users/pkg/errors"
)

const profileOp = "users.profile.me.get"

// ──────────────────────────────────────────────
// UsersProfileMeGet
// ──────────────────────────────────────────────

func TestProfileGet_Success(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, profileOp)

	dn := "Test User"
	bio := "hello world"
	svc.EXPECT().GetProfile(gomock.Any()).Return(&service.ProfileOutput{
		ID:          uuid.New().String(),
		IdentityID:  uuid.New().String(),
		Email:       "user@example.com",
		DisplayName: &dn,
		Bio:         bio,
		CreatedAt:   time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC),
	}, nil)

	resp := authGet(t, ts, "/profile/me")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body httpserver.ProfileResponse
	decodeJSON(t, resp, &body)
	if body.Email != "user@example.com" {
		t.Errorf("email = %q, want user@example.com", body.Email)
	}
	if body.DisplayName == nil || *body.DisplayName != "Test User" {
		t.Errorf("display_name = %v, want %q", body.DisplayName, "Test User")
	}
	if body.Bio == nil || *body.Bio != "hello world" {
		t.Errorf("bio = %v, want %q", body.Bio, "hello world")
	}
}

func TestProfileGet_MissingAuthToken(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, profileOp)

	resp := getNoAuth(t, ts, "/profile/me")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestProfileGet_ProfileNotFound(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, profileOp)

	svc.EXPECT().GetProfile(gomock.Any()).
		Return(nil, apperrors.New("UsersService.GetProfile", postgresrepo.ErrUserProfileNotFound))

	resp := authGet(t, ts, "/profile/me")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.NOTFOUND {
		t.Errorf("code = %q, want %q", body.Code, httpserver.NOTFOUND)
	}
}

func TestProfileGet_ServiceError(t *testing.T) {
	svc := newMock(t)
	ts := setupAuthServer(t, svc, profileOp)

	svc.EXPECT().GetProfile(gomock.Any()).
		Return(nil, errors.New("db down"))

	resp := authGet(t, ts, "/profile/me")
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	var body httpserver.ErrorResponse
	decodeJSON(t, resp, &body)
	if body.Code != httpserver.INTERNALERROR {
		t.Errorf("code = %q, want %q", body.Code, httpserver.INTERNALERROR)
	}
}
