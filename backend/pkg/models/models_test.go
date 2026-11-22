package models_test

import (
	"encoding/json"
	"testing"

	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/models"
)

func TestUpdatePaymentStatusRequest_GetEffectiveStatus(t *testing.T) {
	req1 := models.UpdatePaymentStatusRequest{
		PK:            "123",
		SK:            "METADATA#USER",
		PaymentStatus: "PAID",
		Status:        "PENDING",
	}
	if req1.GetEffectiveStatus() != "PAID" {
		t.Errorf("expected PAID, got %s", req1.GetEffectiveStatus())
	}

	req2 := models.UpdatePaymentStatusRequest{
		PK:     "123",
		SK:     "METADATA#USER",
		Status: "APPROVED",
	}
	if req2.GetEffectiveStatus() != "APPROVED" {
		t.Errorf("expected APPROVED, got %s", req2.GetEffectiveStatus())
	}
}

func TestEvent_JSONSerialization(t *testing.T) {
	bank := "Bancolombia"
	event := models.Event{
		PK:       "event-1",
		SK:       "METADATA#EVENT",
		Name:     "Tech Conference 2026",
		Capacity: "100",
		Price:    "50000",
		Date:     "2026-12-01",
		Hour:     "09:00",
		Status:   "ACTIVE",
		Bank:     &bank,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed models.Event
	err = json.Unmarshal(bytes, &parsed)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.Name != "Tech Conference 2026" {
		t.Errorf("expected 'Tech Conference 2026', got '%s'", parsed.Name)
	}
	if parsed.Bank == nil || *parsed.Bank != "Bancolombia" {
		t.Errorf("expected Bank 'Bancolombia', got %v", parsed.Bank)
	}
}
