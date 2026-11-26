// Package platform reads the runtime configuration shared by every Lambda.
package platform

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// DefaultTimeZone is the zone used to write dates in emails.
const DefaultTimeZone = "America/Bogota"

// ErrMissingEnv reports a required environment variable that is not set.
var ErrMissingEnv = errors.New("missing environment variable")

// Config is the environment of a function.
type Config struct {
	AWS      aws.Config
	Table    string
	Endpoint string
}

// Load reads TABLE_NAME and the optional DYNAMODB_ENDPOINT used with DynamoDB Local.
func Load(ctx context.Context) (Config, error) {
	table, err := Require("TABLE_NAME")
	if err != nil {
		return Config{}, err
	}
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRetryMaxAttempts(5))
	if err != nil {
		return Config{}, fmt.Errorf("aws config: %w", err)
	}

	return Config{AWS: awsConfig, Table: table, Endpoint: os.Getenv("DYNAMODB_ENDPOINT")}, nil
}

// Require returns a non-empty environment variable.
func Require(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%w: %s", ErrMissingEnv, name)
	}

	return value, nil
}

// Location returns TIME_ZONE, or DefaultTimeZone when it is not set.
func Location() (*time.Location, error) {
	name := os.Getenv("TIME_ZONE")
	if name == "" {
		name = DefaultTimeZone
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("time zone: %w", err)
	}

	return location, nil
}

// Logger returns the JSON logger every function writes to CloudWatch.
func Logger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
