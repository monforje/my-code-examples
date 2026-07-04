package consumer

import (
	"context"
	"encoding/json"

	"users/internal/events"
	service "users/internal/services"
)

// handleIdentityUpdated — обработка identity.updated.
/*
	1. Десериализовать data в events.IdentityUpdatedPayload.
	2. Собрать HandleIdentityUpdatedInput из payload:
	   - Email — если непустой.
	   - Status — если непустой.
	   - EmailVerified — если не nil.
	3. Вызвать svc.HandleIdentityUpdated.
	4. Вернуть identityID для processed_events.aggregate_id.
*/
func handleIdentityUpdated(ctx context.Context, svc *service.UsersService, data []byte) (string, error) {
	var payload events.IdentityUpdatedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}

	input := &service.HandleIdentityUpdatedInput{
		IdentityID: payload.IdentityID,
	}
	if payload.Email != "" {
		input.Email = &payload.Email
	}
	if payload.Status != "" {
		input.Status = &payload.Status
	}
	input.EmailVerified = payload.EmailVerified

	if err := svc.HandleIdentityUpdated(ctx, input); err != nil {
		return "", err
	}
	return payload.IdentityID, nil
}
