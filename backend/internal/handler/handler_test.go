package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/handler"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/usecase"
)

func discardLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func sampleEvent() event.Event {
	return event.Event{
		ID: "e1", Name: "Foro", StartsAt: time.Date(2027, 1, 10, 9, 0, 0, 0, time.UTC),
		Capacity: 10, Registered: 4, Status: event.StatusEnabled, Bank: "B", Account: "1",
	}
}

func sampleAttendee() event.Attendee {
	return event.Attendee{
		EventID: "e1", ID: "AAAAAAAAAAAAAAAAAAAAAA", FirstName: "Ana", LastName: "Pérez", Email: "ana@example.com",
		PaymentStatus: event.PaymentPaid, Attendance: event.AttendanceRegistered,
	}
}

// fakeUC answers every use case port with the configured values.
type fakeUC struct {
	err      error
	gotID    string
	gotInput any
}

func (f *fakeUC) eventResult() (event.Event, error)       { return sampleEvent(), f.err }
func (f *fakeUC) attendeeResult() (event.Attendee, error) { return sampleAttendee(), f.err }

type createUC struct{ *fakeUC }

func (u createUC) Execute(_ context.Context, in event.NewEventInput) (event.Event, error) {
	u.gotInput = in

	return u.eventResult()
}

type listUC struct{ *fakeUC }

func (u listUC) Upcoming(context.Context) ([]event.Event, error) {
	return []event.Event{sampleEvent()}, u.err
}
func (u listUC) All(context.Context) ([]event.Event, error) {
	return []event.Event{sampleEvent()}, u.err
}

type statusUC struct{ *fakeUC }

func (u statusUC) Execute(_ context.Context, eventID, rawStatus string) (event.Event, error) {
	u.gotID, u.gotInput = eventID, rawStatus

	return u.eventResult()
}

type registerUC struct{ *fakeUC }

func (u registerUC) Execute(_ context.Context, eventID string, in event.RegistrationInput) (event.Attendee, error) {
	u.gotID, u.gotInput = eventID, in

	return u.attendeeResult()
}

type attendeesUC struct{ *fakeUC }

func (u attendeesUC) Execute(_ context.Context, eventID string) (event.Event, []event.Attendee, error) {
	u.gotID = eventID

	return sampleEvent(), []event.Attendee{sampleAttendee()}, u.err
}

type paymentUC struct{ *fakeUC }

func (u paymentUC) Execute(_ context.Context, eventID, attendeeID, rawStatus string) (event.Attendee, error) {
	u.gotID, u.gotInput = eventID+"/"+attendeeID, rawStatus

	return u.attendeeResult()
}

type checkInUC struct{ *fakeUC }

func (u checkInUC) Execute(_ context.Context, ticket string) (event.Attendee, error) {
	u.gotInput = ticket

	return u.attendeeResult()
}

type remindersUC struct{ *fakeUC }

func (u remindersUC) Execute(context.Context) (usecase.ReminderReport, error) {
	return usecase.ReminderReport{Events: 1, Sent: 2}, u.err
}

type route func(fake *fakeUC) func(context.Context, handler.Request) (handler.Response, error)

func routes() map[string]route {
	logger := discardLogger()

	return map[string]route{
		"create": func(f *fakeUC) func(context.Context, handler.Request) (handler.Response, error) {
			return handler.NewCreateEvent(createUC{f}, logger).Handle
		},
		"public": func(f *fakeUC) func(context.Context, handler.Request) (handler.Response, error) {
			return handler.NewListEvents(listUC{f}, logger).HandlePublic
		},
		"admin": func(f *fakeUC) func(context.Context, handler.Request) (handler.Response, error) {
			return handler.NewListEvents(listUC{f}, logger).HandleAdmin
		},
		"status": func(f *fakeUC) func(context.Context, handler.Request) (handler.Response, error) {
			return handler.NewChangeEventStatus(statusUC{f}, logger).Handle
		},
		"register": func(f *fakeUC) func(context.Context, handler.Request) (handler.Response, error) {
			return handler.NewRegisterAttendee(registerUC{f}, logger).Handle
		},
		"attendees": func(f *fakeUC) func(context.Context, handler.Request) (handler.Response, error) {
			return handler.NewListAttendees(attendeesUC{f}, logger).Handle
		},
		"payment": func(f *fakeUC) func(context.Context, handler.Request) (handler.Response, error) {
			return handler.NewChangePaymentStatus(paymentUC{f}, logger).Handle
		},
		"checkin": func(f *fakeUC) func(context.Context, handler.Request) (handler.Response, error) {
			return handler.NewCheckIn(checkInUC{f}, logger).Handle
		},
	}
}

