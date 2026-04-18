package unit

import (
	"context"
	"fmt"
	"testing"

	"github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/model/node"
	"github.com/KyloRilo/helios/pkg/service/core"
)

type CompStub struct{ compute.ComputeController }

func (c CompStub) CreateNode(_ context.Context, _ *node.Node) (*node.CreateNodeResp, error) {
	return nil, nil
}

func (c CompStub) StartNode(_ context.Context, _ *node.Node) (*node.StartNodeResp, error) {
	return nil, nil
}

func (c CompStub) StopNode(_ context.Context, _ *node.Node) (*node.StopNodeResp, error) {
	return nil, nil
}

func (c CompStub) RemoveNode(_ context.Context, _ *node.Node) (*node.RmNodeResp, error) {
	return nil, nil
}

func genClusterConfig() *model.HCluster {
	return &model.HCluster{
		Name: "test-cluster",
		Services: []model.HService{{
			Name: "test-node",
			Build: &model.Build{
				Context:    ".",
				Dockerfile: "./test/config/helios.hcl",
			},
		}},
	}
}

func initCluster(ctx context.Context, stub compute.ComputeController) core.CoreService {
	return core.NewCoreService(ctx, core.CoreArgs{
		Conf: genClusterConfig(),
		ScalerArgs: compute.ControllerArgs{
			Stub: &stub,
		},
	})
}

func TestInitCluster(t *testing.T) {
	initCluster(t.Context(), CompStub{})
}

func TestCreateClusterPasses(t *testing.T) {
	svc := initCluster(t.Context(), CompStub{})
	err := svc.CreateCluster(t.Context())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

type CreateFailed struct{ CompStub }

func (c CreateFailed) CreateNode(_ context.Context, _ *node.Node) (*node.CreateNodeResp, error) {
	return nil, fmt.Errorf("failed to create container")
}

func TestCreateClusterCreateFails(t *testing.T) {
	svc := initCluster(t.Context(), CreateFailed{})
	err := svc.CreateCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

type StartFailed struct{ CompStub }

func (c StartFailed) StartNode(_ context.Context, _ *node.Node) (*node.StartNodeResp, error) {
	return nil, fmt.Errorf("failed to start container")
}

func TestStartClusterFails(t *testing.T) {
	svc := initCluster(t.Context(), StartFailed{})
	err := svc.StartCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

type StopFailed struct{ CompStub }

func (c StopFailed) StopNode(_ context.Context, _ *node.Node) (*node.StopNodeResp, error) {
	return nil, fmt.Errorf("failed to stop container")
}

func TestStopClusterFails(t *testing.T) {
	svc := initCluster(t.Context(), StopFailed{})
	err := svc.StopCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

func TestTeardownCluster(t *testing.T) {
	svc := initCluster(t.Context(), CompStub{})
	err := svc.TeardownCluster(t.Context())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

type RemoveFailed struct{ CompStub }

func (c RemoveFailed) RemoveNode(ctx context.Context, _ *node.Node) (*node.RmNodeResp, error) {
	return nil, fmt.Errorf("failed to remove container")
}

func TestTeardownClusterRemoveFails(t *testing.T) {
	svc := initCluster(t.Context(), RemoveFailed{})
	err := svc.TeardownCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}
