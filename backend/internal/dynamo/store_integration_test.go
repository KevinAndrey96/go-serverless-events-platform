package dynamo_test

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/dynamo"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

// newLocalStore needs DynamoDB Local, for example:
//
//	docker run --rm -p 8000:8000 amazon/dynamodb-local:3.3.2 -jar DynamoDBLocal.jar -inMemory
//	DYNAMODB_ENDPOINT=http://localhost:8000 go test ./internal/dynamo/...
func newLocalStore(t *testing.T) *dynamo.Store {
	t.Helper()
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		t.Skip("DYNAMODB_ENDPOINT is not set")
	}
	client := dynamodb.New(dynamodb.Options{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider("local", "local", ""),
	})
	table := "events-" + event.NewEventID()
	if err := dynamo.EnsureTable(t.Context(), client, table); err != nil {
		t.Fatalf("EnsureTable() = %v", err)
	}
	if err := dynamo.EnsureTable(t.Context(), client, table); err != nil {
		t.Fatalf("EnsureTable() is not idempotent: %v", err)
	}

	return dynamo.NewStore(client, table)
}

func seedEvent(t *testing.T, store *dynamo.Store, capacity int, startsIn time.Duration) event.Event {
	t.Helper()
	bogota := time.FixedZone("COT", -5*60*60)
	e := event.Event{
		ID: event.NewEventID(), Name: "Foro", StartsAt: time.Now().Add(startsIn).In(bogota).Truncate(time.Second),
		Capacity: capacity, Status: event.StatusEnabled, PriceCOP: 20_000, Bank: "B", Account: "1",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := store.CreateEvent(t.Context(), e); err != nil {
		t.Fatalf("CreateEvent() = %v", err)
	}

	return e
}

func newAttendee(eventID, email string) event.Attendee {
	return event.Attendee{
		EventID: eventID, ID: event.NewAttendeeID(), FirstName: "Ana", LastName: "Pérez", Email: email,
		PaymentStatus: event.PaymentPending, Attendance: event.AttendanceRegistered,
		RegisteredAt: time.Now().UTC().Truncate(time.Second),
	}
}

func TestEvents(t *testing.T) {
	t.Parallel()
	store := newLocalStore(t)
	ctx := t.Context()
	later := seedEvent(t, store, 5, 72*time.Hour)
	soon := seedEvent(t, store, 5, 2*time.Hour)

	got, err := store.GetEvent(ctx, soon.ID)
	if err != nil || got.Name != soon.Name || !got.StartsAt.Equal(soon.StartsAt) || got.Bank != "B" {
		t.Fatalf("GetEvent() = %+v, %v", got, err)
	}
	if err := store.CreateEvent(ctx, soon); err == nil {
		t.Fatal("CreateEvent() overwrote an existing event")
	}
	if _, err := store.GetEvent(ctx, "missing"); !errors.Is(err, event.ErrEventNotFound) {
		t.Fatalf("GetEvent(missing) = %v", err)
	}

	all, err := store.ListEvents(ctx)
	if err != nil || len(all) != 2 || all[0].ID != soon.ID || all[1].ID != later.ID {
		t.Fatalf("ListEvents() = %+v, %v", all, err)
	}
	window, err := store.ListEventsStartingBetween(ctx, time.Now(), time.Now().Add(24*time.Hour))
	if err != nil || len(window) != 1 || window[0].ID != soon.ID {
		t.Fatalf("ListEventsStartingBetween() = %+v, %v", window, err)
	}

	cancelled, err := store.UpdateEventStatus(ctx, soon.ID, event.StatusCancelled)
	if err != nil || cancelled.Status != event.StatusCancelled {
		t.Fatalf("UpdateEventStatus() = %+v, %v", cancelled, err)
	}
	if _, err := store.UpdateEventStatus(ctx, "missing", event.StatusEnabled); !errors.Is(err, event.ErrEventNotFound) {
		t.Fatalf("UpdateEventStatus(missing) = %v", err)
	}
}

func TestRegistrationNeverOverbooks(t *testing.T) {
	t.Parallel()
	store := newLocalStore(t)
	const capacity, applicants = 10, 30
	e := seedEvent(t, store, capacity, time.Hour)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = map[string]int{}
	)
	for i := range applicants {
		wg.Go(func() {
			a := newAttendee(e.ID, "persona"+string(rune('a'+i))+"@example.com")
			err := store.RegisterAttendee(t.Context(), a, time.Now())
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				results["ok"]++
			case errors.Is(err, event.ErrEventFull):
				results["full"]++
			default:
				results[err.Error()]++
			}
		})
	}
	wg.Wait()

	if results["ok"] != capacity || results["full"] != applicants-capacity {
		t.Fatalf("results = %v", results)
	}
	got, err := store.GetEvent(t.Context(), e.ID)
	if err != nil || got.Registered != capacity {
		t.Fatalf("registered = %d, %v", got.Registered, err)
	}
	attendees, err := store.ListAttendees(t.Context(), e.ID)
	if err != nil || len(attendees) != capacity {
		t.Fatalf("ListAttendees() = %d, %v", len(attendees), err)
	}
}

