package postgresrepo

type Store struct {
	repo *Repo
}

func NewStore(repo *Repo) *Store {
	return &Store{repo: repo}
}

func (s *Store) ProcessedEvents() *ProcessedEventsRepo {
	return NewProcessedEventsRepo(s.repo)
}
