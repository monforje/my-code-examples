package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"users/internal/config"
	"users/internal/events"
	"users/internal/models/records"
	postgresrepo "users/internal/repository/postgres"
	service "users/internal/services"
	gitauthservice "users/internal/services/git_auth"
	"users/pkg/logger"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

var testLog = logger.New(&config.LoggerConfig{Level: slog.LevelError, Format: config.FormatText, Output: io.Discard})

type mockUserProfiles struct {
	created  []*records.UserProfile
	updated  []*records.UserProfile
	deleted  []uuid.UUID
	existing map[uuid.UUID]*records.UserProfile
}

func (m *mockUserProfiles) Create(_ context.Context, p *records.UserProfile) error {
	m.created = append(m.created, p)
	return nil
}

func (m *mockUserProfiles) GetByID(_ context.Context, id uuid.UUID) (*records.UserProfile, error) {
	return nil, nil
}

func (m *mockUserProfiles) GetByEmail(_ context.Context, email string) (*records.UserProfile, error) {
	return nil, nil
}

func (m *mockUserProfiles) GetByIdentityID(_ context.Context, id uuid.UUID) (*records.UserProfile, error) {
	if p, ok := m.existing[id]; ok {
		return p, nil
	}
	return nil, postgresrepo.ErrUserProfileNotFound
}

func (m *mockUserProfiles) ExistsByDisplayName(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *mockUserProfiles) Update(_ context.Context, p *records.UserProfile) error {
	m.updated = append(m.updated, p)
	return nil
}

func (m *mockUserProfiles) Delete(_ context.Context, id uuid.UUID) error {
	m.deleted = append(m.deleted, id)
	return nil
}

type mockProcessedEvents struct {
	created  []*records.ProcessedEvent
	existing map[uuid.UUID]*records.ProcessedEvent
}

func (m *mockProcessedEvents) Create(_ context.Context, e *records.ProcessedEvent) error {
	m.created = append(m.created, e)
	return nil
}

func (m *mockProcessedEvents) GetByEventID(_ context.Context, id string) (*records.ProcessedEvent, error) {
	if e, ok := m.existing[uuid.MustParse(id)]; ok {
		return e, nil
	}
	return nil, postgresrepo.ErrProcessedEventNotFound
}

func (m *mockAvatar) Save(_ uuid.UUID, _ string, _ io.Reader) (string, string, error) {
	return "", "", nil
}

type mockAvatar struct {
	deletedKeys []string
}

func (m *mockAvatar) Delete(key string) error {
	m.deletedKeys = append(m.deletedKeys, key)
	return nil
}

func TestHandleIdentityCreated_Success(t *testing.T) {
	mp := &mockUserProfiles{}
	ma := &mockAvatar{}
	svc := service.NewUsersService(testLog, mp, ma, nil, nil)

	identityID := uuid.New()
	data, _ := json.Marshal(events.IdentityCreatedPayload{
		IdentityID: identityID.String(),
		Email:      "new@example.com",
	})

	gotID, err := handleIdentityCreated(context.Background(), svc, data)
	if err != nil {
		t.Fatalf("handleIdentityCreated() error = %v", err)
	}
	if gotID != identityID.String() {
		t.Fatalf("handleIdentityCreated() = %q, want %q", gotID, identityID.String())
	}
	if len(mp.created) != 1 {
		t.Fatalf("Create() called %d times, want 1", len(mp.created))
	}
}

func TestHandleIdentityCreated_InvalidUUID(t *testing.T) {
	svc := service.NewUsersService(testLog, nil, nil, nil, nil)

	data, _ := json.Marshal(events.IdentityCreatedPayload{
		IdentityID: "not-a-uuid",
		Email:      "test@example.com",
	})

	_, err := handleIdentityCreated(context.Background(), svc, data)
	if err == nil {
		t.Fatal("handleIdentityCreated() error = nil, want error")
	}
}

func TestHandleIdentityUpdated_Success(t *testing.T) {
	identityID := uuid.New()
	existing := &records.UserProfile{
		ID:         uuid.New(),
		IdentityID: identityID,
		Email:      "old@example.com",
		Status:     "active",
	}
	mp := &mockUserProfiles{existing: map[uuid.UUID]*records.UserProfile{identityID: existing}}

	svc := service.NewUsersService(testLog, mp, nil, nil, nil)

	newEmail := "new@example.com"
	data, _ := json.Marshal(events.IdentityUpdatedPayload{
		IdentityID: identityID.String(),
		Email:      newEmail,
	})

	gotID, err := handleIdentityUpdated(context.Background(), svc, data)
	if err != nil {
		t.Fatalf("handleIdentityUpdated() error = %v", err)
	}
	if gotID != identityID.String() {
		t.Fatalf("handleIdentityUpdated() = %q, want %q", gotID, identityID.String())
	}
	if len(mp.updated) != 1 {
		t.Fatalf("Update() called %d times, want 1", len(mp.updated))
	}
	if mp.updated[0].Email != newEmail {
		t.Fatalf("Email = %q, want %q", mp.updated[0].Email, newEmail)
	}
}

