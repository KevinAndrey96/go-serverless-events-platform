// Package main implements the Show Events microservice handler
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

// Handler processes the API Gateway request to query events by secondary index
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	paramValue := request.QueryStringParameters["param"]
	if strings.TrimSpace(paramValue) == "" {
		return response.Error(400, "missing required query parameter: 'param' (e.g. 'EVENT')")
	}

	skValue := paramValue
	if !strings.HasPrefix(skValue, "METADATA#") {
		skValue = "METADATA#" + skValue
	}

	db := config.GetDynamoDBClient()
	tableName := config.GetTableName()

	input := &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		IndexName: aws.String("sk-index"),
		KeyConditions: map[string]*dynamodb.Condition{
			"sk": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(skValue),
					},
				},
			},
		},
	}

	result, err := db.QueryWithContext(ctx, input)
	if err != nil {
		return response.Error(500, "failed to query events table: "+err.Error())
	}

	eventsList := make([]models.Event, 0)
	if len(result.Items) > 0 {
		err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &eventsList)
		if err != nil {
			return response.Error(500, "failed to unmarshal database results: "+err.Error())
		}
	}

	return response.JSON(200, eventsList)
}

func main() {
	lambda.Start(Handler)
}
