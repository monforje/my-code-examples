package service_test

import (
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

func TestUsersService_UpdateSettings_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfile(identityID)

	name := "New Name"
	bio := "new bio"
	input := &service.UpdateSettingsInput{
		DisplayName: &name,
		Bio:         &bio,
	}

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)
	deps.userProfiles.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, p *records.UserProfile) error {
		if p.DisplayName != "New Name" {
			t.Fatalf("DisplayName = %q, want New Name", p.DisplayName)
		}
		if p.BIO != "new bio" {
			t.Fatalf("BIO = %q, want new bio", p.BIO)
		}
		return nil
	})

	out, err := svc.UpdateSettings(ctx, input)
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if out.DisplayName == nil || *out.DisplayName != "New Name" {
		t.Fatalf("UpdateSettings().DisplayName = %v, want New Name", out.DisplayName)
	}
}

func TestUsersService_UpdateSettings_PartialUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfile(identityID)

	name := "Updated Name"
	input := &service.UpdateSettingsInput{
		DisplayName: &name,
		Bio:         nil,
	}

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)
	deps.userProfiles.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, p *records.UserProfile) error {
		if p.DisplayName != "Updated Name" {
			t.Fatalf("DisplayName = %q, want Updated Name", p.DisplayName)
		}
		if p.BIO != "hello world" {
			t.Fatalf("BIO = %q, want hello world (unchanged)", p.BIO)
		}
		return nil
	})

	out, err := svc.UpdateSettings(ctx, input)
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if out.DisplayName == nil || *out.DisplayName != "Updated Name" {
		t.Fatalf("UpdateSettings().DisplayName = %v, want Updated Name", out.DisplayName)
	}
}

func TestUsersService_UpdateSettings_ContextMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	ctx := context.Background()
	input := &service.UpdateSettingsInput{}

	_, err := svc.UpdateSettings(ctx, input)
	if !errors.Is(err, authctx.ErrAuthContextMissing) {
		t.Fatalf("UpdateSettings() error = %v, want ErrAuthContextMissing", err)
	}
}

func TestUsersService_UpdateSettings_ProfileNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	input := &service.UpdateSettingsInput{}

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(nil, postgresrepo.ErrUserProfileNotFound)

	_, err := svc.UpdateSettings(ctx, input)
	if err == nil {
		t.Fatal("UpdateSettings() error = nil, want error")
	}
}

func TestUsersService_UpdateSettings_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfile(identityID)
	updateErr := errors.New("update failed")

	name := "New Name"
	input := &service.UpdateSettingsInput{
		DisplayName: &name,
	}

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)
	deps.userProfiles.EXPECT().Update(ctx, gomock.Any()).Return(updateErr)

	_, err := svc.UpdateSettings(ctx, input)
	if err == nil {
		t.Fatal("UpdateSettings() error = nil, want error")
	}
}
