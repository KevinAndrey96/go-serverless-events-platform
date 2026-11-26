package mail

import (
	"context"
	"log/slog"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

// Log writes reminders to the log instead of sending them; the local dev server uses it.
type Log struct {
	logger *slog.Logger
}

// NewLog builds the mailer.
func NewLog(logger *slog.Logger) *Log {
	return &Log{logger: logger}
}

// SendReminder logs the reminder that would be sent.
func (m *Log) SendReminder(ctx context.Context, e event.Event, a event.Attendee) error {
	m.logger.InfoContext(ctx, "reminder", "event_id", e.ID, "attendee_id", a.ID, "to", a.Email)

	return nil
}
