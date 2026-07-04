package postgresrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"users/internal/models/records"
)

var ErrGitUserNotFound = errors.New("git user not found")

type GitUserRepo struct {
	*Repo
}

func NewGitUserRepo(r *Repo) *GitUserRepo {
	return &GitUserRepo{
		Repo: r,
	}
}

func (r *GitUserRepo) Create(ctx context.Context, gitUser *records.GitUser) error {
	_, err := r.Exec(ctx, `
		insert into git_users (id, profile_id, git_token, git_url, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6)
	`, gitUser.ID, gitUser.ProfileID, gitUser.GitToken, gitUser.GitURL, gitUser.CreatedAt, gitUser.UpdatedAt)
	return err
}

func (r *GitUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*records.GitUser, error) {
	gitUser := new(records.GitUser)
	err := r.QueryRow(ctx, `
		select id, profile_id, git_token, git_url, created_at, updated_at
		from git_users
		where id = $1
	`, id).Scan(
		&gitUser.ID,
		&gitUser.ProfileID,
		&gitUser.GitToken,
		&gitUser.GitURL,
		&gitUser.CreatedAt,
		&gitUser.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return gitUser, nil
}

func (r *GitUserRepo) GetByProfileID(ctx context.Context, ProfileID uuid.UUID) ([]*records.GitUser, error) {
	rows, err := r.Query(ctx, `
		select id, profile_id, git_token, git_url, created_at, updated_at
		from git_users
		where profile_id = $1
		order by created_at desc
	`, ProfileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gitUsers := make([]*records.GitUser, 0)
	for rows.Next() {
		gitUser := new(records.GitUser)
		if err := rows.Scan(
			&gitUser.ID,
			&gitUser.ProfileID,
			&gitUser.GitToken,
			&gitUser.GitURL,
			&gitUser.CreatedAt,
			&gitUser.UpdatedAt,
		); err != nil {
			return nil, err
		}
		gitUsers = append(gitUsers, gitUser)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return gitUsers, nil
}

func (r *GitUserRepo) GetByProfileIDAndGitURL(ctx context.Context, ProfileID uuid.UUID, gitURL string) (*records.GitUser, error) {
	gitUser := new(records.GitUser)
	err := r.QueryRow(ctx, `
		select id, profile_id, git_token, git_url, created_at, updated_at
		from git_users
		where profile_id = $1 and git_url = $2
	`, ProfileID, gitURL).Scan(
		&gitUser.ID,
		&gitUser.ProfileID,
		&gitUser.GitToken,
		&gitUser.GitURL,
		&gitUser.CreatedAt,
		&gitUser.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return gitUser, nil
}

func (r *GitUserRepo) Update(ctx context.Context, gitUser *records.GitUser) error {
	tag, err := r.Exec(ctx, `
		update git_users
		set git_token = $2, git_url = $3, updated_at = $4
		where id = $1
	`, gitUser.ID, gitUser.GitToken, gitUser.GitURL, gitUser.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *GitUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.Exec(ctx, `
		delete from git_users
		where id = $1
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *GitUserRepo) DeleteByProfileID(ctx context.Context, ProfileID uuid.UUID) error {
	_, err := r.Exec(ctx, `
		delete from git_users
		where profile_id = $1
	`, ProfileID)
	return err
}
