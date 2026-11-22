// Package main implements the Event Users (Attendees) microservice handler
package main

import (
	"context"
	"strings"

	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/config"
	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/models"
	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/response"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Handler processes the API Gateway request to query attendees for a specific event
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	eventID := request.QueryStringParameters["eventid"]
	if strings.TrimSpace(eventID) == "" {
		eventID = request.QueryStringParameters["param"]
	}

	if strings.TrimSpace(eventID) == "" {
		return response.Error(400, "missing required query parameter: 'param' or 'eventid'")
	}

	indexValue := eventID
	if !strings.HasPrefix(indexValue, "Event#") {
		indexValue = "Event#" + indexValue
	}

	db := config.GetDynamoDBClient()
	tableName := config.GetTableName()

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		IndexName: aws.String("eventid-index"),
		KeyConditions: map[string]*dynamodb.Condition{
			"eventid": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(indexValue),
					},
				},
			},
		},
	}

	result, err := db.QueryWithContext(ctx, queryInput)
	if err != nil {
		return response.Error(500, "failed to query attendees index: "+err.Error())
	}

	attendees := make([]models.UserAttendee, 0)
	if len(result.Items) > 0 {
		err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &attendees)
		if err != nil {
			return response.Error(500, "failed to unmarshal attendees: "+err.Error())
		}
	}

	return response.JSON(200, attendees)
}

func main() {
	lambda.Start(Handler)
}
