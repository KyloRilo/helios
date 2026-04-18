package integration

import (
	"testing"
	"time"

	"github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model/node"
)

func TestComputeLifcycle(t *testing.T) {
	ctx := t.Context()
	exec := func(ctrl compute.ComputeController, n *node.Node) error {
		_, err := ctrl.CreateNode(ctx, n)
		if err != nil {
			t.Errorf("Failed to Create Node '%s' => %s", n.GetName(), err)
		}

		if n.GetId() == "" {
			t.Errorf("Node ID is empty after creation for Node '%s'", n.GetName())
		}

		_, err = ctrl.StartNode(ctx, n)
		if err != nil {
			t.Errorf("Failed to Start Node '%s' => %s", n.GetName(), err)
		}

		time.Sleep(20 * time.Second)
		resp, err := ctrl.ListNodes(ctx)
		if err != nil {
			t.Errorf("Failed to List Services => %s", err)
		}

		if len(resp.Nodes) == 0 {
			t.Errorf("Expected at least 1 service but got 0")
		}

		for _, n := range resp.Nodes {
			t.Logf("Service: %s (ID: %s)", n.GetName(), n.GetId())
		}

		_, err = ctrl.StopNode(ctx, n)
		if err != nil {
			t.Errorf("Failed to Stop Node '%s' => %s", n.GetName(), err)
		}
		_, err = ctrl.RemoveNode(ctx, n)
		if err != nil {
			t.Errorf("Failed to Remove Node '%s' => %s", n.GetName(), err)
		}

		return nil
	}

	for _, svc := range []*node.Node{
		node.NewNode(
			node.WithName("test-1"),
			node.WithImage("alpine:latest"),
			node.WithPorts(map[string]string{
				"22": "22",
			}),
		),
	} {
		func() {
			t.Log("Testing service create")
			ctrl, err := compute.NewComputeController(t.Context(), compute.ControllerArgs{}, nil)
			if err != nil {
				t.Errorf("Failed to create ComputeController => %s", err)
			}

			exec(ctrl, svc)
		}()
	}

}
