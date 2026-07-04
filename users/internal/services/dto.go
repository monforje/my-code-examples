package service

import (
	"io"
	"time"
)

// GetProfileInput — входной DTO для получения профиля.
// IdentityID извлекается из контекста через authctx.
type GetProfileInput struct{}

// ProfileOutput — выходной DTO профиля (маппинг из records.UserProfile).
type ProfileOutput struct {
	ID            string
	IdentityID    string
	Email         string
	DisplayName   *string
	Bio           string
	AvatarURL     *string
	Status        string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// UpdateSettingsInput — входной DTO для обновления настроек профиля.
// IdentityID извлекается из контекста через authctx.
type UpdateSettingsInput struct {
	DisplayName *string
	Bio         *string
}

// UpdateAvatarInput — входной DTO для загрузки аватара.
// IdentityID извлекается из контекста через authctx.
type UpdateAvatarInput struct {
	Filename string
	File     io.Reader
	FileSize int64
}

// UpdateAvatarOutput — выходной DTO после загрузки аватара.
type UpdateAvatarOutput struct {
	AvatarURL string
	UpdatedAt time.Time
}

// DeleteAvatarInput — входной DTO для удаления аватара.
// IdentityID извлекается из контекста через authctx.
type DeleteAvatarInput struct{}

// DeleteAvatarOutput — выходной DTO после удаления аватара.
type DeleteAvatarOutput struct {
	AvatarURL *string
	UpdatedAt time.Time
}

// HandleIdentityCreatedInput — входной DTO для обработки identity.created.
type HandleIdentityCreatedInput struct {
	IdentityID string
	Email      string
}

// HandleIdentityUpdatedInput — входной DTO для обновления профиля при изменении данных в auth.
type HandleIdentityUpdatedInput struct {
	IdentityID    string
	Email         *string
	Status        *string
	EmailVerified *bool
}

// HandleIdentityDeletedInput — входной DTO для soft-delete профиля при удалении аккаунта.
type HandleIdentityDeletedInput struct {
	IdentityID string
}

// GitMeOutput — выходной DTO для git-user/me.
type GitMeOutput struct {
	Username string
	GitToken string
	GitURL   string
}
