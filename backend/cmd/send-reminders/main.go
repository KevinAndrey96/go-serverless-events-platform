// Command send-reminders emails the attendees of the events starting in the next 24 hours.
package main

import (
	"context"
	_ "time/tzdata"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/handler"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/mail"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/platform"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/usecase"
)

func main() {
	deps, err := platform.Bootstrap(context.Background())
	if err != nil {
		platform.Exit(err)
	}
	from, err := platform.Require("MAIL_FROM")
	if err != nil {
		platform.Exit(err)
	}
	location, err := platform.Location()
	if err != nil {
		platform.Exit(err)
	}
	mailer := mail.NewSES(sesv2.NewFromConfig(deps.Config.AWS), from, location)
	lambda.Start(handler.NewSendReminders(usecase.NewSendReminders(deps.Store, mailer, deps.Logger), deps.Logger).Handle)
}
