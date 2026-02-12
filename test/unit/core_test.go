package unit

import (
	"context"
	"testing"

	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/service/core"
)

func TestCoreInit(t *testing.T) {
	_ = core.NewCoreService(nil)
}

func genClusterConfig() *model.HCluster {
	return &model.HCluster{
		Name: "test-cluster",
		Services: []model.HService{{
			Name:  "test-node",
			Image: "test-image",
			Build: &model.Build{
				Context:    ".",
				Dockerfile: "./test/config/helios.hcl",
			},
		}},
	}
}

func initCluster(ctx context.Context) core.CoreService {
	coreService := core.NewCoreService(genClusterConfig())
	return coreService
}

func TestInitCluster(t *testing.T) {
	initCluster(t.Context())
}

func TestCreateClusterPasses(t *testing.T) {
	svc := initCluster(t.Context())
	err := svc.CreateCluster(t.Context())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

func TestCreateClusterCreateFails(t *testing.T) {
	svc := initCluster(t.Context())
	err := svc.CreateCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

func TestStartClusterFails(t *testing.T) {
	svc := initCluster(t.Context())
	err := svc.StartCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

func TestStopClusterFails(t *testing.T) {
	svc := initCluster(t.Context())
	err := svc.StopCluster(t.Context())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

func TestTeardownCluster(t *testing.T) {
	svc := initCluster(t.Context())
	err := svc.TeardownCluster(t.Context())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

func TestTeardownClusterRemoveFails(t *testing.T) {
	svc := initCluster(t.Context())
	err := svc.TeardownCluster(t.Context())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}
