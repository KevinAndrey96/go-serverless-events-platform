package response

import (
	"encoding/json"

	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/models"
	"github.com/aws/aws-lambda-go/events"
)

// CORSHeaders returns standard Cross-Origin Resource Sharing headers for API Gateway
func CORSHeaders() map[string]string {
	return map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PATCH, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Amz-Date, X-Api-Key, X-Amz-Security-Token",
	}
}

// JSON builds an APIGatewayProxyResponse with a serialized JSON payload
func JSON(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	bytes, err := json.Marshal(body)
	if err != nil {
		return Error(500, "failed to marshal response payload")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    CORSHeaders(),
		Body:       string(bytes),
	}, nil
}

// Success returns a 200 OK response with a generic success message or payload
func Success(body interface{}) (events.APIGatewayProxyResponse, error) {
	return JSON(200, body)
}

// Error builds an APIGatewayProxyResponse containing a structured error response
func Error(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	errResp := models.ErrorResponse{
		Error:   "API_ERROR",
		Message: message,
	}
	bytes, _ := json.Marshal(errResp)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    CORSHeaders(),
		Body:       string(bytes),
	}, nil
}