func TestHandleIdentityUpdated_ProfileNotFound(t *testing.T) {
	mp := &mockUserProfiles{existing: map[uuid.UUID]*records.UserProfile{}}
	svc := service.NewUsersService(testLog, mp, nil, nil, nil)

	data, _ := json.Marshal(events.IdentityUpdatedPayload{
		IdentityID: uuid.New().String(),
		Email:      "new@example.com",
	})

	gotID, err := handleIdentityUpdated(context.Background(), svc, data)
	if err != nil {
		t.Fatalf("handleIdentityUpdated() error = %v", err)
	}
	if len(mp.updated) != 0 {
		t.Fatalf("Update() called %d times, want 0 (profile not found)", len(mp.updated))
	}
	_ = gotID
}

func TestHandleIdentityDeleted_Success(t *testing.T) {
	identityID := uuid.New()
	profileID := uuid.New()
	existing := &records.UserProfile{
		ID:              profileID,
		IdentityID:      identityID,
		AvatarObjectKey: "avatars/test.png",
	}
	ma := &mockAvatar{}
	mp := &mockUserProfiles{existing: map[uuid.UUID]*records.UserProfile{identityID: existing}}

	svc := service.NewUsersService(testLog, mp, ma, nil, nil)

	data, _ := json.Marshal(events.IdentityDeletedPayload{
		IdentityID: identityID.String(),
	})

	gotID, err := handleIdentityDeleted(context.Background(), svc, data)
	if err != nil {
		t.Fatalf("handleIdentityDeleted() error = %v", err)
	}
	if gotID != identityID.String() {
		t.Fatalf("handleIdentityDeleted() = %q, want %q", gotID, identityID.String())
	}
	if len(mp.deleted) != 1 {
		t.Fatalf("Delete() called %d times, want 1", len(mp.deleted))
	}
	if mp.deleted[0] != profileID {
		t.Fatalf("Delete() called with %v, want %v", mp.deleted[0], profileID)
	}
	if len(ma.deletedKeys) != 1 {
		t.Fatalf("avatar.Delete() called %d times, want 1", len(ma.deletedKeys))
	}
}

func TestHandleIdentityDeleted_NoAvatar(t *testing.T) {
	identityID := uuid.New()
	existing := &records.UserProfile{
		ID:         uuid.New(),
		IdentityID: identityID,
	}
	ma := &mockAvatar{}
	mp := &mockUserProfiles{existing: map[uuid.UUID]*records.UserProfile{identityID: existing}}

	svc := service.NewUsersService(testLog, mp, ma, nil, nil)

	data, _ := json.Marshal(events.IdentityDeletedPayload{
		IdentityID: identityID.String(),
	})

	_, err := handleIdentityDeleted(context.Background(), svc, data)
	if err != nil {
		t.Fatalf("handleIdentityDeleted() error = %v", err)
	}
	if len(ma.deletedKeys) != 0 {
		t.Fatalf("avatar.Delete() called %d times, want 0 (no avatar)", len(ma.deletedKeys))
	}
}

func TestHandleIdentityDeleted_ProfileNotFound(t *testing.T) {
	mp := &mockUserProfiles{existing: map[uuid.UUID]*records.UserProfile{}}
	ma := &mockAvatar{}
	svc := service.NewUsersService(testLog, mp, ma, nil, nil)

	data, _ := json.Marshal(events.IdentityDeletedPayload{
		IdentityID: uuid.New().String(),
	})

	_, err := handleIdentityDeleted(context.Background(), svc, data)
	if err != nil {
		t.Fatalf("handleIdentityDeleted() error = %v", err)
	}
	if len(mp.deleted) != 0 {
		t.Fatalf("Delete() called %d times, want 0 (profile not found)", len(mp.deleted))
	}
}

func TestMessageHandler_Idempotent(t *testing.T) {
	eventID := uuid.New()
	identityID := uuid.New()

	pe := &mockProcessedEvents{
		existing: map[uuid.UUID]*records.ProcessedEvent{
			eventID: {EventID: eventID.String()},
		},
	}
	mp := &mockUserProfiles{}

	svc := service.NewUsersService(testLog, mp, nil, nil, nil)

	log := logger.New(&config.LoggerConfig{
		Level:  slog.LevelError,
		Format: config.FormatJSON,
	})
	c := &Consumer{
		log:             log,
		svc:             svc,
		processedEvents: pe,
		handlers: map[string]eventHandler{
			"identity.created": handleIdentityCreated,
		},
	}

	data, _ := json.Marshal(events.IdentityCreatedPayload{
		IdentityID: identityID.String(),
		Email:      "test@example.com",
	})
	envelope := EventEnvelope{
		ID:         eventID.String(),
		Type:       "identity.created",
		OccurredAt: time.Now(),
		Data:       data,
	}
	msgData, _ := json.Marshal(envelope)

	c.messageHandler(&nats.Msg{Data: msgData})

	if len(mp.created) != 0 {
		t.Fatalf("Create() called %d times, want 0 (idempotent)", len(mp.created))
	}
}

