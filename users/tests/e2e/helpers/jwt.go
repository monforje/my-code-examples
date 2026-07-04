package e2e_test_helpers

import (
	"testing"
	"time"

	"users/internal/config"
	"users/internal/repository/security"

	"github.com/google/uuid"
)

const TestJWTSigningKey = "e2e-test-signing-key-which-is-long-enough"

func NewTokenManager() *security.Manager {
	return security.NewManager(TestJWTSigningKey, config.FeaturesConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenLen: 32,
	})
}

func GenerateToken(t *testing.T, manager *security.Manager, identityID uuid.UUID) string {
	t.Helper()
	sessionID := uuid.New()
	token, _, err := manager.GenerateAccessToken(identityID, sessionID)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}
