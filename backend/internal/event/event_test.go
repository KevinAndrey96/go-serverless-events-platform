package event_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

func validEvent(now time.Time) event.NewEventInput {
	return event.NewEventInput{
		Name: "Congreso de Ingeniería", StartsAt: now.Add(48 * time.Hour), Capacity: 100,
		PriceCOP: 50_000, Bank: "Bancolombia", Account: "123-456789-00",
	}
}

func TestNewEventInputValidate(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		mutate func(*event.NewEventInput)
		field  string
	}{
		{name: "valid paid event", mutate: func(*event.NewEventInput) {}},
		{name: "free event without bank", mutate: func(in *event.NewEventInput) { in.PriceCOP, in.Bank, in.Account = 0, "", "" }},
		{name: "blank name", mutate: func(in *event.NewEventInput) { in.Name = "  " }, field: "name"},
		{name: "long name", mutate: func(in *event.NewEventInput) { in.Name = strings.Repeat("a", 201) }, field: "name"},
		{name: "missing start", mutate: func(in *event.NewEventInput) { in.StartsAt = time.Time{} }, field: "startsAt"},
		{name: "past start", mutate: func(in *event.NewEventInput) { in.StartsAt = now.Add(-time.Minute) }, field: "startsAt"},
		{name: "zero capacity", mutate: func(in *event.NewEventInput) { in.Capacity = 0 }, field: "capacity"},
		{name: "huge capacity", mutate: func(in *event.NewEventInput) { in.Capacity = 100_001 }, field: "capacity"},
		{name: "negative price", mutate: func(in *event.NewEventInput) { in.PriceCOP = -1 }, field: "priceCop"},
		{name: "paid without account", mutate: func(in *event.NewEventInput) { in.Account = "" }, field: "account"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := validEvent(now)
			tt.mutate(&in)
			assertField(t, in.Validate(now), tt.field)
		})
	}
}

func TestRegistrationInput(t *testing.T) {
	t.Parallel()
	valid := event.RegistrationInput{FirstName: " Ana ", LastName: " Pérez ", Email: " Ana@Example.COM "}
	tests := []struct {
		name   string
		mutate func(*event.RegistrationInput)
		field  string
	}{
		{name: "valid", mutate: func(*event.RegistrationInput) {}},
		{name: "missing first name", mutate: func(in *event.RegistrationInput) { in.FirstName = " " }, field: "firstName"},
		{name: "missing last name", mutate: func(in *event.RegistrationInput) { in.LastName = "" }, field: "lastName"},
		{name: "long names", mutate: func(in *event.RegistrationInput) { in.LastName = strings.Repeat("b", 200) }, field: "lastName"},
		{name: "invalid email", mutate: func(in *event.RegistrationInput) { in.Email = "ana" }, field: "email"},
		{name: "display name email", mutate: func(in *event.RegistrationInput) { in.Email = "Ana <ana@example.com>" }, field: "email"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := valid
			tt.mutate(&in)
			assertField(t, in.Normalize().Validate(), tt.field)
		})
	}

	normalized := valid.Normalize()
	if normalized.Email != "ana@example.com" || normalized.FirstName != "Ana" || normalized.LastName != "Pérez" {
		t.Fatalf("Normalize() = %+v", normalized)
	}
}

func TestParseStatuses(t *testing.T) {
	t.Parallel()
	if s, err := event.ParseStatus("cancelled"); err != nil || s != event.StatusCancelled {
		t.Fatalf("ParseStatus(cancelled) = %q, %v", s, err)
	}
	assertField(t, second(event.ParseStatus("deleted")), "status")
	if s, err := event.ParsePaymentStatus("paid"); err != nil || s != event.PaymentPaid {
		t.Fatalf("ParsePaymentStatus(paid) = %q, %v", s, err)
	}
	assertField(t, second(event.ParsePaymentStatus("refunded")), "paymentStatus")
}

func TestEventSeats(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tests := []struct {
		name  string
		event event.Event
		seats int
		open  bool
	}{
		{name: "open", event: event.Event{Capacity: 3, Registered: 1, Status: event.StatusEnabled, StartsAt: now.Add(time.Hour)}, seats: 2, open: true},
		{name: "full", event: event.Event{Capacity: 3, Registered: 3, Status: event.StatusEnabled, StartsAt: now.Add(time.Hour)}, seats: 0},
		{name: "overbooked legacy data", event: event.Event{Capacity: 3, Registered: 5, Status: event.StatusEnabled, StartsAt: now.Add(time.Hour)}, seats: 0},
		{name: "cancelled", event: event.Event{Capacity: 3, Status: event.StatusCancelled, StartsAt: now.Add(time.Hour)}, seats: 3},
		{name: "started", event: event.Event{Capacity: 3, Status: event.StatusEnabled, StartsAt: now.Add(-time.Hour)}, seats: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.event.SeatsLeft(); got != tt.seats {
				t.Errorf("SeatsLeft() = %d, want %d", got, tt.seats)
			}
			if got := tt.event.IsOpen(now); got != tt.open {
				t.Errorf("IsOpen() = %v, want %v", got, tt.open)
			}
		})
	}
}

func TestTickets(t *testing.T) {
	t.Parallel()
	a := event.Attendee{EventID: event.NewEventID(), ID: event.NewAttendeeID(), FirstName: "Ana", LastName: "Pérez"}
	if a.FullName() != "Ana Pérez" {
		t.Fatalf("FullName() = %q", a.FullName())
	}
	eventID, attendeeID, err := event.ParseTicket(" " + a.Ticket() + "\n")
	if err != nil || eventID != a.EventID || attendeeID != a.ID {
		t.Fatalf("ParseTicket(%q) = %q, %q, %v", a.Ticket(), eventID, attendeeID, err)
	}
	if event.NewAttendeeID() == a.ID || event.NewEventID() == a.EventID {
		t.Fatal("identifiers repeat")
	}

	invalid := []string{"", "abc", "abc.", ".AAAAAAAAAAAAAAAAAAAAAA", "abc.short", "abc.AAAAAAAAAAAAAAAAAAAAA*", "User#1a2b3c"}
	for _, ticket := range invalid {
		if _, _, err := event.ParseTicket(ticket); !errors.Is(err, event.ErrInvalidTicket) {
			t.Errorf("ParseTicket(%q) error = %v, want ErrInvalidTicket", ticket, err)
		}
	}
}

func assertField(t *testing.T, err error, field string) {
	t.Helper()
	if field == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		return
	}
	var invalid event.ValidationError
	if !errors.As(err, &invalid) || invalid.Field != field {
		t.Fatalf("error = %v, want validation error on %q", err, field)
	}
	if invalid.Error() != field+" "+invalid.Reason {
		t.Fatalf("Error() = %q", invalid.Error())
	}
}

func second[T any](_ T, err error) error { return err }
