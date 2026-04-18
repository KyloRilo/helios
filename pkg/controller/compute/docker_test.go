package compute

import (
	"testing"
)

func TestParseCommandDefaultsToSleep(t *testing.T) {
	cmd := parseCommand("   ")
	if len(cmd) != 2 || cmd[0] != "sleep" || cmd[1] != "infinity" {
		t.Fatalf("unexpected default command: %v", cmd)
	}
}

func TestParseCommandSplitsCommand(t *testing.T) {
	cmd := parseCommand("echo hello")
	if len(cmd) != 2 || cmd[0] != "echo" || cmd[1] != "hello" {
		t.Fatalf("unexpected command: %v", cmd)
	}
}

func TestParsePorts(t *testing.T) {
	exposed, bindings, err := parsePorts([]string{"8080:80", "127.0.0.1:8443:443/tcp"})
	if err != nil {
		t.Fatalf("expected no parse error, got %v", err)
	}
	if len(exposed) != 2 {
		t.Fatalf("expected 2 exposed ports, got %d", len(exposed))
	}
	if len(bindings) != 2 {
		t.Fatalf("expected 2 port bindings, got %d", len(bindings))
	}
}

func TestParsePortsRejectsInvalidSpec(t *testing.T) {
	_, _, err := parsePorts([]string{"abc"})
	if err == nil {
		t.Fatal("expected parse error for invalid port spec")
	}
}
