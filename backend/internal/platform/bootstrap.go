package platform

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/KevinAndrey96/go-serverless-events-platform/backend/internal/dynamo"
)

// Deps are the dependencies shared by the functions.
type Deps struct {
	Config Config
	Client *dynamodb.Client
	Store  *dynamo.Store
	Logger *slog.Logger
}

// Bootstrap loads the configuration and builds the store.
func Bootstrap(ctx context.Context) (Deps, error) {
	cfg, err := Load(ctx)
	if err != nil {
		return Deps{}, err
	}
	client := dynamodb.NewFromConfig(cfg.AWS, func(o *dynamodb.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	return Deps{Config: cfg, Client: client, Store: dynamo.NewStore(client, cfg.Table), Logger: Logger()}, nil
}

// Exit logs a startup error and stops the process.
func Exit(err error) {
	Logger().Error("startup", "error", err)
	os.Exit(1)
}
