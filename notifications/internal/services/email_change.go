package services

import "context"

// SendEmailChangeCode — отправляет письмо с кодом смены почты.
func (s *NotificationService) SendEmailChangeCode(ctx context.Context, params SendCodeEmailParams) error {
	return s.sender.SendEmailChangeEmail(ctx, params)
}
