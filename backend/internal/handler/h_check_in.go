package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type ticketCheckIn interface {
	Execute(ctx context.Context, ticket string) (event.Attendee, error)
}

type ticketBody struct {
	Ticket string `json:"ticket"`
}

// CheckIn serves POST /admin/check-ins with the value read from the QR code.
type CheckIn struct {
	uc     ticketCheckIn
	logger *slog.Logger
}

// NewCheckIn builds the handler.
func NewCheckIn(uc ticketCheckIn, logger *slog.Logger) *CheckIn {
	return &CheckIn{uc: uc, logger: logger}
}

// Handle answers with the admitted attendee.
func (h *CheckIn) Handle(ctx context.Context, req Request) (Response, error) {
	var body ticketBody
	if err := decode(req, &body); err != nil {
		return fail(ctx, h.logger, err), nil
	}
	admitted, err := h.uc.Execute(ctx, body.Ticket)
	if err != nil {
		return fail(ctx, h.logger, err), nil
	}
	h.logger.InfoContext(ctx, "attendee checked in", "event_id", admitted.EventID, "attendee_id", admitted.ID)

	return respond(http.StatusOK, toAttendeeView(admitted)), nil
}
