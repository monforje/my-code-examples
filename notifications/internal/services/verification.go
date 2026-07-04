package services

import "context"

// SendVerificationCode — отправляет письмо с кодом подтверждения email.
func (s *NotificationService) SendVerificationCode(ctx context.Context, params SendCodeEmailParams) error {
	return s.sender.SendVerificationEmail(ctx, params)
}
