package postgresrepo

import "context"

type Store struct {
	repo *Repo
}

func NewStore(repo *Repo) *Store {
	return &Store{repo: repo}
}

func (s *Store) WithTx(ctx context.Context, fn func(*Store) error) error {
	return s.repo.WithTx(ctx, func(txRepo *Repo) error {
		return fn(NewStore(txRepo))
	})
}

func (s *Store) Tasks() *TaskRepo {
	return NewTaskRepo(s.repo)
}

func (s *Store) PulledTasks() *PulledTaskRepo {
	return NewPulledTaskRepo(s.repo)
}

func (s *Store) Reports() *ReportRepo {
	return NewReportRepo(s.repo)
}
