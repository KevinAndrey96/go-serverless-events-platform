// Command list-events-admin serves GET /admin/events.
package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/handler"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/platform"
	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/usecase"
)

func main() {
	deps, err := platform.Bootstrap(context.Background())
	if err != nil {
		platform.Exit(err)
	}
	lambda.Start(handler.NewListEvents(usecase.NewListEvents(deps.Store), deps.Logger).HandleAdmin)
}
