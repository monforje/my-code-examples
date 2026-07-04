package services

import "context"

// SendPasswordResetCode — отправляет письмо с кодом восстановления пароля.
func (s *NotificationService) SendPasswordResetCode(ctx context.Context, params SendCodeEmailParams) error {
	return s.sender.SendPasswordResetEmail(ctx, params)
}
