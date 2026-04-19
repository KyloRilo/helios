package compute

import (
	"context"
	"fmt"
	"testing"

	"github.com/KyloRilo/helios/pkg/model/compute"
)

type TestCompCtrl ComputeController
type TestCompCtrlImpl CompImpl

func (c TestCompCtrlImpl) GetProvider() Provider {
	return ProviderTest
}

func (c TestCompCtrlImpl) CreateNode(_ context.Context, _ *compute.Node) (string, error) {
	return "", nil
}
func (c TestCompCtrlImpl) StartNode(_ context.Context, _ *compute.Node) error {
	return nil
}
func (c TestCompCtrlImpl) ListNodes(_ context.Context) ([]*compute.Node, error) {
	return nil, nil
}
func (c TestCompCtrlImpl) StopNode(_ context.Context, _ *compute.Node) error {
	return nil
}
func (c TestCompCtrlImpl) RemoveNode(_ context.Context, _ *compute.Node) error {
	return nil
}

func newTestCtrl(ctx context.Context, stub CtrlShim) (ComputeController, error) {
	ctrl := TestCompCtrlImpl{
		CtrlShim: stub,
	}

	return ctrl, nil
}

type CreatePasses TestCompCtrlImpl

func (c CreatePasses) CreateNode(_ context.Context, _ *compute.Node) (string, error) {
	return "test-id", nil
}

func TestCreateNodeSuccess(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), CreatePasses{})
	if err != nil {
		t.Errorf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	id, err := ctrl.CreateNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	if id != "test-id" {
		t.Errorf("Expected id to be 'test-id' but got %s", id)
	}

	if n.GetId() != "test-id" {
		t.Errorf("Expected node ID to be 'test-id' but got %s", n.GetId())
	}

	if n.GetStatus() != compute.Created {
		t.Errorf("Expected node status to be 'Created' but got %s", n.GetStatus())
	}
}

type CreateFailed TestCompCtrlImpl

func (c CreateFailed) CreateNode(_ context.Context, _ *compute.Node) (string, error) {
	return "", fmt.Errorf("failed to create node")
}

func TestCreateNodeFails(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), CreateFailed{})
	if err != nil {
		t.Errorf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	id, err := ctrl.CreateNode(t.Context(), n)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}

	if id != "" {
		t.Errorf("Expected id to be empty but got %s", id)
	}

	if n.GetId() != "" {
		t.Errorf("Expected node ID to be empty but got %s", n.GetId())
	}

	if n.GetStatus() != compute.Error {
		t.Errorf("Expected node status to be 'Error' but got %s", n.GetStatus())
	}
}

func TestStartNodeSuccess(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), CreatePasses{})
	if err != nil {
		t.Errorf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	_, err = ctrl.CreateNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	err = ctrl.StartNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	if n.GetStatus() != compute.Up {
		t.Errorf("Expected node status to be 'Up' but got %s", n.GetStatus())
	}
}

type StartFailed TestCompCtrlImpl

func (c StartFailed) StartNode(_ context.Context, _ *compute.Node) error {
	return fmt.Errorf("failed to start container")
}

func TestStartNodeFails(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), StartFailed{})
	if err != nil {
		t.Errorf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	_, err = ctrl.CreateNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	err = ctrl.StartNode(t.Context(), n)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}

	if n.GetStatus() != compute.Error {
		t.Errorf("Expected node status to be 'Error' but got %s", n.GetStatus())
	}
}

func TestStopNodeSuccess(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), CreatePasses{})
	if err != nil {
		t.Errorf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	_, err = ctrl.CreateNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	err = ctrl.StartNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	err = ctrl.StopNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	if n.GetStatus() != compute.Down {
		t.Errorf("Expected node status to be 'Down' but got %s", n.GetStatus())
	}
}

type StopFailed TestCompCtrlImpl

func (c StopFailed) StopNode(_ context.Context, _ *compute.Node) error {
	return fmt.Errorf("failed to stop container")
}

func TestStopNodeFails(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), StopFailed{})
	if err != nil {
		t.Errorf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	_, err = ctrl.CreateNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	err = ctrl.StartNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	err = ctrl.StopNode(t.Context(), n)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}

	if n.GetStatus() != compute.Error {
		t.Errorf("Expected node status to be 'Error' but got %s", n.GetStatus())
	}
}

func TestRemoveNodeSuccess(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), CreatePasses{})
	if err != nil {
		t.Errorf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	_, err = ctrl.CreateNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	err = ctrl.RemoveNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}
}

type RemoveFailed TestCompCtrlImpl

func (c RemoveFailed) RemoveNode(_ context.Context, _ *compute.Node) error {
	return fmt.Errorf("failed to remove container")
}

func TestRemoveNodeFails(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), RemoveFailed{})
	if err != nil {
		t.Errorf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	_, err = ctrl.CreateNode(t.Context(), n)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	err = ctrl.RemoveNode(t.Context(), n)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}
}
