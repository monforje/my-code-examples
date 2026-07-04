package backend

import (
	"context"
	"fmt"

	"codurity/internal/api"
)

type Scanner struct {
	apiClient *api.Client
}

func NewScanner(apiClient *api.Client) *Scanner {
	return &Scanner{
		apiClient: apiClient,
	}
}

func (s *Scanner) GetRepoCloneURL(ctx context.Context, owner, name string) (string, error) {
	repo, err := s.apiClient.GetRepo(ctx, owner, name)
	if err != nil {
		return "", fmt.Errorf("get repo: %w", err)
	}
	return repo.CloneURL, nil
}
