package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type attendeeRegistrar interface {
	GetEvent(ctx context.Context, eventID string) (event.Event, error)
	RegisterAttendee(ctx context.Context, a event.Attendee, now time.Time) error
}

// RegisterAttendee signs a person up for an event.
//
// Capacity, event status and one registration per email are enforced by the store in a single
// conditional transaction, so concurrent registrations cannot overbook an event.
type RegisterAttendee struct {
	store attendeeRegistrar
}

// NewRegisterAttendee builds the use case.
func NewRegisterAttendee(store attendeeRegistrar) *RegisterAttendee {
	return &RegisterAttendee{store: store}
}

// Execute validates the input and stores the registration; attendees of free events need no
// payment review.
func (uc *RegisterAttendee) Execute(ctx context.Context, eventID string, in event.RegistrationInput) (event.Attendee, error) {
	in = in.Normalize()
	if err := in.Validate(); err != nil {
		return event.Attendee{}, err
	}
	target, err := uc.store.GetEvent(ctx, eventID)
	if err != nil {
		return event.Attendee{}, fmt.Errorf("attendee registration: %w", err)
	}
	payment := event.PaymentPending
	if target.PriceCOP == 0 {
		payment = event.PaymentPaid
	}
	now := time.Now()
	attendee := event.Attendee{
		EventID:       eventID,
		ID:            event.NewAttendeeID(),
		FirstName:     in.FirstName,
		LastName:      in.LastName,
		Email:         in.Email,
		Phone:         in.Phone,
		Gender:        in.Gender,
		Scholarship:   in.Scholarship,
		PaymentStatus: payment,
		Attendance:    event.AttendanceRegistered,
		RegisteredAt:  now.UTC(),
	}
	if err := uc.store.RegisterAttendee(ctx, attendee, now); err != nil {
		return event.Attendee{}, fmt.Errorf("attendee registration: %w", err)
	}

	return attendee, nil
}
