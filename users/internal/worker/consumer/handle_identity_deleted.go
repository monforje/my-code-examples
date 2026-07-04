package consumer

import (
	"context"
	"encoding/json"

	"users/internal/events"
	service "users/internal/services"
)

// handleIdentityDeleted — обработка identity.deleted.
/*
	1. Десериализовать data в events.IdentityDeletedPayload.
	2. Вызвать svc.HandleIdentityDeleted с identityID из payload.
	3. Вернуть identityID для processed_events.aggregate_id.
*/
func handleIdentityDeleted(ctx context.Context, svc *service.UsersService, data []byte) (string, error) {
	var payload events.IdentityDeletedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	if err := svc.HandleIdentityDeleted(ctx, &service.HandleIdentityDeletedInput{
		IdentityID: payload.IdentityID,
	}); err != nil {
		return "", err
	}
	return payload.IdentityID, nil
}