type mockGitAuth struct {
	calls chan *gitauthservice.RegisterGitUserInput
	err   error
}

func (m *mockGitAuth) RegisterGitUser(_ context.Context, input *gitauthservice.RegisterGitUserInput) (uuid.UUID, error) {
	if m.calls != nil {
		m.calls <- input
	}
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return uuid.New(), nil
}

func (m *mockGitAuth) GetGitMe(_ context.Context, identityID uuid.UUID) (*gitauthservice.GitMeResponse, error) {
	return &gitauthservice.GitMeResponse{
		Username: "test-user",
		GitToken: "test-token",
		GitURL:   "http://gitea.local",
	}, nil
}

func TestHandleIdentityCreated_RegistersGitUser(t *testing.T) {
	gitAuthCalls := make(chan *gitauthservice.RegisterGitUserInput, 1)
	ma := &mockAvatar{}
	mp := &mockUserProfiles{}
	ga := &mockGitAuth{calls: gitAuthCalls}
	svc := service.NewUsersService(testLog, mp, ma, nil, ga)

	identityID := uuid.New()
	email := "git-test@example.com"
	data, _ := json.Marshal(events.IdentityCreatedPayload{
		IdentityID: identityID.String(),
		Email:      email,
	})

	gotID, err := handleIdentityCreated(context.Background(), svc, data)
	if err != nil {
		t.Fatalf("handleIdentityCreated() error = %v", err)
	}
	if gotID != identityID.String() {
		t.Fatalf("handleIdentityCreated() = %q, want %q", gotID, identityID.String())
	}
	if len(mp.created) != 1 {
		t.Fatalf("Create() called %d times, want 1", len(mp.created))
	}

	select {
	case call := <-gitAuthCalls:
		if call.Email != email {
			t.Fatalf("RegisterGitUser email = %q, want %q", call.Email, email)
		}
		if call.ProfileID != mp.created[0].ID {
			t.Fatalf("RegisterGitUser profile_id = %v, want %v", call.ProfileID, mp.created[0].ID)
		}
		if call.Username != mp.created[0].DisplayName {
			t.Fatalf("RegisterGitUser username = %q, want %q", call.Username, mp.created[0].DisplayName)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RegisterGitUser was not called")
	}
}

func TestHandleIdentityCreated_GitAuthError(t *testing.T) {
	gitAuthCalls := make(chan *gitauthservice.RegisterGitUserInput, 1)
	ma := &mockAvatar{}
	mp := &mockUserProfiles{}
	ga := &mockGitAuth{calls: gitAuthCalls, err: errors.New("git auth unavailable")}
	svc := service.NewUsersService(testLog, mp, ma, nil, ga)

	identityID := uuid.New()
	data, _ := json.Marshal(events.IdentityCreatedPayload{
		IdentityID: identityID.String(),
		Email:      "error-test@example.com",
	})

	gotID, err := handleIdentityCreated(context.Background(), svc, data)
	if err != nil {
		t.Fatalf("handleIdentityCreated() error = %v", err)
	}
	if gotID != identityID.String() {
		t.Fatalf("handleIdentityCreated() = %q, want %q", gotID, identityID.String())
	}
	if len(mp.created) != 1 {
		t.Fatalf("Create() called %d times, want 1", len(mp.created))
	}

	select {
	case call := <-gitAuthCalls:
		if call == nil {
			t.Fatal("RegisterGitUser was not called")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RegisterGitUser was not called")
	}
}

func TestHandleIdentityCreated_GitAuthNil(t *testing.T) {
	ma := &mockAvatar{}
	mp := &mockUserProfiles{}
	svc := service.NewUsersService(testLog, mp, ma, nil, nil)

	identityID := uuid.New()
	data, _ := json.Marshal(events.IdentityCreatedPayload{
		IdentityID: identityID.String(),
		Email:      "nil-git-auth@example.com",
	})

	gotID, err := handleIdentityCreated(context.Background(), svc, data)
	if err != nil {
		t.Fatalf("handleIdentityCreated() error = %v", err)
	}
	if gotID != identityID.String() {
		t.Fatalf("handleIdentityCreated() = %q, want %q", gotID, identityID.String())
	}
	if len(mp.created) != 1 {
		t.Fatalf("Create() called %d times, want 1", len(mp.created))
	}
}