func TestRegistrationRules(t *testing.T) {
	t.Parallel()
	store := newLocalStore(t)
	ctx := t.Context()
	open := seedEvent(t, store, 5, time.Hour)

	first := newAttendee(open.ID, "ana@example.com")
	if err := store.RegisterAttendee(ctx, first, time.Now()); err != nil {
		t.Fatalf("RegisterAttendee() = %v", err)
	}
	if err := store.RegisterAttendee(ctx, newAttendee(open.ID, "ana@example.com"), time.Now()); !errors.Is(err, event.ErrAlreadyRegistered) {
		t.Fatalf("duplicate email = %v", err)
	}

	cancelled := seedEvent(t, store, 5, time.Hour)
	if _, err := store.UpdateEventStatus(ctx, cancelled.ID, event.StatusCancelled); err != nil {
		t.Fatal(err)
	}
	if err := store.RegisterAttendee(ctx, newAttendee(cancelled.ID, "b@example.com"), time.Now()); !errors.Is(err, event.ErrEventClosed) {
		t.Fatalf("cancelled event = %v", err)
	}
	started := seedEvent(t, store, 5, time.Hour)
	if err := store.RegisterAttendee(ctx, newAttendee(started.ID, "c@example.com"), time.Now().Add(2*time.Hour)); !errors.Is(err, event.ErrEventClosed) {
		t.Fatalf("started event = %v", err)
	}
	if err := store.RegisterAttendee(ctx, newAttendee("missing", "d@example.com"), time.Now()); !errors.Is(err, event.ErrEventNotFound) {
		t.Fatalf("missing event = %v", err)
	}

	got, err := store.GetEvent(ctx, open.ID)
	if err != nil || got.Registered != 1 {
		t.Fatalf("registered = %d, %v", got.Registered, err)
	}
}

func TestPaymentAndCheckIn(t *testing.T) {
	t.Parallel()
	store := newLocalStore(t)
	ctx := t.Context()
	e := seedEvent(t, store, 5, time.Hour)
	a := newAttendee(e.ID, "ana@example.com")
	if err := store.RegisterAttendee(ctx, a, time.Now()); err != nil {
		t.Fatal(err)
	}

	if _, err := store.CheckIn(ctx, e.ID, a.ID, time.Now()); !errors.Is(err, event.ErrPaymentPending) {
		t.Fatalf("unpaid check-in = %v", err)
	}
	paid, err := store.UpdatePaymentStatus(ctx, e.ID, a.ID, event.PaymentPaid)
	if err != nil || paid.PaymentStatus != event.PaymentPaid || paid.Email != a.Email {
		t.Fatalf("UpdatePaymentStatus() = %+v, %v", paid, err)
	}
	if _, err := store.UpdatePaymentStatus(ctx, e.ID, "missing", event.PaymentPaid); !errors.Is(err, event.ErrAttendeeNotFound) {
		t.Fatalf("UpdatePaymentStatus(missing) = %v", err)
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		admitted int
		refused  int
	)
	for range 5 {
		wg.Go(func() {
			_, err := store.CheckIn(t.Context(), e.ID, a.ID, time.Now())
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				admitted++
			} else if errors.Is(err, event.ErrAlreadyCheckedIn) {
				refused++
			}
		})
	}
	wg.Wait()
	if admitted != 1 || refused != 4 {
		t.Fatalf("admitted = %d, refused = %d", admitted, refused)
	}
	if _, err := store.CheckIn(ctx, e.ID, event.NewAttendeeID(), time.Now()); !errors.Is(err, event.ErrInvalidTicket) {
		t.Fatalf("unknown ticket = %v", err)
	}
	attendees, err := store.ListAttendees(ctx, e.ID)
	if err != nil || len(attendees) != 1 || attendees[0].Attendance != event.AttendanceAttended || attendees[0].CheckedInAt == nil {
		t.Fatalf("ListAttendees() = %+v, %v", attendees, err)
	}
}

func TestReminderClaims(t *testing.T) {
	t.Parallel()
	store := newLocalStore(t)
	ctx := t.Context()
	e := seedEvent(t, store, 5, time.Hour)
	a := newAttendee(e.ID, "ana@example.com")
	if err := store.RegisterAttendee(ctx, a, time.Now()); err != nil {
		t.Fatal(err)
	}

	claimed, err := store.ClaimReminder(ctx, e.ID, a.ID, time.Now())
	if err != nil || !claimed {
		t.Fatalf("first claim = %v, %v", claimed, err)
	}
	if claimed, err := store.ClaimReminder(ctx, e.ID, a.ID, time.Now()); err != nil || claimed {
		t.Fatalf("second claim = %v, %v", claimed, err)
	}
	if err := store.ReleaseReminder(ctx, e.ID, a.ID); err != nil {
		t.Fatalf("ReleaseReminder() = %v", err)
	}
	if claimed, err := store.ClaimReminder(ctx, e.ID, a.ID, time.Now()); err != nil || !claimed {
		t.Fatalf("claim after release = %v, %v", claimed, err)
	}
	if claimed, err := store.ClaimReminder(ctx, e.ID, "missing", time.Now()); err != nil || claimed {
		t.Fatalf("claim of a missing attendee = %v, %v", claimed, err)
	}
	if err := store.ReleaseReminder(ctx, e.ID, "missing"); err == nil {
		t.Fatal("ReleaseReminder() of a missing attendee succeeded")
	}
}