func TestHandlersSuccess(t *testing.T) {
	t.Parallel()
	params := map[string]string{"eventId": "e1", "attendeeId": "a1"}
	tests := []struct {
		route  string
		body   string
		status int
		field  string
		gotID  string
		input  any
	}{
		{route: "create", body: `{"name":"Foro","startsAt":"2027-01-10T09:00:00Z","capacity":10}`, status: http.StatusCreated, field: "capacity"},
		{route: "public", status: http.StatusOK, field: "events"},
		{route: "admin", status: http.StatusOK, field: "events"},
		{route: "status", body: `{"status":"cancelled"}`, status: http.StatusOK, field: "status", gotID: "e1", input: "cancelled"},
		{route: "register", body: `{"firstName":"Ana","lastName":"Pérez","email":"ana@example.com"}`, status: http.StatusCreated, field: "ticket", gotID: "e1"},
		{route: "attendees", status: http.StatusOK, field: "attendees", gotID: "e1"},
		{route: "payment", body: `{"paymentStatus":"paid"}`, status: http.StatusOK, field: "paymentStatus", gotID: "e1/a1", input: "paid"},
		{route: "checkin", body: `{"ticket":"e1.AAAAAAAAAAAAAAAAAAAAAA"}`, status: http.StatusOK, field: "ticket", input: "e1.AAAAAAAAAAAAAAAAAAAAAA"},
	}
	for _, tt := range tests {
		t.Run(tt.route, func(t *testing.T) {
			t.Parallel()
			fake := &fakeUC{}
			resp, err := routes()[tt.route](fake)(context.Background(), handler.Request{Body: tt.body, PathParameters: params})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			body := decodeBody(t, resp)
			if resp.StatusCode != tt.status {
				t.Fatalf("status = %d, want %d, body %s", resp.StatusCode, tt.status, resp.Body)
			}
			if _, ok := body[tt.field]; !ok {
				t.Fatalf("body %s has no %q", resp.Body, tt.field)
			}
			if resp.Headers["Content-Type"] != "application/json; charset=utf-8" || resp.Headers["Cache-Control"] != "no-store" {
				t.Fatalf("headers = %v", resp.Headers)
			}
			if tt.gotID != "" && fake.gotID != tt.gotID {
				t.Fatalf("path id = %q, want %q", fake.gotID, tt.gotID)
			}
			if tt.input != nil && fake.gotInput != tt.input {
				t.Fatalf("input = %v, want %v", fake.gotInput, tt.input)
			}
		})
	}
}

func TestPublicEventsHideOrganizerFields(t *testing.T) {
	t.Parallel()
	resp, err := routes()["public"](&fakeUC{})(context.Background(), handler.Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(resp.Body, "capacity") || strings.Contains(resp.Body, "registered") ||
		!strings.Contains(resp.Body, `"seatsLeft":6`) {
		t.Fatalf("public body = %s", resp.Body)
	}
}

func TestHandlerErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "validation", err: event.ValidationError{Field: "email", Reason: "is not a valid address"}, status: http.StatusUnprocessableEntity, code: "validation"},
		{name: "event not found", err: fmt.Errorf("lookup: %w", event.ErrEventNotFound), status: http.StatusNotFound, code: "event_not_found"},
		{name: "attendee not found", err: event.ErrAttendeeNotFound, status: http.StatusNotFound, code: "attendee_not_found"},
		{name: "invalid ticket", err: event.ErrInvalidTicket, status: http.StatusNotFound, code: "invalid_ticket"},
		{name: "full", err: event.ErrEventFull, status: http.StatusConflict, code: "event_full"},
		{name: "closed", err: event.ErrEventClosed, status: http.StatusConflict, code: "event_closed"},
		{name: "duplicate", err: event.ErrAlreadyRegistered, status: http.StatusConflict, code: "already_registered"},
		{name: "used ticket", err: event.ErrAlreadyCheckedIn, status: http.StatusConflict, code: "already_checked_in"},
		{name: "unpaid", err: event.ErrPaymentPending, status: http.StatusConflict, code: "payment_pending"},
		{name: "unexpected", err: errors.New("dynamodb timeout"), status: http.StatusInternalServerError, code: "internal"},
	}
	for _, tt := range tests {
		for name, build := range routes() {
			t.Run(tt.name+"/"+name, func(t *testing.T) {
				t.Parallel()
				resp, err := build(&fakeUC{err: tt.err})(context.Background(), handler.Request{Body: "{}"})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				body := decodeBody(t, resp)
				if resp.StatusCode != tt.status || body["error"] != tt.code {
					t.Fatalf("got %d %s, want %d %s", resp.StatusCode, resp.Body, tt.status, tt.code)
				}
				if tt.status == http.StatusInternalServerError && strings.Contains(resp.Body, "dynamodb") {
					t.Fatalf("internal details leaked: %s", resp.Body)
				}
			})
		}
	}
}

func TestHandlersRejectBadBodies(t *testing.T) {
	t.Parallel()
	bodies := map[string]handler.Request{
		"malformed":      {Body: "{"},
		"unknown fields": {Body: `{"isAdmin":true}`},
		"base64":         {Body: "e30=", IsBase64Encoded: true},
		"too large":      {Body: `{"name":"` + strings.Repeat("a", 20<<10) + `"}`},
	}
	for _, name := range []string{"create", "status", "register", "payment", "checkin"} {
		for kind, req := range bodies {
			t.Run(name+"/"+kind, func(t *testing.T) {
				t.Parallel()
				fake := &fakeUC{}
				resp, err := routes()[name](fake)(context.Background(), req)
				if err != nil || resp.StatusCode != http.StatusBadRequest {
					t.Fatalf("got %d %s, %v", resp.StatusCode, resp.Body, err)
				}
				if fake.gotInput != nil {
					t.Fatal("use case ran with a bad body")
				}
			})
		}
	}
}

func TestSendReminders(t *testing.T) {
	t.Parallel()
	report, err := handler.NewSendReminders(remindersUC{&fakeUC{}}, discardLogger()).Handle(context.Background())
	if err != nil || report.Sent != 2 {
		t.Fatalf("Handle() = %+v, %v", report, err)
	}
	failing := errors.New("query failed")
	if _, err := handler.NewSendReminders(remindersUC{&fakeUC{err: failing}}, discardLogger()).Handle(context.Background()); !errors.Is(err, failing) {
		t.Fatalf("error = %v, want %v", err, failing)
	}
}

func decodeBody(t *testing.T, resp handler.Response) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("invalid json %q: %v", resp.Body, err)
	}

	return body
}
