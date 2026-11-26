package usecase

import (
	"context"
	"fmt"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type attendeeLister interface {
	GetEvent(ctx context.Context, eventID string) (event.Event, error)
	ListAttendees(ctx context.Context, eventID string) ([]event.Attendee, error)
}

// ListAttendees returns an event with everyone registered for it.
type ListAttendees struct {
	store attendeeLister
}

// NewListAttendees builds the use case.
func NewListAttendees(store attendeeLister) *ListAttendees {
	return &ListAttendees{store: store}
}

// Execute fails with event.ErrEventNotFound for unknown events.
func (uc *ListAttendees) Execute(ctx context.Context, eventID string) (event.Event, []event.Attendee, error) {
	found, err := uc.store.GetEvent(ctx, eventID)
	if err != nil {
		return event.Event{}, nil, fmt.Errorf("attendee listing: %w", err)
	}
	attendees, err := uc.store.ListAttendees(ctx, eventID)
	if err != nil {
		return event.Event{}, nil, fmt.Errorf("attendee listing: %w", err)
	}

	return found, attendees, nil
}
