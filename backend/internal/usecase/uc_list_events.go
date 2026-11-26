package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type eventLister interface {
	ListEvents(ctx context.Context) ([]event.Event, error)
}

// ListEvents returns events ordered by start time.
type ListEvents struct {
	store eventLister
}

// NewListEvents builds the use case.
func NewListEvents(store eventLister) *ListEvents {
	return &ListEvents{store: store}
}

// Upcoming returns the enabled events that have not started, for the public portal.
func (uc *ListEvents) Upcoming(ctx context.Context) ([]event.Event, error) {
	all, err := uc.store.ListEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("upcoming events: %w", err)
	}
	now := time.Now()
	upcoming := make([]event.Event, 0, len(all))
	for _, e := range all {
		if e.Status == event.StatusEnabled && e.StartsAt.After(now) {
			upcoming = append(upcoming, e)
		}
	}

	return upcoming, nil
}

// All returns every event, for organizers.
func (uc *ListEvents) All(ctx context.Context) ([]event.Event, error) {
	all, err := uc.store.ListEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("all events: %w", err)
	}

	return all, nil
}
