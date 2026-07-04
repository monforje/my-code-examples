package username

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

var adjectives = []string{
	"fast", "silent", "dark", "bright", "lucky",
	"wild", "smart", "cold", "red", "blue",
	"brave", "calm", "cool", "deep", "kind",
	"keen", "pure", "sharp", "swift", "true",
}

var nouns = []string{
	"panda", "wolf", "falcon", "tiger", "fox",
	"bear", "eagle", "lion", "shark", "raven",
	"cat", "deer", "dove", "hawk", "lynx",
	"mole", "owl", "seal", "swan", "toad",
}

func randomInt(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

// GenerateUsername returns a random display name in the format adjective_noun_0000.
func GenerateUsername() (string, error) {
	adjIdx, err := randomInt(len(adjectives))
	if err != nil {
		return "", err
	}

	nounIdx, err := randomInt(len(nouns))
	if err != nil {
		return "", err
	}

	number, err := randomInt(10_000)
	if err != nil {
		return "", err
	}

	return strings.ToLower(fmt.Sprintf("%s_%s_%04d", adjectives[adjIdx], nouns[nounIdx], number)), nil
}

// ExistsFunc checks whether a display name is already taken.
type ExistsFunc func(ctx context.Context, displayName string) (bool, error)

// GenerateUniqueUsername generates a display name that is not already taken.
// Retries up to 10 times.
func GenerateUniqueUsername(ctx context.Context, exists ExistsFunc) (string, error) {
	const maxAttempts = 10

	for i := 0; i < maxAttempts; i++ {
		username, err := GenerateUsername()
		if err != nil {
			return "", err
		}

		taken, err := exists(ctx, username)
		if err != nil {
			return "", err
		}
		if !taken {
			return username, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique username after %d attempts", maxAttempts)
}
