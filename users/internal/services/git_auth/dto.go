package gitauthservice

import "github.com/google/uuid"

type RegisterGitUserInput struct {
	ProfileID uuid.UUID
	Username  string
	Email     string
}

type GitMeResponse struct {
	Username string
	GitToken string
	GitURL   string
}
