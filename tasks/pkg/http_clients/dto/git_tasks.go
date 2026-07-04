// Package clientsdto
package clientsdto

type GitTaskCreateRequest struct {
	Username string `json:"username"`
	TaskID   string `json:"task_id"`
}

type GitTaskCreateResponse struct {
	TaskID   string `json:"task_id"`
	Repo     string `json:"repo"`
	CloneURL string `json:"clone_url"`
}
