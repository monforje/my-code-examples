package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"users/internal/authctx"
	"users/internal/models/records"
	postgresrepo "users/internal/repository/postgres"
	service "users/internal/services"
)

func TestUsersService_GetProfile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	profile := testProfile(identityID)

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(profile, nil)

	out, err := svc.GetProfile(ctx)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if out.Email != "test@example.com" {
		t.Fatalf("GetProfile().Email = %q, want test@example.com", out.Email)
	}
	if out.DisplayName == nil || *out.DisplayName != "Test User" {
		t.Fatalf("GetProfile().DisplayName = %v, want Test User", out.DisplayName)
	}
	if out.Bio != "hello world" {
		t.Fatalf("GetProfile().Bio = %q, want hello world", out.Bio)
	}
}

func TestUsersService_GetProfile_ContextMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	ctx := context.Background()

	_, err := svc.GetProfile(ctx)
	if !errors.Is(err, authctx.ErrAuthContextMissing) {
		t.Fatalf("GetProfile() error = %v, want ErrAuthContextMissing", err)
	}
}

func TestUsersService_GetProfile_ProfileNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(nil, postgresrepo.ErrUserProfileNotFound)

	_, err := svc.GetProfile(ctx)
	if err == nil {
		t.Fatal("GetProfile() error = nil, want error")
	}
}

func TestUsersService_GetProfile_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	ctx := authCtx(identityID)
	repoErr := errors.New("db connection failed")

	deps.userProfiles.EXPECT().GetByIdentityID(ctx, identityID).Return(nil, repoErr)

	_, err := svc.GetProfile(ctx)
	if err == nil {
		t.Fatal("GetProfile() error = nil, want error")
	}
}

// ─── HandleIdentityCreated ──────────────────────────────────────────────────

func TestUsersService_HandleIdentityCreated_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(nil, postgresrepo.ErrUserProfileNotFound)
	deps.userProfiles.EXPECT().ExistsByDisplayName(gomock.Any(), gomock.Any()).Return(false, nil)
	deps.avatar.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Return("key", "url", nil)
	var createdProfileID uuid.UUID
	var createdDisplayName string
	deps.userProfiles.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, profile *records.UserProfile) error {
		createdProfileID = profile.ID
		createdDisplayName = profile.DisplayName
		return nil
	})

	err := svc.HandleIdentityCreated(context.Background(), &service.HandleIdentityCreatedInput{
		IdentityID: identityID.String(),
		Email:      "new@example.com",
	})
	if err != nil {
		t.Fatalf("HandleIdentityCreated() error = %v", err)
	}

	select {
	case call := <-deps.gitAuth.calls:
		if call.ProfileID != createdProfileID {
			t.Fatalf("git auth profile id = %v, want %v", call.ProfileID, createdProfileID)
		}
		if call.Username != createdDisplayName {
			t.Fatalf("git auth username = %q, want %q", call.Username, createdDisplayName)
		}
		if call.Email != "new@example.com" {
			t.Fatalf("git auth email = %q, want new@example.com", call.Email)
		}
	case <-time.After(time.Second):
		t.Fatal("git auth RegisterGitUser was not called")
	}
}

func TestUsersService_HandleIdentityCreated_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	existing := testProfile(identityID)

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(existing, nil)

	err := svc.HandleIdentityCreated(context.Background(), &service.HandleIdentityCreatedInput{
		IdentityID: identityID.String(),
		Email:      "new@example.com",
	})
	if err != nil {
		t.Fatalf("HandleIdentityCreated() error = %v, want nil (idempotent)", err)
	}
}

func TestUsersService_HandleIdentityCreated_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	err := svc.HandleIdentityCreated(context.Background(), &service.HandleIdentityCreatedInput{
		IdentityID: "not-a-uuid",
		Email:      "test@example.com",
	})
	if err == nil {
		t.Fatal("HandleIdentityCreated() error = nil, want error")
	}
}

func TestUsersService_HandleIdentityCreated_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	repoErr := errors.New("db connection failed")

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(nil, repoErr)

	err := svc.HandleIdentityCreated(context.Background(), &service.HandleIdentityCreatedInput{
		IdentityID: identityID.String(),
		Email:      "test@example.com",
	})
	if err == nil {
		t.Fatal("HandleIdentityCreated() error = nil, want error")
	}
}

func TestUsersService_HandleIdentityCreated_ExistsByDisplayNameError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	repoErr := errors.New("db connection failed")

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(nil, postgresrepo.ErrUserProfileNotFound)
	deps.userProfiles.EXPECT().ExistsByDisplayName(gomock.Any(), gomock.Any()).Return(false, repoErr)

	err := svc.HandleIdentityCreated(context.Background(), &service.HandleIdentityCreatedInput{
		IdentityID: identityID.String(),
		Email:      "test@example.com",
	})
	if err == nil {
		t.Fatal("HandleIdentityCreated() error = nil, want error")
	}
}

