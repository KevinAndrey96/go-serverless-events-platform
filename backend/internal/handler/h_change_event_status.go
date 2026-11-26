package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type eventStatusChanger interface {
	Execute(ctx context.Context, eventID, rawStatus string) (event.Event, error)
}

type statusBody struct {
	Status string `json:"status"`
}

// ChangeEventStatus serves PATCH /admin/events/{eventId}/status.
type ChangeEventStatus struct {
	uc     eventStatusChanger
	logger *slog.Logger
}

// NewChangeEventStatus builds the handler.
func NewChangeEventStatus(uc eventStatusChanger, logger *slog.Logger) *ChangeEventStatus {
	return &ChangeEventStatus{uc: uc, logger: logger}
}

// Handle answers with the updated event.
func (h *ChangeEventStatus) Handle(ctx context.Context, req Request) (Response, error) {
	var body statusBody
	if err := decode(req, &body); err != nil {
		return fail(ctx, h.logger, err), nil
	}
	eventID := req.PathParameters["eventId"]
	updated, err := h.uc.Execute(ctx, eventID, body.Status)
	if err != nil {
		return fail(ctx, h.logger, err), nil
	}
	h.logger.InfoContext(ctx, "event status changed", "event_id", eventID, "status", updated.Status)

	return respond(http.StatusOK, toAdminEvent(updated)), nil
}
