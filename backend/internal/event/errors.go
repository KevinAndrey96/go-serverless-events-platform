package event

import (
	"errors"
	"fmt"
)

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrAttendeeNotFound  = errors.New("attendee not found")
	ErrEventFull         = errors.New("event is full")
	ErrEventClosed       = errors.New("event does not accept registrations")
	ErrAlreadyRegistered = errors.New("email already registered for this event")
	ErrAlreadyCheckedIn  = errors.New("ticket already used")
	ErrPaymentPending    = errors.New("payment not confirmed")
	ErrInvalidTicket     = errors.New("invalid ticket")
)

// ValidationError rejects an input field.
type ValidationError struct {
	Field  string
	Reason string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s %s", e.Field, e.Reason)
}
