package mail_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/mail"
)

type fakeSES struct {
	in  *sesv2.SendEmailInput
	err error
}

func (f *fakeSES) SendEmail(_ context.Context, in *sesv2.SendEmailInput, _ ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	f.in = in

	return &sesv2.SendEmailOutput{}, f.err
}

func TestSendReminder(t *testing.T) {
	t.Parallel()
	bogota, err := time.LoadLocation("America/Bogota")
	if err != nil {
		t.Fatal(err)
	}
	e := event.Event{ID: "e1", Name: "Foro <IA>", StartsAt: time.Date(2027, 1, 10, 14, 30, 0, 0, time.UTC)}
	tests := []struct {
		name    string
		payment event.PaymentStatus
		want    string
	}{
		{name: "paid", payment: event.PaymentPaid, want: "Tu pago está confirmado."},
		{name: "pending", payment: event.PaymentPending, want: "pendiente de revisión"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := &fakeSES{}
			a := event.Attendee{EventID: "e1", ID: "a1", FirstName: "Ana", Email: "ana@example.com", PaymentStatus: tt.payment}
			if err := mail.NewSES(client, "eventos@example.com", bogota).SendReminder(context.Background(), e, a); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			html := aws.ToString(client.in.Content.Simple.Body.Html.Data)
			for _, part := range []string{tt.want, "10/01/2027 a las 09:30", "e1.a1", "Foro &lt;IA&gt;", "Hola Ana"} {
				if !strings.Contains(html, part) {
					t.Errorf("body misses %q: %s", part, html)
				}
			}
			if aws.ToString(client.in.FromEmailAddress) != "eventos@example.com" ||
				client.in.Destination.ToAddresses[0] != "ana@example.com" ||
				aws.ToString(client.in.Content.Simple.Subject.Data) != "Recordatorio: Foro <IA>" {
				t.Fatalf("input = %+v", client.in)
			}
		})
	}
}

func TestSendReminderErrors(t *testing.T) {
	t.Parallel()
	rejected := errors.New("message rejected")
	err := mail.NewSES(&fakeSES{err: rejected}, "a@example.com", time.UTC).
		SendReminder(context.Background(), event.Event{}, event.Attendee{Email: "b@example.com"})
	if !errors.Is(err, rejected) {
		t.Fatalf("error = %v, want %v", err, rejected)
	}
	log := mail.NewLog(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := log.SendReminder(context.Background(), event.Event{}, event.Attendee{}); err != nil {
		t.Fatalf("log mailer error = %v", err)
	}
}
