package dynamo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

// scriptedAPI answers TransactWriteItems with the scripted errors, in order.
type scriptedAPI struct {
	API

	errs  []error
	calls int
}

func (s *scriptedAPI) TransactWriteItems(
	context.Context, *dynamodb.TransactWriteItemsInput, ...func(*dynamodb.Options),
) (*dynamodb.TransactWriteItemsOutput, error) {
	s.calls++
	if len(s.errs) == 0 {
		return &dynamodb.TransactWriteItemsOutput{}, nil
	}
	err := s.errs[0]
	s.errs = s.errs[1:]

	return nil, err
}

func cancelled(codes ...string) error {
	reasons := make([]types.CancellationReason, len(codes))
	for i, code := range codes {
		reasons[i] = types.CancellationReason{Code: aws.String(code)}
	}

	return &types.TransactionCanceledException{CancellationReasons: reasons}
}

func TestRegisterAttendeeTransactions(t *testing.T) {
	t.Parallel()
	conflict := cancelled("None", transactionConflict, "None")
	throttled := errors.New("throttled")
	tests := []struct {
		name  string
		errs  []error
		calls int
		want  error
	}{
		{name: "retries conflicts", errs: []error{conflict, conflict}, calls: 3},
		{name: "gives up after the last attempt", errs: []error{conflict, conflict, conflict, conflict, conflict}, calls: 5, want: conflict},
		{name: "does not retry other errors", errs: []error{throttled}, calls: 1, want: throttled},
		{name: "duplicate email", errs: []error{cancelled("None", "None", conditionFailed)}, calls: 1, want: event.ErrAlreadyRegistered},
		{name: "unexpected rejection", errs: []error{cancelled("None", conditionFailed, "None")}, calls: 1, want: errUnexpectedRejection},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := &scriptedAPI{errs: tt.errs}
			a := event.Attendee{EventID: "e1", ID: "a1", Email: "ana@example.com"}
			err := NewStore(api, "events").RegisterAttendee(context.Background(), a, time.Now())
			if api.calls != tt.calls {
				t.Fatalf("calls = %d, want %d", api.calls, tt.calls)
			}
			switch {
			case tt.want == nil && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want != nil && !errors.Is(err, tt.want):
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestTransactStopsWhenCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	api := &scriptedAPI{errs: []error{cancelled(transactionConflict)}}
	if _, err := NewStore(api, "events").transact(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
