package config_test

import (
	"os"
	"testing"

	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/config"
)

func TestGetRegion_Default(t *testing.T) {
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")
	region := config.GetRegion()
	if region != "us-east-1" {
		t.Errorf("expected us-east-1, got %s", region)
	}
}

func TestGetRegion_Override(t *testing.T) {
	os.Setenv("AWS_REGION", "us-west-2")
	defer os.Unsetenv("AWS_REGION")

	region := config.GetRegion()
	if region != "us-west-2" {
		t.Errorf("expected us-west-2, got %s", region)
	}
}

func TestGetTableName_Default(t *testing.T) {
	os.Unsetenv("EVENTS_TABLE_NAME")
	os.Unsetenv("TABLE_NAME")
	tableName := config.GetTableName()
	if tableName != "events-table" {
		t.Errorf("expected events-table, got %s", tableName)
	}
}

func TestGetTableName_Override(t *testing.T) {
	os.Setenv("EVENTS_TABLE_NAME", "custom-events-table")
	defer os.Unsetenv("EVENTS_TABLE_NAME")

	tableName := config.GetTableName()
	if tableName != "custom-events-table" {
		t.Errorf("expected custom-events-table, got %s", tableName)
	}
}
