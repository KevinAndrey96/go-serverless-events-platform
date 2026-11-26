package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type eventCreator interface {
	Execute(ctx context.Context, in event.NewEventInput) (event.Event, error)
}

// CreateEvent serves POST /admin/events.
type CreateEvent struct {
	uc     eventCreator
	logger *slog.Logger
}

// NewCreateEvent builds the handler.
func NewCreateEvent(uc eventCreator, logger *slog.Logger) *CreateEvent {
	return &CreateEvent{uc: uc, logger: logger}
}

// Handle answers 201 with the created event.
func (h *CreateEvent) Handle(ctx context.Context, req Request) (Response, error) {
	var in event.NewEventInput
	if err := decode(req, &in); err != nil {
		return fail(ctx, h.logger, err), nil
	}
	created, err := h.uc.Execute(ctx, in)
	if err != nil {
		return fail(ctx, h.logger, err), nil
	}
	h.logger.InfoContext(ctx, "event created", "event_id", created.ID)

	return respond(http.StatusCreated, toAdminEvent(created)), nil
}
