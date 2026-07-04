// Package clientsdto
package clientsdto

type RegisterGitUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type RegisterGitUserResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	GitURL   string `json:"git_url"`
}
