package response_test

import (
	"encoding/json"
	"testing"

	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/models"
	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/response"
)

func TestCORSHeaders(t *testing.T) {
	headers := response.CORSHeaders()
	if headers["Access-Control-Allow-Origin"] != "*" {
		t.Errorf("expected Access-Control-Allow-Origin to be '*', got %s", headers["Access-Control-Allow-Origin"])
	}
	if headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type to be 'application/json', got %s", headers["Content-Type"])
	}
}

func TestJSON(t *testing.T) {
	payload := map[string]string{"foo": "bar"}
	res, err := response.JSON(200, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}

	var parsed map[string]string
	if err := json.Unmarshal([]byte(res.Body), &parsed); err != nil {
		t.Fatalf("failed to parse JSON response body: %v", err)
	}
	if parsed["foo"] != "bar" {
		t.Errorf("expected 'bar', got '%s'", parsed["foo"])
	}
}

func TestError(t *testing.T) {
	res, err := response.Error(404, "item not found")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != 404 {
		t.Errorf("expected 404, got %d", res.StatusCode)
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal([]byte(res.Body), &errResp); err != nil {
		t.Fatalf("failed to parse error body: %v", err)
	}
	if errResp.Error != "API_ERROR" {
		t.Errorf("expected 'API_ERROR', got '%s'", errResp.Error)
	}
	if errResp.Message != "item not found" {
		t.Errorf("expected 'item not found', got '%s'", errResp.Message)
	}
}

func TestSuccess(t *testing.T) {
	res, err := response.Success("ok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
}
