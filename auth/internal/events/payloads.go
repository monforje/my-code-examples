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

type IdentityLoginPayload struct {
	IdentityID string `json:"identity_id"`
	Email      string `json:"email"`
}

type IdentityLogoutPayload struct {
	IdentityID string `json:"identity_id"`
}

type VerificationCodeSendPayload struct {
	IdentityID *string `json:"identity_id,omitempty"`
	Email      string  `json:"email"`
	Code       string  `json:"code"`
	Purpose    string  `json:"purpose"`
}


