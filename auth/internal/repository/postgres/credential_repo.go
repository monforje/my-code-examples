package postgresrepo

import (
	"context"

	"github.com/google/uuid"

	"auth/internal/models/records"
)

type CredentialRepo struct {
	*Repo
}

func NewCredentialRepo(repo *Repo) *CredentialRepo {
	return &CredentialRepo{Repo: repo}
}

func (r *CredentialRepo) Create(ctx context.Context, credential *records.Credential) error {
	_, err := r.Exec(ctx, `
		insert into credentials (identity_id, password_hash, password_changed_at, created_at, updated_at)
		values ($1, $2, $3, $4, $5)
	`, credential.IdentityID, credential.PasswordHash, credential.PasswordChangedAt, credential.CreatedAt, credential.UpdatedAt)
	return err
}

func (r *CredentialRepo) GetByIdentityID(ctx context.Context, identityID uuid.UUID) (*records.Credential, error) {
	credential := new(records.Credential)
	err := r.QueryRow(ctx, `
		select identity_id, password_hash, password_changed_at, created_at, updated_at
		from credentials
		where identity_id = $1
	`, identityID).Scan(
		&credential.IdentityID,
		&credential.PasswordHash,
		&credential.PasswordChangedAt,
		&credential.CreatedAt,
		&credential.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return credential, nil
}

func (r *CredentialRepo) UpdatePassword(ctx context.Context, identityID uuid.UUID, passwordHash string) error {
	_, err := r.Exec(ctx, `
		update credentials
		set password_hash = $2, password_changed_at = now(), updated_at = now()
		where identity_id = $1
	`, identityID, passwordHash)
	return err
}
