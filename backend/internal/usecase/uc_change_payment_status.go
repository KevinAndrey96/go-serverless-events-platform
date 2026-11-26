package usecase

import (
	"context"
	"fmt"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

type paymentStatusUpdater interface {
	UpdatePaymentStatus(ctx context.Context, eventID, attendeeID string, status event.PaymentStatus) (event.Attendee, error)
}

// ChangePaymentStatus records the review of a payment receipt.
type ChangePaymentStatus struct {
	store paymentStatusUpdater
}

// NewChangePaymentStatus builds the use case.
func NewChangePaymentStatus(store paymentStatusUpdater) *ChangePaymentStatus {
	return &ChangePaymentStatus{store: store}
}

// Execute validates the status and updates the attendee.
func (uc *ChangePaymentStatus) Execute(ctx context.Context, eventID, attendeeID, rawStatus string) (event.Attendee, error) {
	status, err := event.ParsePaymentStatus(rawStatus)
	if err != nil {
		return event.Attendee{}, err
	}
	updated, err := uc.store.UpdatePaymentStatus(ctx, eventID, attendeeID, status)
	if err != nil {
		return event.Attendee{}, fmt.Errorf("payment status change: %w", err)
	}

	return updated, nil
}
