package authctx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"auth/internal/authctx"
)

func TestWithAuthAndFromContext(t *testing.T) {
	identityID := uuid.New()
	sessionID := uuid.New()
	ctx := authctx.WithAuth(context.Background(), identityID, sessionID)

	gotIdentityID, gotSessionID, err := authctx.FromContext(ctx)
	if err != nil {
		t.Fatalf("FromContext() error = %v", err)
	}
	if gotIdentityID != identityID {
		t.Fatalf("FromContext() identityID = %s, want %s", gotIdentityID, identityID)
	}
	if gotSessionID != sessionID {
		t.Fatalf("FromContext() sessionID = %s, want %s", gotSessionID, sessionID)
	}
}

func TestFromContext_Missing(t *testing.T) {
	_, _, err := authctx.FromContext(context.Background())
	if !errors.Is(err, authctx.ErrAuthContextMissing) {
		t.Fatalf("FromContext() error = %v, want ErrAuthContextMissing", err)
	}
}

func TestFromContext_NilUUID(t *testing.T) {
	_, _, err := authctx.FromContext(authctx.WithAuth(context.Background(), uuid.Nil, uuid.New()))
	if !errors.Is(err, authctx.ErrAuthContextMissing) {
		t.Fatalf("FromContext() error = %v, want ErrAuthContextMissing", err)
	}
}
