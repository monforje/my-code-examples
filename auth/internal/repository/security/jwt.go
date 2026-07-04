package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"auth/internal/config"
	"auth/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	errInvalidSigningMethod = errors.New("unexpected signing method")
	errTokenInvalid         = errors.New("token is invalid")
	errInvalidSubject       = errors.New("invalid token subject")
	errInvalidSessionID     = errors.New("invalid token session id")
)

type Claims struct {
	jwt.RegisteredClaims
	SessionID string `json:"sid"`
}

type Manager struct {
	signingKey     string
	accessTokenTTL time.Duration
	refreshTokenLen int
}

func NewManager(signingKey string, features config.FeaturesConfig) *Manager {
	return &Manager{
		signingKey:      signingKey,
		accessTokenTTL:  features.AccessTokenTTL,
		refreshTokenLen: features.RefreshTokenLen,
	}
}

// GenerateAccessToken - создает JWT access token с заданным userID и sessionID
//
// Возвращает токен и время его истечения
/*
	1. Создаем Claims с userID, sessionID, временем создания и истечения
	2. Создаем JWT токен с этими Claims и подписываем его
	3. Возвращаем подписанный токен и время истечения
*/
func (m *Manager) GenerateAccessToken(userID, sessionID uuid.UUID) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTokenTTL)

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ID:        uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		SessionID: sessionID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(m.signingKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

// GenerateRefreshToken - создает случайный refresh token и его хэш
/*
	1. Генерируем случайные байты для токена
	2. Кодируем их в строку (hex)
	3. Хэшируем токен для безопасного хранения
	4. Возвращаем токен и его хэш
*/
func (m *Manager) GenerateRefreshToken() (string, string, error) {
	b := make([]byte, m.refreshTokenLen)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}

	token := hex.EncodeToString(b)
	hash := utils.HashSHA256(token)

	return token, hash, nil
}

// ValidateAccessToken - проверяет валидность JWT access token
/*
	1. Парсим токен и извлекаем Claims
	2. Проверяем подпись токена
	3. Возвращаем userID, sessionID и ID токена, если он валиден
*/
func (m *Manager) ValidateAccessToken(tokenString string) (uuid.UUID, uuid.UUID, string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errInvalidSigningMethod
		}
		return []byte(m.signingKey), nil
	})
	if err != nil {
		return uuid.Nil, uuid.Nil, "", err
	}

	if !token.Valid {
		return uuid.Nil, uuid.Nil, "", errTokenInvalid
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, uuid.Nil, "", errInvalidSubject
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return uuid.Nil, uuid.Nil, "", errInvalidSessionID
	}

	return userID, sessionID, claims.ID, nil
}
