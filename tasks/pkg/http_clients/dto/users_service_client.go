// Package clientsdto
package clientsdto

import "github.com/google/uuid"

type GitUserRequest struct {
	IdentityID uuid.UUID `json:"identity_id"`
}

type GitUserResponse struct {
	Username string `json:"username"`
	GitToken string `json:"git_token"`
	GitURL   string `json:"git_url"`
}
