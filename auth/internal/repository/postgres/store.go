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

func (s *Store) Identities() *IdentityRepo {
	return NewIdentityRepo(s.repo)
}

func (s *Store) Credentials() *CredentialRepo {
	return NewCredentialRepo(s.repo)
}

func (s *Store) Sessions() *SessionRepo {
	return NewSessionRepo(s.repo)
}

func (s *Store) VerificationCodes() *VerificationCodeRepo {
	return NewVerificationCodeRepo(s.repo)
}

func (s *Store) PasswordResetTokens() *PasswordResetTokenRepo {
	return NewPasswordResetTokenRepo(s.repo)
}

func (s *Store) PasswordChangeTokens() *PasswordChangeTokenRepo {
	return NewPasswordChangeTokenRepo(s.repo)
}

func (s *Store) EmailChangeRequests() *EmailChangeRequestRepo {
	return NewEmailChangeRequestRepo(s.repo)
}

func (s *Store) AccountDeleteRequests() *AccountDeleteRequestRepo {
	return NewAccountDeleteRequestRepo(s.repo)
}

func (s *Store) AuthEvents() *AuthEventRepo {
	return NewAuthEventRepo(s.repo)
}

func (s *Store) DeviceAuthorizationCodes() *DeviceAuthorizationCodesRepo {
	return NewDeviceAuthorizationCodesRepo(s.repo)
}
