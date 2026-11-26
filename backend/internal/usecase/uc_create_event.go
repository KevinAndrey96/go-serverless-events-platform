// Package usecase holds the application rules; each use case depends only on the small store
// interface it needs.
package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type eventCreator interface {
	CreateEvent(ctx context.Context, e event.Event) error
}

// CreateEvent registers a new enabled event.
type CreateEvent struct {
	store eventCreator
}

// NewCreateEvent builds the use case.
func NewCreateEvent(store eventCreator) *CreateEvent {
	return &CreateEvent{store: store}
}

// Execute validates the input and stores the event.
func (uc *CreateEvent) Execute(ctx context.Context, in event.NewEventInput) (event.Event, error) {
	now := time.Now()
	if err := in.Validate(now); err != nil {
		return event.Event{}, err
	}
	created := event.Event{
		ID:        event.NewEventID(),
		Name:      strings.TrimSpace(in.Name),
		StartsAt:  in.StartsAt,
		Capacity:  in.Capacity,
		PriceCOP:  in.PriceCOP,
		Status:    event.StatusEnabled,
		Bank:      strings.TrimSpace(in.Bank),
		Account:   strings.TrimSpace(in.Account),
		CreatedAt: now.UTC(),
	}
	if err := uc.store.CreateEvent(ctx, created); err != nil {
		return event.Event{}, fmt.Errorf("event creation: %w", err)
	}

	return created, nil
}
