package unit

import (
	"context"
	"fmt"
	"testing"

	compCtrl "github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/model/compute"
	"github.com/KyloRilo/helios/pkg/service/core"
	"github.com/google/uuid"
)

type CompStub struct{ compCtrl.CtrlShim }

func (c CompStub) createNode(_ context.Context, _ *compute.Node) (string, error) {
	return "", nil
}

func (c CompStub) startNode(_ context.Context, _ *compute.Node) error {
	return nil
}

func (c CompStub) stopNode(_ context.Context, _ *compute.Node) error {
	return nil
}

func (c CompStub) removeNode(_ context.Context, _ *compute.Node) error {
	return nil
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

func setNodes(svc core.CoreService) {
	nodes := svc.GenNodes(svc.GetConfig().Services)
	for _, n := range nodes {
		n.SetId(uuid.New().String())
	}

	svc.SetNodes(nodes)
}

func initCluster(ctx context.Context, stub compCtrl.CtrlShim) core.CoreService {
	svc := core.NewCoreService(ctx, core.CoreArgs{
		Conf: genClusterConfig(),
		ScalerArgs: compCtrl.ControllerArgs{
			Stub: &stub,
		},
	})

	setNodes(svc)
	return svc
}

func TestInitCluster(t *testing.T) {
	initCluster(t.Context(), CompStub{})
}

type CreatePasses CompStub

func (c CreatePasses) createNode(_ context.Context, n *compute.Node) (string, error) {
	return "some-id", nil
}

func TestCreateClusterPasses(t *testing.T) {
	svc := initCluster(t.Context(), CreatePasses{})

	fmt.Println(svc.GetNodes())
	fmt.Println("iterr nodes")
	for n := range svc.GetNodes() {
		fmt.Println(n)
	}
	fmt.Println("iterr done")
	err := svc.CreateCluster(t.Context())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

type CreateFailed CompStub

func (c CreateFailed) createNode(_ context.Context, _ *compute.Node) (string, error) {
	return "", fmt.Errorf("failed to create container")
}

func TestCreateClusterCreateFails(t *testing.T) {
	svc := initCluster(t.Context(), CreateFailed{})
	err := svc.CreateCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

type StartFailed CompStub

func (c StartFailed) startNode(_ context.Context, _ *compute.Node) error {
	return fmt.Errorf("failed to start container")
}

func TestStartClusterFails(t *testing.T) {
	svc := initCluster(t.Context(), StartFailed{})
	err := svc.StartCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

type StopFailed CompStub

func (c StopFailed) stopNode(_ context.Context, _ *compute.Node) error {
	return fmt.Errorf("failed to stop container")
}

func TestStopClusterFails(t *testing.T) {
	svc := initCluster(t.Context(), StopFailed{})
	err := svc.StopCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

type TeardownPasses CompStub

func (t TeardownPasses) removeNode(_ context.Context, _ *compute.Node) error {
	return nil
}

func TestTeardownCluster(t *testing.T) {
	svc := initCluster(t.Context(), TeardownPasses{})
	err := svc.TeardownCluster(t.Context())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

type RemoveFailed CompStub

func (c RemoveFailed) removeNode(ctx context.Context, _ *compute.Node) error {
	return fmt.Errorf("failed to remove container")
}

func TestTeardownClusterRemoveFails(t *testing.T) {
	svc := initCluster(t.Context(), RemoveFailed{})
	err := svc.TeardownCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}
