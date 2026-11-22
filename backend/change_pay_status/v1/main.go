// Package main implements the Change Payment Status microservice handler
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

// Handler processes the API Gateway request to update attendee payment status
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if strings.TrimSpace(request.Body) == "" {
		return response.Error(400, "request body cannot be empty")
	}

	var req models.UpdatePaymentStatusRequest
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return response.Error(400, "invalid JSON payload: "+err.Error())
	}

	effectiveStatus := req.GetEffectiveStatus()
	if req.PK == "" || req.SK == "" || effectiveStatus == "" {
		return response.Error(422, "missing required fields: pk, sk, and payment status are mandatory")
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
			"#S": aws.String("paymentstatus"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":s": {
				S: aws.String(effectiveStatus),
			},
		},
	}

	_, err = db.UpdateItemWithContext(ctx, input)
	if err != nil {
		return response.Error(500, "failed to update payment status: "+err.Error())
	}

	return response.Success(map[string]interface{}{
		"message":        "payment status updated successfully",
		"pk":             req.PK,
		"sk":             req.SK,
		"payment_status": effectiveStatus,
	})
}

func main() {
	lambda.Start(Handler)
}
