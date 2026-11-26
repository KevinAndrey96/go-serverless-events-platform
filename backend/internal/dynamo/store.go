package dynamo

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

const (
	maxTransactionAttempts = 5
	conditionFailed        = "ConditionalCheckFailed"
	transactionConflict    = "TransactionConflict"
)

var errUnexpectedRejection = errors.New("registration rejected")

// API is the subset of the DynamoDB client the store uses.
type API interface {
	PutItem(ctx context.Context, in *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, in *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, in *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	Query(ctx context.Context, in *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	TransactWriteItems(ctx context.Context, in *dynamodb.TransactWriteItemsInput, opts ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
}

// Store implements every persistence port of the use cases.
type Store struct {
	client API
	table  string
}

// NewStore builds a store over one table.
func NewStore(client API, table string) *Store {
	return &Store{client: client, table: table}
}

func key(pk, sk string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: pk},
		"SK": &types.AttributeValueMemberS{Value: sk},
	}
}

func str(value string) *types.AttributeValueMemberS {
	return &types.AttributeValueMemberS{Value: value}
}

func isConditionFailure(err error) bool {
	var failed *types.ConditionalCheckFailedException

	return errors.As(err, &failed)
}

// CreateEvent stores a new event; identifiers never collide with an existing one.
func (s *Store) CreateEvent(ctx context.Context, e event.Event) error {
	item, err := attributevalue.MarshalMap(newEventItem(e))
	if err != nil {
		return fmt.Errorf("event encoding: %w", err)
	}
	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	if err != nil {
		return fmt.Errorf("event put: %w", err)
	}

	return nil
}

// GetEvent fails with event.ErrEventNotFound when there is no such event.
func (s *Store) GetEvent(ctx context.Context, eventID string) (event.Event, error) {
	out, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(s.table),
		Key:            key(eventPK(eventID), eventSK),
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return event.Event{}, fmt.Errorf("event get: %w", err)
	}
	if out.Item == nil {
		return event.Event{}, event.ErrEventNotFound
	}

	return decodeEvent(out.Item)
}

// ListEvents returns every event ordered by start time.
func (s *Store) ListEvents(ctx context.Context) ([]event.Event, error) {
	return s.queryEvents(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(s.table),
		IndexName:                 aws.String(GSI1),
		KeyConditionExpression:    aws.String("GSI1PK = :events"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":events": str(eventsGSI1PK)},
	})
}

// ListEventsStartingBetween returns the events whose start falls in [from, to].
func (s *Store) ListEventsStartingBetween(ctx context.Context, from, to time.Time) ([]event.Event, error) {
	return s.queryEvents(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.table),
		IndexName:              aws.String(GSI1),
		KeyConditionExpression: aws.String("GSI1PK = :events AND GSI1SK BETWEEN :from AND :to"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":events": str(eventsGSI1PK),
			":from":   str(startsKey(from)),
			":to":     str(startsKey(to) + "~"),
		},
	})
}

func (s *Store) queryEvents(ctx context.Context, in *dynamodb.QueryInput) ([]event.Event, error) {
	items, err := s.queryAll(ctx, in)
	if err != nil {
		return nil, err
	}
	found := make([]event.Event, 0, len(items))
	for _, item := range items {
		e, err := decodeEvent(item)
		if err != nil {
			return nil, err
		}
		found = append(found, e)
	}

	return found, nil
}

// UpdateEventStatus fails with event.ErrEventNotFound when there is no such event.
func (s *Store) UpdateEventStatus(ctx context.Context, eventID string, status event.Status) (event.Event, error) {
	out, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(s.table),
		Key:                       key(eventPK(eventID), eventSK),
		UpdateExpression:          aws.String("SET #status = :status"),
		ConditionExpression:       aws.String("attribute_exists(PK)"),
		ExpressionAttributeNames:  map[string]string{"#status": "status"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":status": str(string(status))},
		ReturnValues:              types.ReturnValueAllNew,
	})
	if isConditionFailure(err) {
		return event.Event{}, event.ErrEventNotFound
	}
	if err != nil {
		return event.Event{}, fmt.Errorf("event status update: %w", err)
	}

	return decodeEvent(out.Attributes)
}

