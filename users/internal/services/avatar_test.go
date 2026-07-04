package service_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"users/internal/authctx"
	"users/internal/models/records"
	postgresrepo "users/internal/repository/postgres"
	service "users/internal/services"
)

func TestUsersService_UpdateAvatar_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfileNoAvatar(identityID)
	file := bytes.NewReader([]byte("avatar content"))

	input := &service.UpdateAvatarInput{
		Filename: "avatar.png",
		File:     file,
		FileSize: 14,
	}

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)
	deps.avatar.EXPECT().Save(identityID, "avatar.png", file).Return("new-avatar.png", "/uploads/avatars/new-avatar.png", nil)
	deps.userProfiles.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, p *records.UserProfile) error {
		if p.AvatarURL != "/uploads/avatars/new-avatar.png" {
			t.Fatalf("AvatarURL = %q, want /uploads/avatars/new-avatar.png", p.AvatarURL)
		}
		if p.AvatarObjectKey != "new-avatar.png" {
			t.Fatalf("AvatarObjectKey = %q, want new-avatar.png", p.AvatarObjectKey)
		}
		return nil
	})

	out, err := svc.UpdateAvatar(ctx, input)
	if err != nil {
		t.Fatalf("UpdateAvatar() error = %v", err)
	}
	if out.AvatarURL != "/uploads/avatars/new-avatar.png" {
		t.Fatalf("UpdateAvatar().AvatarURL = %q, want /uploads/avatars/new-avatar.png", out.AvatarURL)
	}
}

func TestUsersService_UpdateAvatar_ReplaceExisting(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfile(identityID)
	file := bytes.NewReader([]byte("new avatar content"))

	input := &service.UpdateAvatarInput{
		Filename: "avatar2.png",
		File:     file,
		FileSize: 19,
	}

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)
	deps.avatar.EXPECT().Delete("avatars/123.png").Return(nil)
	deps.avatar.EXPECT().Save(identityID, "avatar2.png", file).Return("new-avatar2.png", "/uploads/avatars/new-avatar2.png", nil)
	deps.userProfiles.EXPECT().Update(ctx, gomock.Any()).Return(nil)

	out, err := svc.UpdateAvatar(ctx, input)
	if err != nil {
		t.Fatalf("UpdateAvatar() error = %v", err)
	}
	if out.AvatarURL != "/uploads/avatars/new-avatar2.png" {
		t.Fatalf("UpdateAvatar().AvatarURL = %q, want /uploads/avatars/new-avatar2.png", out.AvatarURL)
	}
}

func TestUsersService_UpdateAvatar_ContextMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	ctx := context.Background()
	input := &service.UpdateAvatarInput{}

	_, err := svc.UpdateAvatar(ctx, input)
	if !errors.Is(err, authctx.ErrAuthContextMissing) {
		t.Fatalf("UpdateAvatar() error = %v, want ErrAuthContextMissing", err)
	}
}

func TestUsersService_UpdateAvatar_ProfileNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	input := &service.UpdateAvatarInput{}

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(nil, postgresrepo.ErrUserProfileNotFound)

	_, err := svc.UpdateAvatar(ctx, input)
	if err == nil {
		t.Fatal("UpdateAvatar() error = nil, want error")
	}
}

func TestUsersService_UpdateAvatar_DeleteOldError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfile(identityID)
	file := bytes.NewReader([]byte("avatar"))
	deleteErr := errors.New("delete failed")

	input := &service.UpdateAvatarInput{
		Filename: "avatar.png",
		File:     file,
		FileSize: 7,
	}

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)
	deps.avatar.EXPECT().Delete("avatars/123.png").Return(deleteErr)

	_, err := svc.UpdateAvatar(ctx, input)
	if err == nil {
		t.Fatal("UpdateAvatar() error = nil, want error")
	}
}

func TestUsersService_UpdateAvatar_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfileNoAvatar(identityID)
	file := bytes.NewReader([]byte("avatar"))
	saveErr := errors.New("save failed")

	input := &service.UpdateAvatarInput{
		Filename: "avatar.png",
		File:     file,
		FileSize: 7,
	}

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)
	deps.avatar.EXPECT().Save(identityID, "avatar.png", file).Return("", "", saveErr)

	_, err := svc.UpdateAvatar(ctx, input)
	if err == nil {
		t.Fatal("UpdateAvatar() error = nil, want error")
	}
}

func TestUsersService_UpdateAvatar_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfileNoAvatar(identityID)
	file := bytes.NewReader([]byte("avatar"))
	updateErr := errors.New("update failed")

	input := &service.UpdateAvatarInput{
		Filename: "avatar.png",
		File:     file,
		FileSize: 7,
	}

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)
	deps.avatar.EXPECT().Save(identityID, "avatar.png", file).Return("new-avatar.png", "/uploads/avatars/new-avatar.png", nil)
	deps.userProfiles.EXPECT().Update(ctx, gomock.Any()).Return(updateErr)

	_, err := svc.UpdateAvatar(ctx, input)
	if err == nil {
		t.Fatal("UpdateAvatar() error = nil, want error")
	}
}

func TestUsersService_DeleteAvatar_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfile(identityID)

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)
	deps.avatar.EXPECT().Delete("avatars/123.png").Return(nil)
	deps.userProfiles.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, p *records.UserProfile) error {
		if p.AvatarURL != "" {
			t.Fatalf("AvatarURL = %q, want empty", p.AvatarURL)
		}
		if p.AvatarObjectKey != "" {
			t.Fatalf("AvatarObjectKey = %q, want empty", p.AvatarObjectKey)
		}
		return nil
	})

	out, err := svc.DeleteAvatar(ctx)
	if err != nil {
		t.Fatalf("DeleteAvatar() error = %v", err)
	}
	if out.AvatarURL != nil {
		t.Fatalf("DeleteAvatar().AvatarURL = %v, want nil", out.AvatarURL)
	}
}

func TestUsersService_DeleteAvatar_ContextMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	ctx := context.Background()

	_, err := svc.DeleteAvatar(ctx)
	if !errors.Is(err, authctx.ErrAuthContextMissing) {
		t.Fatalf("DeleteAvatar() error = %v, want ErrAuthContextMissing", err)
	}
}

func TestUsersService_DeleteAvatar_ProfileNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(nil, postgresrepo.ErrUserProfileNotFound)

	_, err := svc.DeleteAvatar(ctx)
	if err == nil {
		t.Fatal("DeleteAvatar() error = nil, want error")
	}
}

func TestUsersService_DeleteAvatar_AvatarNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfileNoAvatar(identityID)

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)

	_, err := svc.DeleteAvatar(ctx)
	if err == nil {
		t.Fatal("DeleteAvatar() error = nil, want error")
	}
	if !errors.Is(err, service.ErrAvatarNotFound) {
		t.Fatalf("DeleteAvatar() error = %v, want ErrAvatarNotFound", err)
	}
}

func TestUsersService_DeleteAvatar_DeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfile(identityID)
	deleteErr := errors.New("delete failed")

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)
	deps.avatar.EXPECT().Delete("avatars/123.png").Return(deleteErr)

	_, err := svc.DeleteAvatar(ctx)
	if err == nil {
		t.Fatal("DeleteAvatar() error = nil, want error")
	}
}
