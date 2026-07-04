package username_test

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"users/internal/generate/username"
)

func TestGenerateUsername_Format(t *testing.T) {
	name, err := username.GenerateUsername()
	if err != nil {
		t.Fatalf("GenerateUsername() error = %v", err)
	}

	matched, _ := regexp.MatchString(`^[a-z]+_[a-z]+_\d{4}$`, name)
	if !matched {
		t.Errorf("GenerateUsername() = %q, want format adjective_noun_0000", name)
	}
}

func TestGenerateUsername_Lowercase(t *testing.T) {
	for i := 0; i < 100; i++ {
		name, err := username.GenerateUsername()
		if err != nil {
			t.Fatalf("GenerateUsername() error = %v", err)
		}
		if name != strings.ToLower(name) {
			t.Errorf("GenerateUsername() = %q, want lowercase", name)
		}
	}
}

func TestGenerateUsername_Randomness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		name, err := username.GenerateUsername()
		if err != nil {
			t.Fatalf("GenerateUsername() error = %v", err)
		}
		if seen[name] {
			t.Errorf("GenerateUsername() duplicate = %q", name)
		}
		seen[name] = true
	}
}

func TestGenerateUniqueUsername_Success(t *testing.T) {
	name, err := username.GenerateUniqueUsername(context.Background(), func(_ context.Context, _ string) (bool, error) {
		return false, nil
	})
	if err != nil {
		t.Fatalf("GenerateUniqueUsername() error = %v", err)
	}
	if name == "" {
		t.Fatal("GenerateUniqueUsername() returned empty string")
	}
}

func TestGenerateUniqueUsername_AllTaken(t *testing.T) {
	_, err := username.GenerateUniqueUsername(context.Background(), func(_ context.Context, _ string) (bool, error) {
		return true, nil
	})
	if err == nil {
		t.Fatal("GenerateUniqueUsername() error = nil, want error (all names taken)")
	}
	if !strings.Contains(err.Error(), "failed to generate unique username") {
		t.Errorf("error = %q, want 'failed to generate unique username'", err.Error())
	}
}

func TestGenerateUniqueUsername_ExistsError(t *testing.T) {
	_, err := username.GenerateUniqueUsername(context.Background(), func(_ context.Context, _ string) (bool, error) {
		return false, context.Canceled
	})
	if err == nil {
		t.Fatal("GenerateUniqueUsername() error = nil, want error")
	}
}
