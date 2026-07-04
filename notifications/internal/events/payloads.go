package events

type VerificationCodeSendPayload struct {
	IdentityID *string `json:"identity_id,omitempty"`
	Email      string  `json:"email"`
	Code       string  `json:"code"`
	Purpose    string  `json:"purpose"`
}
