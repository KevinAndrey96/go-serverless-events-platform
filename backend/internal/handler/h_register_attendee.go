package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type attendeeRegistrar interface {
	Execute(ctx context.Context, eventID string, in event.RegistrationInput) (event.Attendee, error)
}

// RegisterAttendee serves POST /events/{eventId}/attendees.
type RegisterAttendee struct {
	uc     attendeeRegistrar
	logger *slog.Logger
}

// NewRegisterAttendee builds the handler.
func NewRegisterAttendee(uc attendeeRegistrar, logger *slog.Logger) *RegisterAttendee {
	return &RegisterAttendee{uc: uc, logger: logger}
}

// Handle answers 201 with the registration and its ticket.
func (h *RegisterAttendee) Handle(ctx context.Context, req Request) (Response, error) {
	var in event.RegistrationInput
	if err := decode(req, &in); err != nil {
		return fail(ctx, h.logger, err), nil
	}
	eventID := req.PathParameters["eventId"]
	registered, err := h.uc.Execute(ctx, eventID, in)
	if err != nil {
		return fail(ctx, h.logger, err), nil
	}
	h.logger.InfoContext(ctx, "attendee registered", "event_id", eventID, "attendee_id", registered.ID)

	return respond(http.StatusCreated, toAttendeeView(registered)), nil
}
