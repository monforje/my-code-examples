package tasks

type GitTaskRequest struct {
	TaskName string `json:"task_name"`
}

type GitTaskResponse struct {
	TaskName string `json:"task_name"`
	Repo     string `json:"repo"`
	CloneURL string `json:"clone_url"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
