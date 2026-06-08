package compute

import (
	"context"
	"fmt"
	"testing"

	"github.com/KyloRilo/helios/pkg/model/compute"
)

type TestCompCtrl struct{ CtrlShim }

func (c TestCompCtrl) createNode(_ context.Context, _ *compute.Node) (string, error) {
	return "", nil
}
func (c TestCompCtrl) startNode(_ context.Context, _ *compute.Node) error {
	return nil
}
func (c TestCompCtrl) listNodes(_ context.Context) ([]*compute.Node, error) {
	return nil, nil
}
func (c TestCompCtrl) stopNode(_ context.Context, _ *compute.Node) error {
	return nil
}
func (c TestCompCtrl) removeNode(_ context.Context, _ *compute.Node) error {
	return nil
}

func newTestCtrl(ctx context.Context, stub CtrlShim) (ComputeController, error) {
	return NewComputeController(ctx, ControllerArgs{
		stub: &stub,
	})
}

type CreatePasses struct{ TestCompCtrl }

func (c CreatePasses) createNode(_ context.Context, _ *compute.Node) (string, error) {
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

	if n.Id != "test-id" {
		t.Errorf("Expected node ID to be 'test-id' but got %s", n.Id)
	}

	if n.Status != compute.Created {
		t.Errorf("Expected node status to be 'Created' but got %s", n.Status)
	}
}

type CreateFailed struct{ TestCompCtrl }

func (c CreateFailed) createNode(_ context.Context, _ *compute.Node) (string, error) {
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

	if n.Id != "" {
		t.Errorf("Expected node ID to be empty but got %s", n.Id)
	}

	if n.Status != compute.Error {
		t.Errorf("Expected node status to be 'Error' but got %s", n.Status)
	}
}

type StartPasses struct{ CreatePasses }

func (c StartPasses) startNode(_ context.Context, _ *compute.Node) error {
	return nil
}

func TestStartNodeSuccess(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), StartPasses{})
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

	if n.Status != compute.Up {
		t.Errorf("Expected node status to be 'Up' but got %s", n.Status)
	}
}

type StartFailed struct{ CreatePasses }

func (c StartFailed) startNode(_ context.Context, _ *compute.Node) error {
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

	if n.Status != compute.Error {
		t.Errorf("Expected node status to be 'Error' but got %s", n.Status)
	}
}

type StopPasses struct{ StartPasses }

func (c StopPasses) stopNode(_ context.Context, _ *compute.Node) error {
	return nil
}

func TestStopNodeSuccess(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), StopPasses{})
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

	if n.Status != compute.Down {
		t.Errorf("Expected node status to be 'Down' but got %s", n.Status)
	}
}

type StopFailed struct{ StartPasses }

func (c StopFailed) stopNode(_ context.Context, _ *compute.Node) error {
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

	if n.Status != compute.Error {
		t.Errorf("Expected node status to be 'Error' but got %s", n.Status)
	}
}

type RemovePasses struct{ CreatePasses }

func TestRemoveNodeSuccess(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), RemovePasses{})
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

type RemoveFailed struct{ CreatePasses }

func (c RemoveFailed) removeNode(_ context.Context, _ *compute.Node) error {
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

	if n.Status != compute.Error {
		t.Errorf("Expected node status to be 'Error' but got %s", n.Status)
	}
}
