package postgresrepo

import (
	"context"

	"github.com/google/uuid"

	"auth/internal/models/records"
)

type AuthEventRepo struct {
	*Repo
}

func NewAuthEventRepo(repo *Repo) *AuthEventRepo {
	return &AuthEventRepo{Repo: repo}
}

func (r *AuthEventRepo) Create(ctx context.Context, event *records.AuthEvent) error {
	_, err := r.Exec(ctx, `
		insert into auth_events (id, identity_id, event_type, ip_address, user_agent, metadata, created_at)
		values ($1, $2, $3, $4, $5, $6, $7)
	`, event.ID, event.IdentityID, event.EventType, event.IPAddress, event.UserAgent, event.Metadata, event.CreatedAt)
	return err
}

func (r *AuthEventRepo) GetByID(ctx context.Context, id uuid.UUID) (*records.AuthEvent, error) {
	return r.scanAuthEvent(r.QueryRow(ctx, `
		select id, identity_id, event_type, ip_address, user_agent, metadata, created_at
		from auth_events
		where id = $1
	`, id))
}

func (r *AuthEventRepo) GetByIdentityID(ctx context.Context, identityID uuid.UUID) ([]*records.AuthEvent, error) {
	rows, err := r.Query(ctx, `
		select id, identity_id, event_type, ip_address, user_agent, metadata, created_at
		from auth_events
		where identity_id = $1
		order by created_at desc
	`, identityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*records.AuthEvent, 0)
	for rows.Next() {
		event, err := r.scanAuthEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *AuthEventRepo) GetByEventType(ctx context.Context, eventType string) ([]*records.AuthEvent, error) {
	rows, err := r.Query(ctx, `
		select id, identity_id, event_type, ip_address, user_agent, metadata, created_at
		from auth_events
		where event_type = $1
		order by created_at desc
	`, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*records.AuthEvent, 0)
	for rows.Next() {
		event, err := r.scanAuthEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *AuthEventRepo) scanAuthEvent(row rowScanner) (*records.AuthEvent, error) {
	event := new(records.AuthEvent)
	err := row.Scan(
		&event.ID,
		&event.IdentityID,
		&event.EventType,
		&event.IPAddress,
		&event.UserAgent,
		&event.Metadata,
		&event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return event, nil
}
