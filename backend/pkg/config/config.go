package config

import (
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	dynamoClient *dynamodb.DynamoDB
	clientOnce   sync.Once
)

// GetRegion returns the configured AWS Region, defaulting to us-east-1
func GetRegion() string {
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		return region
	}
	return "us-east-1"
}

// GetTableName returns the DynamoDB single-table name from env, defaulting to events-table
func GetTableName() string {
	if table := os.Getenv("EVENTS_TABLE_NAME"); table != "" {
		return table
	}
	if table := os.Getenv("TABLE_NAME"); table != "" {
		return table
	}
	return "events-table"
}

// GetDynamoDBClient returns a singleton instance of the DynamoDB service client
func GetDynamoDBClient() *dynamodb.DynamoDB {
	clientOnce.Do(func() {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(GetRegion()),
		}))
		dynamoClient = dynamodb.New(sess)
	})
	return dynamoClient
}
