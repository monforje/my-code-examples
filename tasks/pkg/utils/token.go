package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashSHA256 - хеширование токена (sha256)
func HashSHA256(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}
