package postgresrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"tasks/internal/models/records"
)

var ErrPulledTaskNotFound = errors.New("pulled task not found")

type PulledTaskRepo struct {
	*Repo
}

func NewPulledTaskRepo(repo *Repo) *PulledTaskRepo {
	return &PulledTaskRepo{Repo: repo}
}

func (r *PulledTaskRepo) Create(ctx context.Context, pt *records.PulledTask) error {
	_, err := r.Exec(ctx, `
		INSERT INTO pulled_tasks (id, identity_id, task_id, repo, clone_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, pt.ID, pt.IdentityID, pt.TaskID, pt.Repo, pt.CloneURL, pt.CreatedAt)
	return err
}

func (r *PulledTaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*records.PulledTask, error) {
	pt := new(records.PulledTask)
	err := r.QueryRow(ctx, `
		SELECT id, identity_id, task_id, repo, clone_url, created_at
		FROM pulled_tasks
		WHERE id = $1
	`, id).Scan(
		&pt.ID,
		&pt.IdentityID,
		&pt.TaskID,
		&pt.Repo,
		&pt.CloneURL,
		&pt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return pt, nil
}

func (r *PulledTaskRepo) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*records.PulledTask, error) {
	pt := new(records.PulledTask)
	err := r.QueryRow(ctx, `
		SELECT id, identity_id, task_id, repo, clone_url, created_at
		FROM pulled_tasks
		WHERE task_id = $1
	`, taskID).Scan(
		&pt.ID,
		&pt.IdentityID,
		&pt.TaskID,
		&pt.Repo,
		&pt.CloneURL,
		&pt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return pt, nil
}

func (r *PulledTaskRepo) GetByTaskIDAndIdentityID(ctx context.Context, taskID, identityID uuid.UUID) (*records.PulledTask, error) {
	pt := new(records.PulledTask)
	err := r.QueryRow(ctx, `
		SELECT id, identity_id, task_id, repo, clone_url, created_at
		FROM pulled_tasks
		WHERE task_id = $1 AND identity_id = $2
	`, taskID, identityID).Scan(
		&pt.ID,
		&pt.IdentityID,
		&pt.TaskID,
		&pt.Repo,
		&pt.CloneURL,
		&pt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return pt, nil
}

func (r *PulledTaskRepo) ListByIdentityID(ctx context.Context, identityID uuid.UUID) ([]records.PulledTask, error) {
	rows, err := r.Query(ctx, `
		SELECT id, identity_id, task_id, repo, clone_url, created_at
		FROM pulled_tasks
		WHERE identity_id = $1
		ORDER BY created_at DESC
	`, identityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []records.PulledTask
	for rows.Next() {
		var pt records.PulledTask
		if err := rows.Scan(&pt.ID, &pt.IdentityID, &pt.TaskID, &pt.Repo, &pt.CloneURL, &pt.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, pt)
	}
	return items, rows.Err()
}

func (r *PulledTaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.Exec(ctx, `
		DELETE FROM pulled_tasks
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPulledTaskNotFound
	}
	return nil
}

func (r *PulledTaskRepo) ExistsByTaskIDAndIdentityID(ctx context.Context, taskID, identityID uuid.UUID) (bool, error) {
	var exists bool
	err := r.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM pulled_tasks WHERE task_id = $1 AND identity_id = $2)
	`, taskID, identityID).Scan(&exists)
	return exists, err
}
