package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/usecase"
)

type reminderSender interface {
	Execute(ctx context.Context) (usecase.ReminderReport, error)
}

// SendReminders runs on an EventBridge schedule.
type SendReminders struct {
	uc     reminderSender
	logger *slog.Logger
}

// NewSendReminders builds the handler.
func NewSendReminders(uc reminderSender, logger *slog.Logger) *SendReminders {
	return &SendReminders{uc: uc, logger: logger}
}

// Handle fails the invocation only when the events cannot be read; delivery failures are retried
// by the next scheduled run.
func (h *SendReminders) Handle(ctx context.Context) (usecase.ReminderReport, error) {
	report, err := h.uc.Execute(ctx)
	if err != nil {
		return usecase.ReminderReport{}, fmt.Errorf("reminders: %w", err)
	}
	h.logger.InfoContext(ctx, "reminders sent",
		"events", report.Events, "sent", report.Sent, "skipped", report.Skipped, "failed", report.Failed)

	return report, nil
}
