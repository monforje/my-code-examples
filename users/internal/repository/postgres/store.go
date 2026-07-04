package postgresrepo

import "context"

type Store struct {
	repo *Repo
}

func NewStore(repo *Repo) *Store {
	return &Store{repo: repo}
}

func (s *Store) UserProfiles() *UserProfileRepo {
	return NewUserProfileRepo(s.repo)
}

func (s *Store) ProcessedEvents() *ProcessedEventsRepo {
	return NewProcessedEventsRepo(s.repo)
}

func (s *Store) GitUsers() *GitUserRepo {
	return NewGitUserRepo(s.repo)
}

func (s *Store) WithTx(ctx context.Context, fn func(*Store) error) error {
	return s.repo.WithTx(ctx, func(txRepo *Repo) error {
		return fn(NewStore(txRepo))
	})
}