// RegisterAttendee takes a seat, stores the attendee and reserves the email in one transaction.
func (s *Store) RegisterAttendee(ctx context.Context, a event.Attendee, now time.Time) error {
	attendee, err := attributevalue.MarshalMap(newAttendeeItem(a))
	if err != nil {
		return fmt.Errorf("attendee encoding: %w", err)
	}
	guard, err := attributevalue.MarshalMap(emailGuardItem{
		PK: eventPK(a.EventID), SK: emailSK(a.Email), Type: "email", AttendeeID: a.ID,
	})
	if err != nil {
		return fmt.Errorf("email guard encoding: %w", err)
	}
	actions := []types.TransactWriteItem{
		{Update: &types.Update{
			TableName:        aws.String(s.table),
			Key:              key(eventPK(a.EventID), eventSK),
			UpdateExpression: aws.String("SET #registered = #registered + :one"),
			ConditionExpression: aws.String(
				"attribute_exists(PK) AND #status = :enabled AND #registered < #capacity AND startsAtUnix > :now"),
			ExpressionAttributeNames: map[string]string{
				"#status": "status", "#registered": "registered", "#capacity": "capacity",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":one":     &types.AttributeValueMemberN{Value: "1"},
				":enabled": str(string(event.StatusEnabled)),
				":now":     &types.AttributeValueMemberN{Value: strconv.FormatInt(now.Unix(), 10)},
			},
		}},
		{Put: &types.Put{
			TableName: aws.String(s.table), Item: attendee, ConditionExpression: aws.String("attribute_not_exists(PK)"),
		}},
		{Put: &types.Put{
			TableName: aws.String(s.table), Item: guard, ConditionExpression: aws.String("attribute_not_exists(PK)"),
		}},
	}
	reasons, err := s.transact(ctx, actions)
	if err != nil {
		return err
	}
	switch {
	case reasons == nil:
		return nil
	case reasons[2] == conditionFailed:
		return event.ErrAlreadyRegistered
	case reasons[0] == conditionFailed:
		return s.whyClosed(ctx, a.EventID, now)
	default:
		return fmt.Errorf("%w: %v", errUnexpectedRejection, reasons)
	}
}

func (s *Store) whyClosed(ctx context.Context, eventID string, now time.Time) error {
	e, err := s.GetEvent(ctx, eventID)
	if err != nil {
		return err
	}
	if e.Status != event.StatusEnabled || !e.StartsAt.After(now) {
		return event.ErrEventClosed
	}

	return event.ErrEventFull
}

// transact runs a transaction, retrying conflicts with jittered exponential backoff. It returns the
// cancellation reason codes when a condition rejected the transaction.
func (s *Store) transact(ctx context.Context, actions []types.TransactWriteItem) ([]string, error) {
	for attempt := 1; ; attempt++ {
		_, err := s.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: actions})
		if err == nil {
			return nil, nil
		}
		var cancelled *types.TransactionCanceledException
		if !errors.As(err, &cancelled) {
			return nil, fmt.Errorf("transaction: %w", err)
		}
		codes := make([]string, len(cancelled.CancellationReasons))
		for i, reason := range cancelled.CancellationReasons {
			codes[i] = aws.ToString(reason.Code)
		}
		if slices.Contains(codes, conditionFailed) {
			return codes, nil
		}
		if !slices.Contains(codes, transactionConflict) || attempt == maxTransactionAttempts {
			return nil, fmt.Errorf("transaction: %w", err)
		}
		backoff := time.Duration(rand.Int64N(int64(20*time.Millisecond) << attempt))
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("transaction retry: %w", ctx.Err())
		case <-time.After(backoff):
		}
	}
}

// ListAttendees returns everyone registered for an event, in no particular order.
func (s *Store) ListAttendees(ctx context.Context, eventID string) ([]event.Attendee, error) {
	items, err := s.queryAll(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.table),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :attendees)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        str(eventPK(eventID)),
			":attendees": str(attendeePrefix),
		},
	})
	if err != nil {
		return nil, err
	}
	found := make([]event.Attendee, 0, len(items))
	for _, item := range items {
		a, err := decodeAttendee(item)
		if err != nil {
			return nil, err
		}
		found = append(found, a)
	}

	return found, nil
}

