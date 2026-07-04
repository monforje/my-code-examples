package kafkarepo

import "auth/internal/events"

type EventType = events.EventType

const (
	EventIdentityCreated = events.EventIdentityCreated
	EventIdentityUpdated = events.EventIdentityUpdated
	EventIdentityDeleted = events.EventIdentityDeleted
	EventIdentityLogin   = events.EventIdentityLogin
	EventIdentityLogout  = events.EventIdentityLogout

	EventVerificationCodeSend   = events.EventVerificationCodeSend
	EventPasswordResetCodeSend  = events.EventPasswordResetCodeSend
	EventPasswordChangeCodeSend = events.EventPasswordChangeCodeSend
	EventEmailChangeCodeSend    = events.EventEmailChangeCodeSend
	EventAccountDeleteCodeSend  = events.EventAccountDeleteCodeSend

)

type Event = events.Event
