package postgresrepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"auth/internal/models/records"
	postgres "auth/internal/repository/postgres"
)

func TestDeviceAuthorizationCodesRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	dacRepo := postgres.NewDeviceAuthorizationCodesRepo(repo)
	cleanupTable(t, repo, "device_authorization_codes")

	now := time.Now()
	dac := &records.DeviceAuthorizationCode{
		ID:             uuid.New(),
		DeviceCodeHash: "device_hash_" + uuid.New().String()[:8],
		UserCode:       "ABCD-EFGH",
		Status:         "pending",
		ExpiresAt:      now.Add(10 * time.Minute),
		Interval:       3,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := dacRepo.Create(context.Background(), dac)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := dacRepo.GetByDeviceCodeHash(context.Background(), dac.DeviceCodeHash)
	if err != nil {
		t.Fatalf("GetByDeviceCodeHash() error = %v", err)
	}

	if got.UserCode != dac.UserCode {
		t.Errorf("UserCode = %v, want %v", got.UserCode, dac.UserCode)
	}
	if got.Status != "pending" {
		t.Errorf("Status = %v, want pending", got.Status)
	}
}

func TestDeviceAuthorizationCodesRepo_GetByDeviceCodeHash_NotFound(t *testing.T) {
	repo := newTestDB(t)
	dacRepo := postgres.NewDeviceAuthorizationCodesRepo(repo)
	cleanupTable(t, repo, "device_authorization_codes")

	_, err := dacRepo.GetByDeviceCodeHash(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeviceAuthorizationCodesRepo_GetByUserCode(t *testing.T) {
	repo := newTestDB(t)
	dacRepo := postgres.NewDeviceAuthorizationCodesRepo(repo)
	cleanupTable(t, repo, "device_authorization_codes")

	now := time.Now()
	userCode := "XXXX-YYYY"
	dac := &records.DeviceAuthorizationCode{
		ID:             uuid.New(),
		DeviceCodeHash: "hash_" + uuid.New().String()[:8],
		UserCode:       userCode,
		Status:         "pending",
		ExpiresAt:      now.Add(10 * time.Minute),
		Interval:       3,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := dacRepo.Create(context.Background(), dac); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := dacRepo.GetByUserCode(context.Background(), userCode)
	if err != nil {
		t.Fatalf("GetByUserCode() error = %v", err)
	}

	if got.DeviceCodeHash != dac.DeviceCodeHash {
		t.Errorf("DeviceCodeHash = %v, want %v", got.DeviceCodeHash, dac.DeviceCodeHash)
	}
}

func TestDeviceAuthorizationCodesRepo_Confirm(t *testing.T) {
	repo := newTestDB(t)
	dacRepo := postgres.NewDeviceAuthorizationCodesRepo(repo)
	cleanupTable(t, repo, "device_authorization_codes")

	identityID := createTestIdentity(t, repo)
	now := time.Now()
	dac := &records.DeviceAuthorizationCode{
		ID:             uuid.New(),
		DeviceCodeHash: "confirm_hash_" + uuid.New().String()[:8],
		UserCode:       "CONF-IRM",
		Status:         "pending",
		ExpiresAt:      now.Add(10 * time.Minute),
		Interval:       3,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := dacRepo.Create(context.Background(), dac); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := dacRepo.Confirm(context.Background(), dac.ID, identityID); err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}

	got, err := dacRepo.GetByDeviceCodeHash(context.Background(), dac.DeviceCodeHash)
	if err != nil {
		t.Fatalf("GetByDeviceCodeHash() error = %v", err)
	}

	if got.Status != "confirmed" {
		t.Errorf("Status = %v, want confirmed", got.Status)
	}
	if got.IdentityID == nil {
		t.Fatal("IdentityID is nil, want non-nil")
	}
	if *got.IdentityID != identityID {
		t.Errorf("IdentityID = %v, want %v", *got.IdentityID, identityID)
	}
	if got.ConfirmedAt == nil {
		t.Error("ConfirmedAt is nil, want non-nil")
	}
}

func TestDeviceAuthorizationCodesRepo_Confirm_AlreadyConfirmed(t *testing.T) {
	repo := newTestDB(t)
	dacRepo := postgres.NewDeviceAuthorizationCodesRepo(repo)
	cleanupTable(t, repo, "device_authorization_codes")

	identityID := createTestIdentity(t, repo)
	now := time.Now()
	dac := &records.DeviceAuthorizationCode{
		ID:             uuid.New(),
		DeviceCodeHash: "already_hash_" + uuid.New().String()[:8],
		UserCode:       "ALRD-CONF",
		Status:         "pending",
		ExpiresAt:      now.Add(10 * time.Minute),
		Interval:       3,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := dacRepo.Create(context.Background(), dac); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := dacRepo.Confirm(context.Background(), dac.ID, identityID); err != nil {
		t.Fatalf("Confirm() first call error = %v", err)
	}

	err := dacRepo.Confirm(context.Background(), dac.ID, identityID)
	if err == nil {
		t.Fatal("expected error on second Confirm(), got nil")
	}
}

func TestDeviceAuthorizationCodesRepo_Confirm_Expired(t *testing.T) {
	repo := newTestDB(t)
	dacRepo := postgres.NewDeviceAuthorizationCodesRepo(repo)
	cleanupTable(t, repo, "device_authorization_codes")

	identityID := createTestIdentity(t, repo)
	now := time.Now()
	dac := &records.DeviceAuthorizationCode{
		ID:             uuid.New(),
		DeviceCodeHash: "expired_hash_" + uuid.New().String()[:8],
		UserCode:       "EXPI-RED",
		Status:         "pending",
		ExpiresAt:      now.Add(-1 * time.Minute), // already expired
		Interval:       3,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := dacRepo.Create(context.Background(), dac); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := dacRepo.Confirm(context.Background(), dac.ID, identityID)
	if err == nil {
		t.Fatal("expected error for expired code, got nil")
	}
}

func TestDeviceAuthorizationCodesRepo_UpdateLastPolledAt(t *testing.T) {
	repo := newTestDB(t)
	dacRepo := postgres.NewDeviceAuthorizationCodesRepo(repo)
	cleanupTable(t, repo, "device_authorization_codes")

	now := time.Now()
	dac := &records.DeviceAuthorizationCode{
		ID:             uuid.New(),
		DeviceCodeHash: "poll_hash_" + uuid.New().String()[:8],
		UserCode:       "POLL-TEST",
		Status:         "pending",
		ExpiresAt:      now.Add(10 * time.Minute),
		Interval:       3,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := dacRepo.Create(context.Background(), dac); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := dacRepo.UpdateLastPolledAt(context.Background(), dac.ID); err != nil {
		t.Fatalf("UpdateLastPolledAt() error = %v", err)
	}

	got, err := dacRepo.GetByDeviceCodeHash(context.Background(), dac.DeviceCodeHash)
	if err != nil {
		t.Fatalf("GetByDeviceCodeHash() error = %v", err)
	}

	if got.LastPolledAt == nil {
		t.Error("LastPolledAt is nil, want non-nil")
	}
}

func TestDeviceAuthorizationCodesRepo_DeleteExpired(t *testing.T) {
	repo := newTestDB(t)
	dacRepo := postgres.NewDeviceAuthorizationCodesRepo(repo)
	cleanupTable(t, repo, "device_authorization_codes")

	now := time.Now()

	// Create expired record
	expired := &records.DeviceAuthorizationCode{
		ID:             uuid.New(),
		DeviceCodeHash: "expired_del_" + uuid.New().String()[:8],
		UserCode:       "EXPI-DEL",
		Status:         "pending",
		ExpiresAt:      now.Add(-1 * time.Minute),
		Interval:       3,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := dacRepo.Create(context.Background(), expired); err != nil {
		t.Fatalf("Create() expired error = %v", err)
	}

	// Create valid record
	valid := &records.DeviceAuthorizationCode{
		ID:             uuid.New(),
		DeviceCodeHash: "valid_del_" + uuid.New().String()[:8],
		UserCode:       "VALID-DEL",
		Status:         "pending",
		ExpiresAt:      now.Add(10 * time.Minute),
		Interval:       3,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := dacRepo.Create(context.Background(), valid); err != nil {
		t.Fatalf("Create() valid error = %v", err)
	}

	if err := dacRepo.DeleteExpired(context.Background()); err != nil {
		t.Fatalf("DeleteExpired() error = %v", err)
	}

	// expired should be gone
	_, err := dacRepo.GetByDeviceCodeHash(context.Background(), expired.DeviceCodeHash)
	if err == nil {
		t.Fatal("expected error for expired record, got nil")
	}

	// valid should still exist
	_, err = dacRepo.GetByDeviceCodeHash(context.Background(), valid.DeviceCodeHash)
	if err != nil {
		t.Fatalf("GetByDeviceCodeHash() valid record error = %v", err)
	}
}
