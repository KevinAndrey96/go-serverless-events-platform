// Package dynamo stores events and attendees in a single DynamoDB table.
//
//	Entity          PK               SK                    GSI1PK   GSI1SK
//	Event           EVENT#<id>       EVENT                 EVENTS   <startsAt UTC>#<id>
//	Attendee        EVENT#<id>       ATTENDEE#<attendee>
//	Email guard     EVENT#<id>       EMAIL#<email>
//
// An event and its attendees share a partition, so listing attendees is one Query. The email
// guard item makes "one registration per email and event" a conditional write.
package dynamo

import (
	"time"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

const (
	eventPrefix    = "EVENT#"
	eventSK        = "EVENT"
	attendeePrefix = "ATTENDEE#"
	emailPrefix    = "EMAIL#"
	eventsGSI1PK   = "EVENTS"

	// GSI1 lists events ordered by start time.
	GSI1 = "GSI1"
)

func eventPK(eventID string) string       { return eventPrefix + eventID }
func attendeeSK(attendeeID string) string { return attendeePrefix + attendeeID }
func emailSK(email string) string         { return emailPrefix + email }
func eventGSI1SK(e event.Event) string    { return startsKey(e.StartsAt) + "#" + e.ID }
func startsKey(t time.Time) string        { return t.UTC().Format(time.RFC3339) }

type eventItem struct {
	PK         string    `dynamodbav:"PK"`
	SK         string    `dynamodbav:"SK"`
	GSI1PK     string    `dynamodbav:"GSI1PK"`
	GSI1SK     string    `dynamodbav:"GSI1SK"`
	Type       string    `dynamodbav:"type"`
	ID         string    `dynamodbav:"id"`
	Name       string    `dynamodbav:"name"`
	StartsAt   time.Time `dynamodbav:"startsAt"`
	StartsUnix int64     `dynamodbav:"startsAtUnix"`
	Capacity   int       `dynamodbav:"capacity"`
	Registered int       `dynamodbav:"registered"`
	PriceCOP   int64     `dynamodbav:"priceCop"`
	Status     string    `dynamodbav:"status"`
	Bank       string    `dynamodbav:"bank,omitempty"`
	Account    string    `dynamodbav:"account,omitempty"`
	CreatedAt  time.Time `dynamodbav:"createdAt"`
}

func newEventItem(e event.Event) eventItem {
	return eventItem{
		PK:         eventPK(e.ID),
		SK:         eventSK,
		GSI1PK:     eventsGSI1PK,
		GSI1SK:     eventGSI1SK(e),
		Type:       "event",
		ID:         e.ID,
		Name:       e.Name,
		StartsAt:   e.StartsAt,
		StartsUnix: e.StartsAt.Unix(),
		Capacity:   e.Capacity,
		Registered: e.Registered,
		PriceCOP:   e.PriceCOP,
		Status:     string(e.Status),
		Bank:       e.Bank,
		Account:    e.Account,
		CreatedAt:  e.CreatedAt,
	}
}

func (i eventItem) toDomain() event.Event {
	return event.Event{
		ID:         i.ID,
		Name:       i.Name,
		StartsAt:   i.StartsAt,
		Capacity:   i.Capacity,
		Registered: i.Registered,
		PriceCOP:   i.PriceCOP,
		Status:     event.Status(i.Status),
		Bank:       i.Bank,
		Account:    i.Account,
		CreatedAt:  i.CreatedAt,
	}
}

type attendeeItem struct {
	PK             string     `dynamodbav:"PK"`
	SK             string     `dynamodbav:"SK"`
	Type           string     `dynamodbav:"type"`
	EventID        string     `dynamodbav:"eventId"`
	ID             string     `dynamodbav:"id"`
	FirstName      string     `dynamodbav:"firstName"`
	LastName       string     `dynamodbav:"lastName"`
	Email          string     `dynamodbav:"email"`
	Phone          string     `dynamodbav:"phone,omitempty"`
	Gender         string     `dynamodbav:"gender,omitempty"`
	Scholarship    bool       `dynamodbav:"scholarship"`
	PaymentStatus  string     `dynamodbav:"paymentStatus"`
	Attendance     string     `dynamodbav:"attendance"`
	RegisteredAt   time.Time  `dynamodbav:"registeredAt"`
	CheckedInAt    *time.Time `dynamodbav:"checkedInAt,omitempty"`
	ReminderSentAt *time.Time `dynamodbav:"reminderSentAt,omitempty"`
}

func newAttendeeItem(a event.Attendee) attendeeItem {
	return attendeeItem{
		PK:             eventPK(a.EventID),
		SK:             attendeeSK(a.ID),
		Type:           "attendee",
		EventID:        a.EventID,
		ID:             a.ID,
		FirstName:      a.FirstName,
		LastName:       a.LastName,
		Email:          a.Email,
		Phone:          a.Phone,
		Gender:         a.Gender,
		Scholarship:    a.Scholarship,
		PaymentStatus:  string(a.PaymentStatus),
		Attendance:     string(a.Attendance),
		RegisteredAt:   a.RegisteredAt,
		CheckedInAt:    a.CheckedInAt,
		ReminderSentAt: a.ReminderSentAt,
	}
}

func (i attendeeItem) toDomain() event.Attendee {
	return event.Attendee{
		EventID:        i.EventID,
		ID:             i.ID,
		FirstName:      i.FirstName,
		LastName:       i.LastName,
		Email:          i.Email,
		Phone:          i.Phone,
		Gender:         i.Gender,
		Scholarship:    i.Scholarship,
		PaymentStatus:  event.PaymentStatus(i.PaymentStatus),
		Attendance:     event.Attendance(i.Attendance),
		RegisteredAt:   i.RegisteredAt,
		CheckedInAt:    i.CheckedInAt,
		ReminderSentAt: i.ReminderSentAt,
	}
}

type emailGuardItem struct {
	PK         string `dynamodbav:"PK"`
	SK         string `dynamodbav:"SK"`
	Type       string `dynamodbav:"type"`
	AttendeeID string `dynamodbav:"attendeeId"`
}
