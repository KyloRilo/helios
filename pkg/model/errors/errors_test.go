package errors

import (
	"strings"
	"testing"
)

func TestInvalidNodeStatusError(t *testing.T) {
	err := InvalidNodeStatus{
		NodeName:   "web-server",
		NodeStatus: "READY",
		Action:     "Start",
		Expected:   []string{"CREATED", "DOWN"},
	}

	msg := err.Error()

	if !strings.Contains(msg, "web-server") {
		t.Errorf("expected error to contain node name, got: %s", msg)
	}
	if !strings.Contains(msg, "READY") {
		t.Errorf("expected error to contain node status, got: %s", msg)
	}
	if !strings.Contains(msg, "Start") {
		t.Errorf("expected error to contain action, got: %s", msg)
	}
	if !strings.Contains(msg, "CREATED") || !strings.Contains(msg, "DOWN") {
		t.Errorf("expected error to contain expected statuses, got: %s", msg)
	}
}

func TestBaseErrorFormat(t *testing.T) {
	be := BaseError{}
	result := be.error("hello %s %d", "world", 42)
	if result != "hello world 42" {
		t.Errorf("expected 'hello world 42' but got '%s'", result)
	}
}
