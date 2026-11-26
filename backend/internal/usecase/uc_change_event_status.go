package usecase

import (
	"context"
	"fmt"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type eventStatusUpdater interface {
	UpdateEventStatus(ctx context.Context, eventID string, status event.Status) (event.Event, error)
}

// ChangeEventStatus enables or cancels an event.
type ChangeEventStatus struct {
	store eventStatusUpdater
}

// NewChangeEventStatus builds the use case.
func NewChangeEventStatus(store eventStatusUpdater) *ChangeEventStatus {
	return &ChangeEventStatus{store: store}
}

// Execute validates the status and updates the event.
func (uc *ChangeEventStatus) Execute(ctx context.Context, eventID, rawStatus string) (event.Event, error) {
	status, err := event.ParseStatus(rawStatus)
	if err != nil {
		return event.Event{}, err
	}
	updated, err := uc.store.UpdateEventStatus(ctx, eventID, status)
	if err != nil {
		return event.Event{}, fmt.Errorf("event status change: %w", err)
	}

	return updated, nil
}
