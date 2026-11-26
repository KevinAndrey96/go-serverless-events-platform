// Package handler adapts API Gateway HTTP API requests to the use cases.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

const maxBodyBytes = 16 << 10

// Request and Response are the API Gateway HTTP API (payload 2.0) types.
type (
	Request  = events.APIGatewayV2HTTPRequest
	Response = events.APIGatewayV2HTTPResponse
)

var errBadBody = errors.New("invalid json body")

type errorBody struct {
	Error   string `json:"error"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

func respond(status int, body any) Response {
	payload, err := json.Marshal(body)
	if err != nil {
		status = http.StatusInternalServerError
		payload = []byte(`{"error":"internal","message":"internal error"}`)
	}

	return Response{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type":           "application/json; charset=utf-8",
			"Cache-Control":          "no-store",
			"X-Content-Type-Options": "nosniff",
		},
		Body: string(payload),
	}
}

type domainError struct {
	err    error
	status int
	code   string
}

// domainErrors are the failures clients can act on; the code is the stable contract.
func domainErrors() []domainError {
	return []domainError{
		{event.ErrEventNotFound, http.StatusNotFound, "event_not_found"},
		{event.ErrAttendeeNotFound, http.StatusNotFound, "attendee_not_found"},
		{event.ErrInvalidTicket, http.StatusNotFound, "invalid_ticket"},
		{event.ErrEventFull, http.StatusConflict, "event_full"},
		{event.ErrEventClosed, http.StatusConflict, "event_closed"},
		{event.ErrAlreadyRegistered, http.StatusConflict, "already_registered"},
		{event.ErrAlreadyCheckedIn, http.StatusConflict, "already_checked_in"},
		{event.ErrPaymentPending, http.StatusConflict, "payment_pending"},
	}
}

// fail maps domain errors to HTTP statuses; anything unexpected is logged and hidden as a 500.
func fail(ctx context.Context, logger *slog.Logger, err error) Response {
	var invalid event.ValidationError
	if errors.As(err, &invalid) {
		return respond(http.StatusUnprocessableEntity, errorBody{Error: "validation", Field: invalid.Field, Message: invalid.Reason})
	}
	if errors.Is(err, errBadBody) {
		return respond(http.StatusBadRequest, errorBody{Error: "bad_request", Message: errBadBody.Error()})
	}
	for _, known := range domainErrors() {
		if errors.Is(err, known.err) {
			return respond(known.status, errorBody{Error: known.code, Message: known.err.Error()})
		}
	}
	logger.ErrorContext(ctx, "request", "error", err)

	return respond(http.StatusInternalServerError, errorBody{Error: "internal", Message: "internal error"})
}

func decode(req Request, target any) error {
	if len(req.Body) > maxBodyBytes || req.IsBase64Encoded {
		return errBadBody
	}
	decoder := json.NewDecoder(strings.NewReader(req.Body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errBadBody
	}

	return nil
}

type publicEvent struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	StartsAt  time.Time `json:"startsAt"`
	SeatsLeft int       `json:"seatsLeft"`
	PriceCOP  int64     `json:"priceCop"`
	Bank      string    `json:"bank,omitempty"`
	Account   string    `json:"account,omitempty"`
}

func toPublicEvent(e event.Event) publicEvent {
	return publicEvent{
		ID: e.ID, Name: e.Name, StartsAt: e.StartsAt, SeatsLeft: e.SeatsLeft(),
		PriceCOP: e.PriceCOP, Bank: e.Bank, Account: e.Account,
	}
}

type adminEvent struct {
	publicEvent

	Capacity   int       `json:"capacity"`
	Registered int       `json:"registered"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

func toAdminEvent(e event.Event) adminEvent {
	return adminEvent{
		publicEvent: toPublicEvent(e), Capacity: e.Capacity, Registered: e.Registered,
		Status: string(e.Status), CreatedAt: e.CreatedAt,
	}
}

type attendeeView struct {
	ID             string     `json:"id"`
	EventID        string     `json:"eventId"`
	Ticket         string     `json:"ticket"`
	FirstName      string     `json:"firstName"`
	LastName       string     `json:"lastName"`
	Email          string     `json:"email"`
	Phone          string     `json:"phone,omitempty"`
	Gender         string     `json:"gender,omitempty"`
	Scholarship    bool       `json:"scholarship"`
	PaymentStatus  string     `json:"paymentStatus"`
	Attendance     string     `json:"attendance"`
	RegisteredAt   time.Time  `json:"registeredAt"`
	CheckedInAt    *time.Time `json:"checkedInAt,omitempty"`
	ReminderSentAt *time.Time `json:"reminderSentAt,omitempty"`
}

func toAttendeeView(a event.Attendee) attendeeView {
	return attendeeView{
		ID: a.ID, EventID: a.EventID, Ticket: a.Ticket(), FirstName: a.FirstName, LastName: a.LastName,
		Email: a.Email, Phone: a.Phone, Gender: a.Gender, Scholarship: a.Scholarship,
		PaymentStatus: string(a.PaymentStatus), Attendance: string(a.Attendance),
		RegisteredAt: a.RegisteredAt, CheckedInAt: a.CheckedInAt, ReminderSentAt: a.ReminderSentAt,
	}
}

func mapSlice[T, V any](items []T, convert func(T) V) []V {
	out := make([]V, 0, len(items))
	for _, item := range items {
		out = append(out, convert(item))
	}

	return out
}
