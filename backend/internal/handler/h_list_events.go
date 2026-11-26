package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type eventLister interface {
	Upcoming(ctx context.Context) ([]event.Event, error)
	All(ctx context.Context) ([]event.Event, error)
}

// ListEvents serves GET /events for the public and GET /admin/events for organizers.
type ListEvents struct {
	uc     eventLister
	logger *slog.Logger
}

// NewListEvents builds the handler.
func NewListEvents(uc eventLister, logger *slog.Logger) *ListEvents {
	return &ListEvents{uc: uc, logger: logger}
}

// HandlePublic answers with the upcoming enabled events, without organizer fields.
func (h *ListEvents) HandlePublic(ctx context.Context, _ Request) (Response, error) {
	upcoming, err := h.uc.Upcoming(ctx)
	if err != nil {
		return fail(ctx, h.logger, err), nil
	}

	return respond(http.StatusOK, map[string]any{"events": mapSlice(upcoming, toPublicEvent)}), nil
}

// HandleAdmin answers with every event.
func (h *ListEvents) HandleAdmin(ctx context.Context, _ Request) (Response, error) {
	all, err := h.uc.All(ctx)
	if err != nil {
		return fail(ctx, h.logger, err), nil
	}

	return respond(http.StatusOK, map[string]any{"events": mapSlice(all, toAdminEvent)}), nil
}
