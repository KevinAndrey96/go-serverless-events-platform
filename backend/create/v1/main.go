// Package main implements the Create Event microservice handler
package main

import (
	"context"
	"encoding/json"

	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/config"
	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/models"
	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/response"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Handler processes the API Gateway request to register a new event
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.Body == "" {
		return response.Error(400, "request body cannot be empty")
	}

	var eventItem models.Event
	err := json.Unmarshal([]byte(request.Body), &eventItem)
	if err != nil {
		return response.Error(400, "invalid JSON payload: "+err.Error())
	}

	if eventItem.PK == "" || eventItem.Name == "" {
		return response.Error(422, "missing required fields: pk and name are mandatory")
	}

	if eventItem.SK == "" {
		eventItem.SK = "METADATA#EVENT"
	}

	db := config.GetDynamoDBClient()
	tableName := config.GetTableName()

	item, err := dynamodbattribute.MarshalMap(eventItem)
	if err != nil {
		return response.Error(500, "failed to marshal item for database storage")
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	}

	_, err = db.PutItemWithContext(ctx, input)
	if err != nil {
		return response.Error(500, "database insertion failed: "+err.Error())
	}

	return response.Success(map[string]interface{}{
		"message": "event successfully created",
		"pk":      eventItem.PK,
		"sk":      eventItem.SK,
	})
}

func main() {
	lambda.Start(Handler)
}
