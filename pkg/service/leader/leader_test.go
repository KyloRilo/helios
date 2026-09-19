package leader

import (
	"testing"

	"github.com/KyloRilo/helios/pkg/model"
)

func TestParseLeaderConfig(t *testing.T) {
	conf, err := model.ParseLeaderConfig(`
		name = "test-leader"
		host = "127.0.0.1"
		port = 6330
	`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if conf.Name != "test-leader" {
		t.Errorf("expected name 'test-leader' but got '%s'", conf.Name)
	}

	if conf.Host != "127.0.0.1" {
		t.Errorf("expected host '127.0.0.1' but got '%s'", conf.Host)
	}

	if conf.Port != 6330 {
		t.Errorf("expected port 6330 but got %d", conf.Port)
	}
}

func TestParseLeaderConfigMissingHost(t *testing.T) {
	_, err := model.ParseLeaderConfig(`
		name = "test-leader"
		port = 6330
	`)
	if err == nil {
		t.Error("expected error for missing host but got nil")
	}
}

func TestParseLeaderConfigMissingPort(t *testing.T) {
	_, err := model.ParseLeaderConfig(`
		name = "test-leader"
		host = "127.0.0.1"
	`)
	if err == nil {
		t.Error("expected error for missing port but got nil")
	}
}
