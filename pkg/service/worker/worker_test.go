package worker

import (
	"testing"

	"github.com/KyloRilo/helios/pkg/model"
)

func TestParseWorkerConfig(t *testing.T) {
	conf, err := model.ParseWorkerConfig(``)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if conf == nil {
		t.Error("expected config but got nil")
	}
}
