package gittasksservice

type GitTaskCreateInput struct {
	TaskName string
}

type GitTaskCreateOutput struct {
	TaskName string
	Repo     string
	CloneURL string
}
