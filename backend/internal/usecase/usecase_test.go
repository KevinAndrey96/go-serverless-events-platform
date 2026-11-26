package usecase_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/usecase"
)

func upcomingEvent(id string, in time.Duration) event.Event {
	return event.Event{ID: id, Name: id, StartsAt: time.Now().Add(in), Capacity: 10, Status: event.StatusEnabled}
}

func TestCreateEvent(t *testing.T) {
	t.Parallel()
	in := event.NewEventInput{Name: "  Foro  ", StartsAt: time.Now().Add(time.Hour), Capacity: 5, Bank: " B ", Account: " 1 "}
	tests := []struct {
		name    string
		in      event.NewEventInput
		err     error
		wantErr bool
	}{
		{name: "stores a trimmed enabled event", in: in},
		{name: "rejects invalid input", in: event.NewEventInput{}, wantErr: true},
		{name: "wraps store errors", in: in, err: errStore, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := newFakeStore()
			store.err = tt.err
			created, err := usecase.NewCreateEvent(store).Execute(context.Background(), tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				if tt.err != nil && !errors.Is(err, tt.err) {
					t.Fatalf("error = %v, want wrapped %v", err, tt.err)
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if created.Name != "Foro" || created.Bank != "B" || created.Account != "1" ||
				created.Status != event.StatusEnabled || created.ID == "" || created.CreatedAt.IsZero() {
				t.Fatalf("created = %+v", created)
			}
			if _, ok := store.events[created.ID]; !ok {
				t.Fatal("event was not stored")
			}
		})
	}
}

func TestListEvents(t *testing.T) {
	t.Parallel()
	cancelled := upcomingEvent("cancelled", time.Hour)
	cancelled.Status = event.StatusCancelled
	store := newFakeStore(upcomingEvent("next", time.Hour), upcomingEvent("past", -time.Hour), cancelled)
	uc := usecase.NewListEvents(store)

	upcoming, err := uc.Upcoming(context.Background())
	if err != nil || len(upcoming) != 1 || upcoming[0].ID != "next" {
		t.Fatalf("Upcoming() = %+v, %v", upcoming, err)
	}
	all, err := uc.All(context.Background())
	if err != nil || len(all) != 3 {
		t.Fatalf("All() = %+v, %v", all, err)
	}

	store.err = errStore
	if _, err := uc.Upcoming(context.Background()); !errors.Is(err, errStore) {
		t.Fatalf("Upcoming() error = %v", err)
	}
	if _, err := uc.All(context.Background()); !errors.Is(err, errStore) {
		t.Fatalf("All() error = %v", err)
	}
}

func TestChangeEventStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		eventID string
		status  string
		want    error
	}{
		{name: "cancels", eventID: "e1", status: "cancelled"},
		{name: "unknown status", eventID: "e1", status: "archived", want: event.ValidationError{}},
		{name: "unknown event", eventID: "nope", status: "enabled", want: event.ErrEventNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := newFakeStore(upcomingEvent("e1", time.Hour))
			updated, err := usecase.NewChangeEventStatus(store).Execute(context.Background(), tt.eventID, tt.status)
			assertErr(t, err, tt.want)
			if tt.want == nil && updated.Status != event.StatusCancelled {
				t.Fatalf("status = %q", updated.Status)
			}
		})
	}
}

func TestRegisterAttendee(t *testing.T) {
	t.Parallel()
	valid := event.RegistrationInput{FirstName: "Ana", LastName: "Pérez", Email: " ANA@example.com", Scholarship: true}
	tests := []struct {
		name    string
		eventID string
		in      event.RegistrationInput
		store   error
		payment event.PaymentStatus
		want    error
	}{
		{name: "paid events start pending", eventID: "paid", in: valid, payment: event.PaymentPending},
		{name: "free events need no payment", eventID: "free", in: valid, payment: event.PaymentPaid},
		{name: "rejects invalid input", eventID: "paid", in: event.RegistrationInput{FirstName: "Ana"}, want: event.ValidationError{}},
		{name: "rejects unknown events", eventID: "missing", in: valid, want: event.ErrEventNotFound},
		{name: "propagates store errors", eventID: "paid", in: valid, store: errStore, want: errStore},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			paid := upcomingEvent("paid", time.Hour)
			paid.PriceCOP = 30_000
			store := newFakeStore(paid, upcomingEvent("free", time.Hour))
			store.err = tt.store
			before := time.Now()
			a, err := usecase.NewRegisterAttendee(store).Execute(context.Background(), tt.eventID, tt.in)
			assertErr(t, err, tt.want)
			if tt.want != nil {
				return
			}
			if a.EventID != tt.eventID || a.Email != "ana@example.com" || !a.Scholarship || len(a.ID) != 22 ||
				a.PaymentStatus != tt.payment || a.Attendance != event.AttendanceRegistered {
				t.Fatalf("attendee = %+v", a)
			}
			if store.lastNow.Before(before) {
				t.Fatal("store did not receive the registration time")
			}
		})
	}
}

func TestListAttendees(t *testing.T) {
	t.Parallel()
	store := newFakeStore(upcomingEvent("e1", time.Hour))
	store.attendees["e1"] = []event.Attendee{{ID: "a1", EventID: "e1"}}
	uc := usecase.NewListAttendees(store)

	found, attendees, err := uc.Execute(context.Background(), "e1")
	if err != nil || found.ID != "e1" || len(attendees) != 1 {
		t.Fatalf("Execute() = %+v, %+v, %v", found, attendees, err)
	}
	_, _, err = uc.Execute(context.Background(), "missing")
	assertErr(t, err, event.ErrEventNotFound)

	store.listErr["e1"] = errStore
	_, _, err = uc.Execute(context.Background(), "e1")
	assertErr(t, err, errStore)
}

func TestChangePaymentStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		attendeeID string
		status     string
		want       error
	}{
		{name: "marks paid", attendeeID: "a1", status: "paid"},
		{name: "unknown status", attendeeID: "a1", status: "refunded", want: event.ValidationError{}},
		{name: "unknown attendee", attendeeID: "zz", status: "paid", want: event.ErrAttendeeNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := newFakeStore()
			store.attendees["e1"] = []event.Attendee{{ID: "a1", EventID: "e1", PaymentStatus: event.PaymentPending}}
			updated, err := usecase.NewChangePaymentStatus(store).Execute(context.Background(), "e1", tt.attendeeID, tt.status)
			assertErr(t, err, tt.want)
			if tt.want == nil && updated.PaymentStatus != event.PaymentPaid {
				t.Fatalf("payment = %q", updated.PaymentStatus)
			}
		})
	}
}

func TestCheckIn(t *testing.T) {
	t.Parallel()
	attendee := event.Attendee{EventID: "e1", ID: event.NewAttendeeID(), PaymentStatus: event.PaymentPaid}
	tests := []struct {
		name   string
		ticket string
		want   error
	}{
		{name: "admits a valid ticket", ticket: attendee.Ticket()},
		{name: "rejects malformed tickets", ticket: "User#a1b2c3", want: event.ErrInvalidTicket},
		{name: "rejects unknown tickets", ticket: "e1." + event.NewAttendeeID(), want: event.ErrInvalidTicket},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := newFakeStore()
			store.attendees["e1"] = []event.Attendee{attendee}
			admitted, err := usecase.NewCheckIn(store).Execute(context.Background(), tt.ticket)
			assertErr(t, err, tt.want)
			if tt.want == nil && (admitted.Attendance != event.AttendanceAttended || admitted.CheckedInAt == nil) {
				t.Fatalf("admitted = %+v", admitted)
			}
		})
	}
}

func TestSendReminders(t *testing.T) {
	t.Parallel()
	sentAt := time.Now().Add(-time.Hour)
	soon, later := upcomingEvent("soon", 3*time.Hour), upcomingEvent("later", 72*time.Hour)
	cancelled := upcomingEvent("cancelled", 2*time.Hour)
	cancelled.Status = event.StatusCancelled
	broken := upcomingEvent("broken", 4*time.Hour)
	store := newFakeStore(soon, later, cancelled, broken)
	store.attendees["soon"] = []event.Attendee{
		{EventID: "soon", ID: "a1", Email: "a1@example.com"},
		{EventID: "soon", ID: "a2", Email: "a2@example.com", ReminderSentAt: &sentAt},
		{EventID: "soon", ID: "a3", Email: "bounce@example.com"},
	}
	store.attendees["later"] = []event.Attendee{{EventID: "later", ID: "b1", Email: "b1@example.com"}}
	store.attendees["cancelled"] = []event.Attendee{{EventID: "cancelled", ID: "c1", Email: "c1@example.com"}}
	store.listErr["broken"] = errStore
	mailer := &fakeMailer{failOn: map[string]bool{"bounce@example.com": true}}
	uc := usecase.NewSendReminders(store, mailer, discardLogger())

	report, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := usecase.ReminderReport{Events: 2, Sent: 1, Skipped: 1, Failed: 2}
	if report != want {
		t.Fatalf("report = %+v, want %+v", report, want)
	}
	if !slices.Equal(mailer.sent, []string{"a1@example.com"}) || !slices.Equal(store.released, []string{"a3"}) {
		t.Fatalf("sent = %v, released = %v", mailer.sent, store.released)
	}

	again, err := uc.Execute(context.Background())
	if err != nil || again.Sent != 0 || len(mailer.sent) != 1 {
		t.Fatalf("second run = %+v, %v; sent = %v", again, err, mailer.sent)
	}
}

func TestSendRemindersClaims(t *testing.T) {
	t.Parallel()
	store := newFakeStore(upcomingEvent("soon", time.Hour))
	store.attendees["soon"] = []event.Attendee{{EventID: "soon", ID: "a1", Email: "a1@example.com"}}
	store.claimErr = errStore
	mailer := &fakeMailer{}

	report, err := usecase.NewSendReminders(store, mailer, discardLogger()).Execute(context.Background())
	if err != nil || report.Failed != 1 || len(mailer.sent) != 0 {
		t.Fatalf("report = %+v, %v; sent = %v", report, err, mailer.sent)
	}

	store.err = errStore
	if _, err := usecase.NewSendReminders(store, mailer, discardLogger()).Execute(context.Background()); !errors.Is(err, errStore) {
		t.Fatalf("error = %v, want %v", err, errStore)
	}
}

func assertErr(t *testing.T, err, want error) {
	t.Helper()
	var invalid event.ValidationError
	switch {
	case want == nil && err != nil:
		t.Fatalf("unexpected error: %v", err)
	case want == nil:
	case errors.As(want, &invalid):
		if !errors.As(err, &invalid) {
			t.Fatalf("error = %v, want a validation error", err)
		}
	case !errors.Is(err, want):
		t.Fatalf("error = %v, want %v", err, want)
	}
}
