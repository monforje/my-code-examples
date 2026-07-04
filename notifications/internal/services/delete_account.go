package services

import "context"

// SendDeleteAccountCode — отправляет письмо с кодом удаления аккаунта.
func (s *NotificationService) SendDeleteAccountCode(ctx context.Context, params SendCodeEmailParams) error {
	return s.sender.SendDeleteAccountEmail(ctx, params)
}
