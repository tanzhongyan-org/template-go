package health

import (
	"testing"
	"time"
)

func TestGetStatus(t *testing.T) {
	appName := "test-app"
	env := "test"

	before := time.Now().UTC()
	status := GetStatus(appName, env)
	after := time.Now().UTC()

	if status.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", status.Status)
	}
	if status.AppName != appName {
		t.Errorf("expected appName '%s', got '%s'", appName, status.AppName)
	}
	if status.Environment != env {
		t.Errorf("expected environment '%s', got '%s'", env, status.Environment)
	}
	if status.Timestamp.Before(before) || status.Timestamp.After(after) {
		t.Errorf("timestamp out of expected range")
	}
}
