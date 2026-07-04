package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// HashSHA256 - хеширование токена (sha256)
func HashSHA256(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// GenerateRandomString генерирует случайную строку указанной длины из hex-символов.
func GenerateRandomString(length int) string {
	b := make([]byte, (length+1)/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:length]
}
