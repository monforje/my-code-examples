package postgresrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"auth/internal/models/records"
)

type DeviceAuthorizationCodesRepo struct {
	*Repo
}

func NewDeviceAuthorizationCodesRepo(r *Repo) *DeviceAuthorizationCodesRepo {
	return &DeviceAuthorizationCodesRepo{r}
}

func (r *DeviceAuthorizationCodesRepo) Create(ctx context.Context, dac *records.DeviceAuthorizationCode) error {
	_, err := r.Exec(ctx, `
		insert into device_authorization_codes (id, device_code_hash, user_code, identity_id, status, expires_at, interval, created_at, updated_at, confirmed_at, last_polled_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, dac.ID, dac.DeviceCodeHash, dac.UserCode, dac.IdentityID, dac.Status, dac.ExpiresAt, dac.Interval, dac.CreatedAt, dac.UpdatedAt, dac.ConfirmedAt, dac.LastPolledAt)
	return err
}

func (r *DeviceAuthorizationCodesRepo) GetByDeviceCodeHash(ctx context.Context, deviceCodeHash string) (*records.DeviceAuthorizationCode, error) {
	return r.scan(r.QueryRow(ctx, `
		select id, device_code_hash, user_code, identity_id, status, expires_at, interval, created_at, updated_at, confirmed_at, last_polled_at
		from device_authorization_codes
		where device_code_hash = $1
	`, deviceCodeHash))
}

func (r *DeviceAuthorizationCodesRepo) GetByUserCode(ctx context.Context, userCode string) (*records.DeviceAuthorizationCode, error) {
	return r.scan(r.QueryRow(ctx, `
		select id, device_code_hash, user_code, identity_id, status, expires_at, interval, created_at, updated_at, confirmed_at, last_polled_at
		from device_authorization_codes
		where user_code = $1
	`, userCode))
}

func (r *DeviceAuthorizationCodesRepo) Confirm(ctx context.Context, id uuid.UUID, identityID uuid.UUID) error {
	tag, err := r.Exec(ctx, `
		update device_authorization_codes
		set identity_id = $2, status = 'confirmed', confirmed_at = now(), updated_at = now()
		where id = $1 and status = 'pending' and expires_at > now()
	`, id, identityID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *DeviceAuthorizationCodesRepo) UpdateLastPolledAt(ctx context.Context, id uuid.UUID) error {
	_, err := r.Exec(ctx, `
		update device_authorization_codes
		set last_polled_at = now(), updated_at = now()
		where id = $1
	`, id)
	return err
}

func (r *DeviceAuthorizationCodesRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.Exec(ctx, `delete from device_authorization_codes where expires_at <= now()`)
	return err
}

func (r *DeviceAuthorizationCodesRepo) scan(row rowScanner) (*records.DeviceAuthorizationCode, error) {
	dac := new(records.DeviceAuthorizationCode)
	err := row.Scan(
		&dac.ID,
		&dac.DeviceCodeHash,
		&dac.UserCode,
		&dac.IdentityID,
		&dac.Status,
		&dac.ExpiresAt,
		&dac.Interval,
		&dac.CreatedAt,
		&dac.UpdatedAt,
		&dac.ConfirmedAt,
		&dac.LastPolledAt,
	)
	if err != nil {
		return nil, err
	}
	return dac, nil
}
