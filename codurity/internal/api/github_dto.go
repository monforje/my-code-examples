package api

type RepoResponse struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	CloneURL      string `json:"clone_url"`
	SSHURL        string `json:"ssh_url"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Fork          bool   `json:"fork"`
	Stargazers    int    `json:"stargazers_count"`
	Forks         int    `json:"forks_count"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	DocURL  string `json:"documentation_url"`
}
