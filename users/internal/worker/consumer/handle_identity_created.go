package consumer

import (
	"context"
	"encoding/json"

	"users/internal/events"
	service "users/internal/services"
)

// handleIdentityCreated — обработка identity.created.
/*
	1. Десериализовать data в events.IdentityCreatedPayload.
	2. Вызвать svc.HandleIdentityCreated с identityID и email из payload.
	3. Вернуть identityID для processed_events.aggregate_id.
*/
func handleIdentityCreated(ctx context.Context, svc *service.UsersService, data []byte) (string, error) {
	var payload events.IdentityCreatedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	if err := svc.HandleIdentityCreated(ctx, &service.HandleIdentityCreatedInput{
		IdentityID: payload.IdentityID,
		Email:      payload.Email,
	}); err != nil {
		return "", err
	}
	return payload.IdentityID, nil
}
