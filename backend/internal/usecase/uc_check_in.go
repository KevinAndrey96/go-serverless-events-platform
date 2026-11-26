package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type attendeeCheckIn interface {
	CheckIn(ctx context.Context, eventID, attendeeID string, at time.Time) (event.Attendee, error)
}

// CheckIn admits the holder of a ticket at the door.
//
// The store marks the attendee in one conditional write that requires a confirmed payment and a
// ticket not used before, so two scanners reading the same QR admit only one person.
type CheckIn struct {
	store attendeeCheckIn
}

// NewCheckIn builds the use case.
func NewCheckIn(store attendeeCheckIn) *CheckIn {
	return &CheckIn{store: store}
}

// Execute parses the ticket and records the check-in.
func (uc *CheckIn) Execute(ctx context.Context, ticket string) (event.Attendee, error) {
	eventID, attendeeID, err := event.ParseTicket(ticket)
	if err != nil {
		return event.Attendee{}, err
	}
	admitted, err := uc.store.CheckIn(ctx, eventID, attendeeID, time.Now().UTC())
	if err != nil {
		return event.Attendee{}, fmt.Errorf("check-in: %w", err)
	}

	return admitted, nil
}