func TestUsersService_HandleIdentityCreated_AvatarSaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	saveErr := errors.New("disk full")

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(nil, postgresrepo.ErrUserProfileNotFound)
	deps.userProfiles.EXPECT().ExistsByDisplayName(gomock.Any(), gomock.Any()).Return(false, nil)
	deps.avatar.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Return("", "", saveErr)

	err := svc.HandleIdentityCreated(context.Background(), &service.HandleIdentityCreatedInput{
		IdentityID: identityID.String(),
		Email:      "test@example.com",
	})
	if err == nil {
		t.Fatal("HandleIdentityCreated() error = nil, want error")
	}
}

// ─── HandleIdentityUpdated ──────────────────────────────────────────────────

func TestUsersService_HandleIdentityUpdated_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	existing := testProfile(identityID)

	newEmail := "updated@example.com"
	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(existing, nil)
	deps.userProfiles.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	err := svc.HandleIdentityUpdated(context.Background(), &service.HandleIdentityUpdatedInput{
		IdentityID: identityID.String(),
		Email:      &newEmail,
	})
	if err != nil {
		t.Fatalf("HandleIdentityUpdated() error = %v", err)
	}
}

func TestUsersService_HandleIdentityUpdated_ProfileNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(nil, postgresrepo.ErrUserProfileNotFound)

	err := svc.HandleIdentityUpdated(context.Background(), &service.HandleIdentityUpdatedInput{
		IdentityID: identityID.String(),
		Email:      strPtr("new@example.com"),
	})
	if err != nil {
		t.Fatalf("HandleIdentityUpdated() error = %v, want nil (profile not found)", err)
	}
}

func TestUsersService_HandleIdentityUpdated_NoChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	existing := testProfile(identityID)

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(existing, nil)

	err := svc.HandleIdentityUpdated(context.Background(), &service.HandleIdentityUpdatedInput{
		IdentityID: identityID.String(),
	})
	if err != nil {
		t.Fatalf("HandleIdentityUpdated() error = %v, want nil (no changes)", err)
	}
}

func TestUsersService_HandleIdentityUpdated_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	existing := testProfile(identityID)
	repoErr := errors.New("db connection failed")

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(existing, nil)
	deps.userProfiles.EXPECT().Update(gomock.Any(), gomock.Any()).Return(repoErr)

	newEmail := "updated@example.com"
	err := svc.HandleIdentityUpdated(context.Background(), &service.HandleIdentityUpdatedInput{
		IdentityID: identityID.String(),
		Email:      &newEmail,
	})
	if err == nil {
		t.Fatal("HandleIdentityUpdated() error = nil, want error")
	}
}

// ─── HandleIdentityDeleted ──────────────────────────────────────────────────

func TestUsersService_HandleIdentityDeleted_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	existing := testProfileNoAvatar(identityID)

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(existing, nil)
	deps.userProfiles.EXPECT().Delete(gomock.Any(), existing.ID).Return(nil)

	err := svc.HandleIdentityDeleted(context.Background(), &service.HandleIdentityDeletedInput{
		IdentityID: identityID.String(),
	})
	if err != nil {
		t.Fatalf("HandleIdentityDeleted() error = %v", err)
	}
}

func TestUsersService_HandleIdentityDeleted_WithAvatar(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	existing := testProfile(identityID)

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(existing, nil)
	deps.avatar.EXPECT().Delete(existing.AvatarObjectKey).Return(nil)
	deps.userProfiles.EXPECT().Delete(gomock.Any(), existing.ID).Return(nil)

	err := svc.HandleIdentityDeleted(context.Background(), &service.HandleIdentityDeletedInput{
		IdentityID: identityID.String(),
	})
	if err != nil {
		t.Fatalf("HandleIdentityDeleted() error = %v", err)
	}
}

func TestUsersService_HandleIdentityDeleted_ProfileNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(nil, postgresrepo.ErrUserProfileNotFound)

	err := svc.HandleIdentityDeleted(context.Background(), &service.HandleIdentityDeletedInput{
		IdentityID: identityID.String(),
	})
	if err != nil {
		t.Fatalf("HandleIdentityDeleted() error = %v, want nil (idempotent)", err)
	}
}

func TestUsersService_HandleIdentityDeleted_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps)

	identityID := uuid.New()
	existing := testProfileNoAvatar(identityID)
	repoErr := errors.New("db connection failed")

	deps.userProfiles.EXPECT().GetByIdentityID(gomock.Any(), identityID).Return(existing, nil)
	deps.userProfiles.EXPECT().Delete(gomock.Any(), existing.ID).Return(repoErr)

	err := svc.HandleIdentityDeleted(context.Background(), &service.HandleIdentityDeletedInput{
		IdentityID: identityID.String(),
	})
	if err == nil {
		t.Fatal("HandleIdentityDeleted() error = nil, want error")
	}
}
