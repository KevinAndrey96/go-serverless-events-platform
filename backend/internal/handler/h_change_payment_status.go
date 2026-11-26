package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type paymentStatusChanger interface {
	Execute(ctx context.Context, eventID, attendeeID, rawStatus string) (event.Attendee, error)
}

type paymentBody struct {
	PaymentStatus string `json:"paymentStatus"`
}

// ChangePaymentStatus serves PATCH /admin/events/{eventId}/attendees/{attendeeId}/payment.
type ChangePaymentStatus struct {
	uc     paymentStatusChanger
	logger *slog.Logger
}

// NewChangePaymentStatus builds the handler.
func NewChangePaymentStatus(uc paymentStatusChanger, logger *slog.Logger) *ChangePaymentStatus {
	return &ChangePaymentStatus{uc: uc, logger: logger}
}

// Handle answers with the updated attendee.
func (h *ChangePaymentStatus) Handle(ctx context.Context, req Request) (Response, error) {
	var body paymentBody
	if err := decode(req, &body); err != nil {
		return fail(ctx, h.logger, err), nil
	}
	eventID, attendeeID := req.PathParameters["eventId"], req.PathParameters["attendeeId"]
	updated, err := h.uc.Execute(ctx, eventID, attendeeID, body.PaymentStatus)
	if err != nil {
		return fail(ctx, h.logger, err), nil
	}
	h.logger.InfoContext(ctx, "payment status changed",
		"event_id", eventID, "attendee_id", attendeeID, "payment_status", updated.PaymentStatus)

	return respond(http.StatusOK, toAttendeeView(updated)), nil
}
