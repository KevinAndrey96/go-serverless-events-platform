package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type attendeeLister interface {
	Execute(ctx context.Context, eventID string) (event.Event, []event.Attendee, error)
}

// ListAttendees serves GET /admin/events/{eventId}/attendees.
type ListAttendees struct {
	uc     attendeeLister
	logger *slog.Logger
}

// NewListAttendees builds the handler.
func NewListAttendees(uc attendeeLister, logger *slog.Logger) *ListAttendees {
	return &ListAttendees{uc: uc, logger: logger}
}

// Handle answers with the event and its attendees.
func (h *ListAttendees) Handle(ctx context.Context, req Request) (Response, error) {
	found, attendees, err := h.uc.Execute(ctx, req.PathParameters["eventId"])
	if err != nil {
		return fail(ctx, h.logger, err), nil
	}

	return respond(http.StatusOK, map[string]any{
		"event":     toAdminEvent(found),
		"attendees": mapSlice(attendees, toAttendeeView),
	}), nil
}
