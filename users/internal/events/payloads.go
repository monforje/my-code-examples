package events

type IdentityCreatedPayload struct {
	IdentityID string `json:"identity_id"`
	Email      string `json:"email"`
}

type IdentityUpdatedPayload struct {
	IdentityID    string `json:"identity_id"`
	Email         string `json:"email,omitempty"`
	Status        string `json:"status,omitempty"`
	EmailVerified *bool  `json:"email_verified,omitempty"`
}

type IdentityDeletedPayload struct {
	IdentityID string `json:"identity_id"`
}
