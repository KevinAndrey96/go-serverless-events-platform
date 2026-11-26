package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

// ReminderWindow is how far ahead the worker looks for events.
const ReminderWindow = 24 * time.Hour

type reminderStore interface {
	ListEventsStartingBetween(ctx context.Context, from, to time.Time) ([]event.Event, error)
	ListAttendees(ctx context.Context, eventID string) ([]event.Attendee, error)
	ClaimReminder(ctx context.Context, eventID, attendeeID string, at time.Time) (bool, error)
	ReleaseReminder(ctx context.Context, eventID, attendeeID string) error
}

type reminderMailer interface {
	SendReminder(ctx context.Context, e event.Event, a event.Attendee) error
}

// ReminderReport summarizes one run of the worker.
type ReminderReport struct {
	Events  int `json:"events"`
	Sent    int `json:"sent"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
}

// SendReminders emails every attendee of the events starting in the next 24 hours, once.
//
// Each attendee is claimed with a conditional write before the email goes out and released if
// sending fails, so overlapping or repeated runs never send the same reminder twice.
type SendReminders struct {
	store  reminderStore
	mailer reminderMailer
	logger *slog.Logger
}

// NewSendReminders builds the use case.
func NewSendReminders(store reminderStore, mailer reminderMailer, logger *slog.Logger) *SendReminders {
	return &SendReminders{store: store, mailer: mailer, logger: logger}
}

// Execute runs one reminder cycle; it only fails when the events cannot be read.
func (uc *SendReminders) Execute(ctx context.Context) (ReminderReport, error) {
	now := time.Now().UTC()
	upcoming, err := uc.store.ListEventsStartingBetween(ctx, now, now.Add(ReminderWindow))
	if err != nil {
		return ReminderReport{}, fmt.Errorf("reminder events: %w", err)
	}
	var report ReminderReport
	for _, e := range upcoming {
		if e.Status != event.StatusEnabled {
			continue
		}
		report.Events++
		attendees, err := uc.store.ListAttendees(ctx, e.ID)
		if err != nil {
			uc.logger.ErrorContext(ctx, "reminder attendees", "event_id", e.ID, "error", err)
			report.Failed++

			continue
		}
		for _, a := range attendees {
			uc.remind(ctx, e, a, now, &report)
		}
	}

	return report, nil
}

func (uc *SendReminders) remind(ctx context.Context, e event.Event, a event.Attendee, now time.Time, report *ReminderReport) {
	if a.ReminderSentAt != nil {
		report.Skipped++

		return
	}
	claimed, err := uc.store.ClaimReminder(ctx, e.ID, a.ID, now)
	if err != nil {
		uc.logger.ErrorContext(ctx, "reminder claim", "event_id", e.ID, "attendee_id", a.ID, "error", err)
		report.Failed++

		return
	}
	if !claimed {
		report.Skipped++

		return
	}
	if err := uc.mailer.SendReminder(ctx, e, a); err != nil {
		release := uc.store.ReleaseReminder(ctx, e.ID, a.ID)
		uc.logger.ErrorContext(ctx, "reminder delivery", "event_id", e.ID, "attendee_id", a.ID,
			"error", errors.Join(err, release))
		report.Failed++

		return
	}
	report.Sent++
}
