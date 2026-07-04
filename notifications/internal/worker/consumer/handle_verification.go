package consumer

import (
	"context"
	"encoding/json"
	"errors"

	"notifications/internal/events"
	service "notifications/internal/services"
)

func handleVerificationCode(ctx context.Context, svc *service.NotificationService, data []byte) (string, error) {
	var payload events.VerificationCodeSendPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	if payload.IdentityID == nil {
		return "", errors.New("identity_id is required")
	}

	return *payload.IdentityID, svc.SendVerificationCode(ctx, service.SendCodeEmailParams{
		To:   payload.Email,
		Code: payload.Code,
	})
}
