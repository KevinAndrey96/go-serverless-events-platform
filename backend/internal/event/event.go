// Package event holds the domain model of the platform: events, attendees and their rules.
package event

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

// Status tells whether an event accepts registrations.
type Status string

const (
	StatusEnabled   Status = "enabled"
	StatusCancelled Status = "cancelled"
)

// PaymentStatus tracks the receipt review of an attendee.
type PaymentStatus string

const (
	PaymentPending PaymentStatus = "pending"
	PaymentPaid    PaymentStatus = "paid"
)

// Attendance tracks the door check-in of an attendee.
type Attendance string

const (
	AttendanceRegistered Attendance = "registered"
	AttendanceAttended   Attendance = "attended"
)

const (
	maxNameLength = 200
	maxCapacity   = 100_000
)

// Event is something people register for.
type Event struct {
	ID         string
	Name       string
	StartsAt   time.Time
	Capacity   int
	Registered int
	PriceCOP   int64
	Status     Status
	Bank       string
	Account    string
	CreatedAt  time.Time
}

// SeatsLeft returns how many registrations the event still accepts.
func (e Event) SeatsLeft() int {
	return max(e.Capacity-e.Registered, 0)
}

// IsOpen reports whether the event accepts registrations at the given moment.
func (e Event) IsOpen(at time.Time) bool {
	return e.Status == StatusEnabled && e.StartsAt.After(at) && e.SeatsLeft() > 0
}

// Attendee is a person registered for an event.
type Attendee struct {
	EventID        string
	ID             string
	FirstName      string
	LastName       string
	Email          string
	Phone          string
	Gender         string
	Scholarship    bool
	PaymentStatus  PaymentStatus
	Attendance     Attendance
	RegisteredAt   time.Time
	CheckedInAt    *time.Time
	ReminderSentAt *time.Time
}

// Ticket returns the value printed in the QR code.
func (a Attendee) Ticket() string {
	return a.EventID + ticketSeparator + a.ID
}

// FullName joins first and last name.
func (a Attendee) FullName() string {
	return strings.TrimSpace(a.FirstName + " " + a.LastName)
}

// NewEventInput is what an organizer submits to create an event.
type NewEventInput struct {
	Name     string    `json:"name"`
	StartsAt time.Time `json:"startsAt"`
	Capacity int       `json:"capacity"`
	PriceCOP int64     `json:"priceCop"`
	Bank     string    `json:"bank"`
	Account  string    `json:"account"`
}

// Validate checks the input against the event rules at the given moment.
func (in NewEventInput) Validate(now time.Time) error {
	switch {
	case strings.TrimSpace(in.Name) == "":
		return ValidationError{Field: "name", Reason: "is required"}
	case utf8.RuneCountInString(in.Name) > maxNameLength:
		return ValidationError{Field: "name", Reason: fmt.Sprintf("exceeds %d characters", maxNameLength)}
	case in.StartsAt.IsZero():
		return ValidationError{Field: "startsAt", Reason: "is required"}
	case !in.StartsAt.After(now):
		return ValidationError{Field: "startsAt", Reason: "must be in the future"}
	case in.Capacity < 1 || in.Capacity > maxCapacity:
		return ValidationError{Field: "capacity", Reason: fmt.Sprintf("must be between 1 and %d", maxCapacity)}
	case in.PriceCOP < 0:
		return ValidationError{Field: "priceCop", Reason: "cannot be negative"}
	case in.PriceCOP > 0 && (strings.TrimSpace(in.Bank) == "" || strings.TrimSpace(in.Account) == ""):
		return ValidationError{Field: "account", Reason: "bank and account are required for paid events"}
	}

	return nil
}

// RegistrationInput is what a person submits to register for an event.
type RegistrationInput struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Gender      string `json:"gender"`
	Scholarship bool   `json:"scholarship"`
}

// Normalize trims every field and lowercases the email.
func (in RegistrationInput) Normalize() RegistrationInput {
	return RegistrationInput{
		FirstName:   strings.TrimSpace(in.FirstName),
		LastName:    strings.TrimSpace(in.LastName),
		Email:       strings.ToLower(strings.TrimSpace(in.Email)),
		Phone:       strings.TrimSpace(in.Phone),
		Gender:      strings.TrimSpace(in.Gender),
		Scholarship: in.Scholarship,
	}
}

// Validate checks a normalized registration.
func (in RegistrationInput) Validate() error {
	switch {
	case in.FirstName == "":
		return ValidationError{Field: "firstName", Reason: "is required"}
	case in.LastName == "":
		return ValidationError{Field: "lastName", Reason: "is required"}
	case utf8.RuneCountInString(in.FirstName+in.LastName) > maxNameLength:
		return ValidationError{Field: "lastName", Reason: fmt.Sprintf("names exceed %d characters", maxNameLength)}
	}
	address, err := mail.ParseAddress(in.Email)
	if err != nil || address.Address != in.Email {
		return ValidationError{Field: "email", Reason: "is not a valid address"}
	}

	return nil
}

// ParseStatus accepts only known event statuses.
func ParseStatus(raw string) (Status, error) {
	switch Status(raw) {
	case StatusEnabled, StatusCancelled:
		return Status(raw), nil
	}

	return "", ValidationError{Field: "status", Reason: "must be enabled or cancelled"}
}

// ParsePaymentStatus accepts only known payment statuses.
func ParsePaymentStatus(raw string) (PaymentStatus, error) {
	switch PaymentStatus(raw) {
	case PaymentPending, PaymentPaid:
		return PaymentStatus(raw), nil
	}

	return "", ValidationError{Field: "paymentStatus", Reason: "must be pending or paid"}
}