// UpdatePaymentStatus fails with event.ErrAttendeeNotFound when there is no such attendee.
func (s *Store) UpdatePaymentStatus(
	ctx context.Context, eventID, attendeeID string, status event.PaymentStatus,
) (event.Attendee, error) {
	out, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(s.table),
		Key:                       key(eventPK(eventID), attendeeSK(attendeeID)),
		UpdateExpression:          aws.String("SET paymentStatus = :status"),
		ConditionExpression:       aws.String("attribute_exists(PK)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":status": str(string(status))},
		ReturnValues:              types.ReturnValueAllNew,
	})
	if isConditionFailure(err) {
		return event.Attendee{}, event.ErrAttendeeNotFound
	}
	if err != nil {
		return event.Attendee{}, fmt.Errorf("payment status update: %w", err)
	}

	return decodeAttendee(out.Attributes)
}

// CheckIn admits a paid attendee once.
func (s *Store) CheckIn(ctx context.Context, eventID, attendeeID string, at time.Time) (event.Attendee, error) {
	out, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:        aws.String(s.table),
		Key:              key(eventPK(eventID), attendeeSK(attendeeID)),
		UpdateExpression: aws.String("SET attendance = :attended, checkedInAt = :at"),
		ConditionExpression: aws.String(
			"attribute_exists(PK) AND paymentStatus = :paid AND attendance = :registered"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":attended":   str(string(event.AttendanceAttended)),
			":registered": str(string(event.AttendanceRegistered)),
			":paid":       str(string(event.PaymentPaid)),
			":at":         str(at.UTC().Format(time.RFC3339Nano)),
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if isConditionFailure(err) {
		return event.Attendee{}, s.whyRefused(ctx, eventID, attendeeID)
	}
	if err != nil {
		return event.Attendee{}, fmt.Errorf("check-in update: %w", err)
	}

	return decodeAttendee(out.Attributes)
}

func (s *Store) whyRefused(ctx context.Context, eventID, attendeeID string) error {
	out, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(s.table),
		Key:            key(eventPK(eventID), attendeeSK(attendeeID)),
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("attendee get: %w", err)
	}
	if out.Item == nil {
		return event.ErrInvalidTicket
	}
	a, err := decodeAttendee(out.Item)
	if err != nil {
		return err
	}
	if a.Attendance == event.AttendanceAttended {
		return event.ErrAlreadyCheckedIn
	}

	return event.ErrPaymentPending
}

// ClaimReminder marks the reminder as sent; it returns false when another run already did.
func (s *Store) ClaimReminder(ctx context.Context, eventID, attendeeID string, at time.Time) (bool, error) {
	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(s.table),
		Key:                       key(eventPK(eventID), attendeeSK(attendeeID)),
		UpdateExpression:          aws.String("SET reminderSentAt = :at"),
		ConditionExpression:       aws.String("attribute_exists(PK) AND attribute_not_exists(reminderSentAt)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":at": str(at.UTC().Format(time.RFC3339Nano))},
	})
	if isConditionFailure(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reminder claim: %w", err)
	}

	return true, nil
}

// ReleaseReminder undoes a claim so the next run retries the email.
func (s *Store) ReleaseReminder(ctx context.Context, eventID, attendeeID string) error {
	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:           aws.String(s.table),
		Key:                 key(eventPK(eventID), attendeeSK(attendeeID)),
		UpdateExpression:    aws.String("REMOVE reminderSentAt"),
		ConditionExpression: aws.String("attribute_exists(PK)"),
	})
	if err != nil {
		return fmt.Errorf("reminder release: %w", err)
	}

	return nil
}

func (s *Store) queryAll(ctx context.Context, in *dynamodb.QueryInput) ([]map[string]types.AttributeValue, error) {
	var items []map[string]types.AttributeValue
	pages := dynamodb.NewQueryPaginator(s.client, in)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("query: %w", err)
		}
		items = append(items, page.Items...)
	}

	return items, nil
}

func decodeEvent(item map[string]types.AttributeValue) (event.Event, error) {
	var decoded eventItem
	if err := attributevalue.UnmarshalMap(item, &decoded); err != nil {
		return event.Event{}, fmt.Errorf("event decoding: %w", err)
	}

	return decoded.toDomain(), nil
}

func decodeAttendee(item map[string]types.AttributeValue) (event.Attendee, error) {
	var decoded attendeeItem
	if err := attributevalue.UnmarshalMap(item, &decoded); err != nil {
		return event.Attendee{}, fmt.Errorf("attendee decoding: %w", err)
	}

	return decoded.toDomain(), nil
}
