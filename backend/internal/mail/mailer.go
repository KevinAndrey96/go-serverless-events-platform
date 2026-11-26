// Package mail sends event emails through Amazon SES.
package mail

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/event"
)

const reminderBody = `<p>Hola {{.Name}},</p>
<p>Te recordamos que <strong>{{.Event}}</strong> empieza el {{.StartsAt}}.</p>
<p>Presenta este código en la entrada: <code>{{.Ticket}}</code></p>
<p>{{if .Paid}}Tu pago está confirmado.{{else}}Tu pago aún está pendiente de revisión; sin pago confirmado no podrás ingresar.{{end}}</p>`

// API is the subset of the SES v2 client the mailer uses.
type API interface {
	SendEmail(ctx context.Context, in *sesv2.SendEmailInput, opts ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

// SES sends reminders from a verified sender identity.
type SES struct {
	client   API
	from     string
	location *time.Location
	body     *template.Template
}

// NewSES builds a mailer; dates are written in the given location.
func NewSES(client API, from string, location *time.Location) *SES {
	return &SES{
		client:   client,
		from:     from,
		location: location,
		body:     template.Must(template.New("reminder").Parse(reminderBody)),
	}
}

// SendReminder emails one attendee about the event.
func (m *SES) SendReminder(ctx context.Context, e event.Event, a event.Attendee) error {
	var html bytes.Buffer
	err := m.body.Execute(&html, map[string]any{
		"Name":     a.FirstName,
		"Event":    e.Name,
		"StartsAt": e.StartsAt.In(m.location).Format("02/01/2006 a las 15:04"),
		"Ticket":   a.Ticket(),
		"Paid":     a.PaymentStatus == event.PaymentPaid,
	})
	if err != nil {
		return fmt.Errorf("reminder rendering: %w", err)
	}
	_, err = m.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(m.from),
		Destination:      &types.Destination{ToAddresses: []string{a.Email}},
		Content: &types.EmailContent{Simple: &types.Message{
			Subject: &types.Content{Data: aws.String("Recordatorio: " + e.Name), Charset: aws.String("UTF-8")},
			Body:    &types.Body{Html: &types.Content{Data: aws.String(html.String()), Charset: aws.String("UTF-8")}},
		}},
	})
	if err != nil {
		return fmt.Errorf("ses send: %w", err)
	}

	return nil
}
