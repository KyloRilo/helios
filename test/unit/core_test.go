package unit

import (
	"context"
	"fmt"
	"testing"

	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/service/core"
)

type stub interface {
	model.ComputeController
}

type CompStub struct{}

func (c CompStub) Authenticate(_ context.Context) (interface{}, error) {
	return nil, nil
}

func (c CompStub) CreateNode(_ context.Context, _ *model.Node) (interface{}, error) {
	return nil, nil
}

func (c CompStub) StartNode(_ context.Context, _ *model.Node) (interface{}, error) {
	return nil, nil
}

func (c CompStub) StopNode(_ context.Context, _ *model.Node) (interface{}, error) {
	return nil, nil
}

func (c CompStub) RemoveNode(_ context.Context, _ *model.Node) (interface{}, error) {
	return nil, nil
}

func TestCoreInit(t *testing.T) {
	_ = core.NewCoreService()
}

func genClusterConfig() *model.HCluster {
	return &model.HCluster{
		Name: "test-cluster",
		Services: []model.HService{{
			Image: "test-image",
			Build: &model.Build{
				Context:    ".",
				Dockerfile: "./test/config/helios.hcl",
			},
		}},
	}
}

func initCluster(ctx context.Context, ctrlStub stub) (core.CoreService, error) {
	coreService := core.NewCoreService()
	err := coreService.InitCluster(ctx, genClusterConfig(), ctrlStub)
	return coreService, err
}
func TestInitCluster(t *testing.T) {
	if _, err := initCluster(t.Context(), nil); err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

type CreateSuccess struct{ CompStub }

func (c CreateSuccess) CreateNode(ctx context.Context, node *model.Node) (interface{}, error) {
	return nil, nil
}

func TestCreateClusterPasses(t *testing.T) {
	svc, err := initCluster(t.Context(), &CompStub{})
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	err = svc.CreateCluster(t.Context(), genClusterConfig())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

type CreateFailed struct{ CompStub }

func (c CreateFailed) CreateNode(_ context.Context, _ *model.Node) (interface{}, error) {
	return nil, fmt.Errorf("failed to create container")
}

func TestCreateClusterCreateFails(t *testing.T) {
	svc, err := initCluster(t.Context(), CreateFailed{})
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	err = svc.CreateCluster(t.Context(), nil)
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

type StartFailed struct{ CompStub }

func (c StartFailed) StartNode(_ context.Context, _ *model.Node) (interface{}, error) {
	return nil, fmt.Errorf("failed to start container")
}

func TestStartClusterFails(t *testing.T) {
	svc, err := initCluster(t.Context(), StartFailed{})
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	err = svc.StartCluster(t.Context(), nil)
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

func TestTeardownCluster(t *testing.T) {
	svc, err := initCluster(t.Context(), &CompStub{})
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	err = svc.TeardownCluster(t.Context(), genClusterConfig())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

type StopFailed struct{ CompStub }

func (c StopFailed) StopNode(ctx context.Context, _ *model.Node) (interface{}, error) {
	return nil, fmt.Errorf("failed to stop container")
}

func TestDestroyServiceStopFails(t *testing.T) {
	svc, err := initCluster(t.Context(), &StopFailed{})
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	err = svc.StopCluster(t.Context(), genClusterConfig())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

type RemoveFailed struct{ CompStub }

func (c RemoveFailed) RemoveNode(ctx context.Context, _ *model.Node) (interface{}, error) {
	return nil, fmt.Errorf("failed to remove container")
}

func TestDestroyServiceRemoveFails(t *testing.T) {
	svc, err := initCluster(t.Context(), &RemoveFailed{})
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	err = svc.TeardownCluster(t.Context(), genClusterConfig())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}
