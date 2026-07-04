package authctx

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrAuthContextMissing = errors.New("auth context missing")

type contextKey string

const (
	identityIDContextKey contextKey = "identity_id"
	sessionIDContextKey  contextKey = "session_id"
)

func WithAuth(ctx context.Context, identityID, sessionID uuid.UUID) context.Context {
	ctx = context.WithValue(ctx, identityIDContextKey, identityID)
	return context.WithValue(ctx, sessionIDContextKey, sessionID)
}

func FromContext(ctx context.Context) (uuid.UUID, uuid.UUID, error) {
	identityID, ok := ctx.Value(identityIDContextKey).(uuid.UUID)
	if !ok || identityID == uuid.Nil {
		return uuid.Nil, uuid.Nil, ErrAuthContextMissing
	}

	sessionID, ok := ctx.Value(sessionIDContextKey).(uuid.UUID)
	if !ok || sessionID == uuid.Nil {
		return uuid.Nil, uuid.Nil, ErrAuthContextMissing
	}

	return identityID, sessionID, nil
}
