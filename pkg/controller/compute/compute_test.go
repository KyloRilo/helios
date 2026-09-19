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

func TestIsValidProvider(t *testing.T) {
	valid := []Provider{ProviderDocker, ProviderECR, ProviderGCR}
	for _, p := range valid {
		if !IsValidProvider(p) {
			t.Errorf("expected provider '%s' to be valid", p)
		}
	}

	invalid := []Provider{ProviderTest, "", "random"}
	for _, p := range invalid {
		if IsValidProvider(p) {
			t.Errorf("expected provider '%s' to be invalid", p)
		}
	}
}

func TestIsECR(t *testing.T) {
	if !isECR("123456789.dkr.ecr.us-east-1.amazonaws.com/myrepo:latest") {
		t.Error("expected ECR image to be detected")
	}
	if isECR("nginx:latest") {
		t.Error("expected non-ECR image to not match")
	}
	if isECR("gcr.io/myproject/myimage") {
		t.Error("expected GCR image to not match ECR")
	}
}

func TestIsDockerHub(t *testing.T) {
	if !isDockerHub("nginx:latest") {
		t.Error("expected short name to be DockerHub")
	}
	if !isDockerHub("library/nginx") {
		t.Error("expected library/ prefix to be DockerHub")
	}
	if !isDockerHub("docker.io/myuser/myimage") {
		t.Error("expected docker.io prefix to be DockerHub")
	}
	if isDockerHub("gcr.io/myproject/myimage") {
		t.Error("expected gcr.io to not be DockerHub")
	}
	if isDockerHub("123456789.dkr.ecr.us-east-1.amazonaws.com/myrepo") {
		t.Error("expected ECR to not be DockerHub")
	}
}

func TestGetProvider(t *testing.T) {
	stub := TestCompCtrl{}
	var shim CtrlShim = stub
	ctrl, err := NewComputeController(t.Context(), ControllerArgs{stub: &shim})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if ctrl.GetProvider() != "" {
		t.Errorf("expected empty provider for stub but got '%s'", ctrl.GetProvider())
	}
}

func TestListNodes(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), TestCompCtrl{})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	nodes, err := ctrl.ListNodes(t.Context())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
	if nodes != nil {
		t.Errorf("expected nil nodes from stub but got %v", nodes)
	}
}

func TestCreateNodeInvalidStatus(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), CreatePasses{})
	if err != nil {
		t.Fatalf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	_, err = ctrl.CreateNode(t.Context(), n)
	if err != nil {
		t.Fatalf("first create should pass: %v", err)
	}

	_, err = ctrl.CreateNode(t.Context(), n)
	if err == nil {
		t.Error("expected error when creating an already-created node")
	}
}

func TestStartNodeInvalidStatus(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), StartPasses{})
	if err != nil {
		t.Fatalf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	err = ctrl.StartNode(t.Context(), n)
	if err == nil {
		t.Error("expected error when starting a Ready node")
	}
}

func TestStopNodeInvalidStatus(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), StopPasses{})
	if err != nil {
		t.Fatalf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	err = ctrl.StopNode(t.Context(), n)
	if err == nil {
		t.Error("expected error when stopping a Ready node")
	}
}

func TestRemoveNodeInvalidStatus(t *testing.T) {
	ctrl, err := newTestCtrl(t.Context(), RemovePasses{})
	if err != nil {
		t.Fatalf("Failed to create ComputeController => %s", err)
	}

	n := compute.NewNode(compute.WithName("test-node"))
	err = ctrl.RemoveNode(t.Context(), n)
	if err == nil {
		t.Error("expected error when removing a Ready node")
	}
}

func TestGetExpectedStatusesDefault(t *testing.T) {
	action := NodeAction("Invalid")
	statuses := action.getExpectedStatuses()
	if statuses != nil {
		t.Errorf("expected nil for unknown action but got %v", statuses)
	}
}
