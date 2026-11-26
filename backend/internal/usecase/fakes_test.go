package usecase_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

var errStore = errors.New("store unavailable")

// fakeStore keeps state in memory; set err to make every call fail.
type fakeStore struct {
	mu        sync.Mutex
	events    map[string]event.Event
	attendees map[string][]event.Attendee
	err       error
	claimErr  error
	listErr   map[string]error
	released  []string
	lastNow   time.Time
}

func newFakeStore(events ...event.Event) *fakeStore {
	s := &fakeStore{events: map[string]event.Event{}, attendees: map[string][]event.Attendee{}, listErr: map[string]error{}}
	for _, e := range events {
		s.events[e.ID] = e
	}

	return s
}

func (s *fakeStore) CreateEvent(_ context.Context, e event.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.events[e.ID] = e

	return nil
}

func (s *fakeStore) GetEvent(_ context.Context, eventID string) (event.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return event.Event{}, s.err
	}
	e, ok := s.events[eventID]
	if !ok {
		return event.Event{}, event.ErrEventNotFound
	}

	return e, nil
}

func (s *fakeStore) ListEvents(_ context.Context) ([]event.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return nil, s.err
	}
	all := make([]event.Event, 0, len(s.events))
	for _, e := range s.events {
		all = append(all, e)
	}

	return all, nil
}

func (s *fakeStore) ListEventsStartingBetween(ctx context.Context, from, to time.Time) ([]event.Event, error) {
	all, err := s.ListEvents(ctx)
	if err != nil {
		return nil, err
	}
	var between []event.Event
	for _, e := range all {
		if !e.StartsAt.Before(from) && !e.StartsAt.After(to) {
			between = append(between, e)
		}
	}

	return between, nil
}

func (s *fakeStore) UpdateEventStatus(_ context.Context, eventID string, status event.Status) (event.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return event.Event{}, s.err
	}
	e, ok := s.events[eventID]
	if !ok {
		return event.Event{}, event.ErrEventNotFound
	}
	e.Status = status
	s.events[eventID] = e

	return e, nil
}

func (s *fakeStore) RegisterAttendee(_ context.Context, a event.Attendee, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastNow = now
	if s.err != nil {
		return s.err
	}
	s.attendees[a.EventID] = append(s.attendees[a.EventID], a)

	return nil
}

func (s *fakeStore) ListAttendees(_ context.Context, eventID string) ([]event.Attendee, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.listErr[eventID]; err != nil {
		return nil, err
	}
	if s.err != nil {
		return nil, s.err
	}

	return s.attendees[eventID], nil
}

func (s *fakeStore) find(eventID, attendeeID string) (*event.Attendee, error) {
	for i := range s.attendees[eventID] {
		if s.attendees[eventID][i].ID == attendeeID {
			return &s.attendees[eventID][i], nil
		}
	}

	return nil, event.ErrAttendeeNotFound
}

func (s *fakeStore) UpdatePaymentStatus(
	_ context.Context, eventID, attendeeID string, status event.PaymentStatus,
) (event.Attendee, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return event.Attendee{}, s.err
	}
	a, err := s.find(eventID, attendeeID)
	if err != nil {
		return event.Attendee{}, err
	}
	a.PaymentStatus = status

	return *a, nil
}

func (s *fakeStore) CheckIn(_ context.Context, eventID, attendeeID string, at time.Time) (event.Attendee, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return event.Attendee{}, s.err
	}
	a, err := s.find(eventID, attendeeID)
	if err != nil {
		return event.Attendee{}, event.ErrInvalidTicket
	}
	a.Attendance, a.CheckedInAt = event.AttendanceAttended, &at

	return *a, nil
}

func (s *fakeStore) ClaimReminder(_ context.Context, eventID, attendeeID string, at time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.claimErr != nil {
		return false, s.claimErr
	}
	a, err := s.find(eventID, attendeeID)
	if err != nil {
		return false, err
	}
	if a.ReminderSentAt != nil {
		return false, nil
	}
	a.ReminderSentAt = &at

	return true, nil
}

func (s *fakeStore) ReleaseReminder(_ context.Context, eventID, attendeeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, err := s.find(eventID, attendeeID)
	if err != nil {
		return err
	}
	a.ReminderSentAt = nil
	s.released = append(s.released, attendeeID)

	return nil
}

// fakeMailer records deliveries and fails for the listed addresses.
type fakeMailer struct {
	mu     sync.Mutex
	sent   []string
	failOn map[string]bool
}

func (m *fakeMailer) SendReminder(_ context.Context, _ event.Event, a event.Attendee) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOn[a.Email] {
		return errors.New("smtp rejected")
	}
	m.sent = append(m.sent, a.Email)

	return nil
}

func discardLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }
