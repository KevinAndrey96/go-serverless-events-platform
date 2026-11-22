// Package main implements the Change Status microservice handler
package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/config"
	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/models"
	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/response"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// Handler processes the API Gateway request to update an item's status
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if strings.TrimSpace(request.Body) == "" {
		return response.Error(400, "request body cannot be empty")
	}

	var req models.UpdateStatusRequest
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return response.Error(400, "invalid JSON payload: "+err.Error())
	}

	if req.PK == "" || req.SK == "" || req.Status == "" {
		return response.Error(422, "missing required fields: pk, sk, and status are mandatory")
	}

	db := config.GetDynamoDBClient()
	tableName := config.GetTableName()

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {
				S: aws.String(req.PK),
			},
			"sk": {
				S: aws.String(req.SK),
			},
		},
		UpdateExpression: aws.String("SET #S = :s"),
		ExpressionAttributeNames: map[string]*string{
			"#S": aws.String("status"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":s": {
				S: aws.String(req.Status),
			},
		},
	}

	_, err = db.UpdateItemWithContext(ctx, input)
	if err != nil {
		return response.Error(500, "failed to update item status: "+err.Error())
	}

	return response.Success(map[string]interface{}{
		"message": "status updated successfully",
		"pk":      req.PK,
		"sk":      req.SK,
		"status":  req.Status,
	})
}

func main() {
	lambda.Start(Handler)
}
