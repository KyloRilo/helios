package unit

import (
	"context"
	"fmt"
	"testing"

	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/service/core"
)

type CompStub struct{}

func (c CompStub) Authenticate(ctx context.Context) error {
	return nil
}

func (c CompStub) CreateNode(ctx context.Context, node *model.Node) error {
	return nil
}

func (c CompStub) StartNode(ctx context.Context, meta *model.Node) error {
	return nil
}

func (c CompStub) StopNode(ctx context.Context, meta *model.Node) error {
	return nil
}

func (c CompStub) RemoveNode(ctx context.Context, meta *model.Node) error {
	return nil
}

func TestCoreInit(t *testing.T) {
	_ = core.CoreService{CompController: CompStub{}}
}

type CreateSuccess struct{ CompStub }

func (c CreateSuccess) CreateNode(ctx context.Context, node *model.Node) error {
	return nil
}

func genServcConfig() model.HService {
	return model.HService{
		Name:  "test-service",
		Image: "test-image",
		Build: &model.Build{
			Context:    ".",
			Dockerfile: "./test/config/helios.hcl",
		},
	}
}

func TestCreateServicePasses(t *testing.T) {
	coreService := core.CoreService{CompController: CreateSuccess{}}
	coreService.CreateService(t.Context(), genServcConfig())
}

type CreateFailed struct{ CompStub }

func (c CreateFailed) CreateNode(ctx context.Context, _ *model.Node) error {
	return fmt.Errorf("failed to create container")
}

func TestCreateServiceCreateFails(t *testing.T) {
	coreService := core.CoreService{CompController: CreateFailed{}}
	err := coreService.CreateService(t.Context(), genServcConfig())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

type StartFailed struct{ CompStub }

func (c StartFailed) StartNode(ctx context.Context, _ *model.Node) error {
	return fmt.Errorf("failed to start container")
}

func TestCreateServiceStartFails(t *testing.T) {
	coreService := core.CoreService{CompController: StartFailed{}}
	err := coreService.CreateService(t.Context(), genServcConfig())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

func TestDestroyService(t *testing.T) {
	coreService := core.CoreService{CompController: CompStub{}}
	meta := &model.ContainerMeta{
		Id: "test-id",
	}
	err := coreService.DestroyService(t.Context(), meta)
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

type StopFailed struct{ CompStub }

func (c StopFailed) StopNode(ctx context.Context, _ *model.Node) error {
	return fmt.Errorf("failed to stop container")
}

func TestDestroyServiceStopFails(t *testing.T) {
	coreService := core.CoreService{CompController: StopFailed{}}
	meta := &model.ContainerMeta{
		Id: "test-id",
	}
	err := coreService.DestroyService(t.Context(), meta)
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

type RemoveFailed struct{ CompStub }

func (c RemoveFailed) RemoveNode(ctx context.Context, _ *model.Node) error {
	return fmt.Errorf("failed to remove container")
}

func TestDestroyServiceRemoveFails(t *testing.T) {
	coreService := core.CoreService{CompController: RemoveFailed{}}
	meta := &model.ContainerMeta{
		Id: "test-id",
	}
	err := coreService.DestroyService(t.Context(), meta)
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}
