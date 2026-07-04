package services

import "context"

// SendPasswordChangeCode — отправляет письмо с кодом смены пароля.
func (s *NotificationService) SendPasswordChangeCode(ctx context.Context, params SendCodeEmailParams) error {
	return s.sender.SendPasswordChangeEmail(ctx, params)
}
