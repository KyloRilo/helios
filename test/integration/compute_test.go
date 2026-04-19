package integration

import (
	"testing"
	"time"

	compCtrl "github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model/compute"
)

func TestComputeLifcycle(t *testing.T) {
	ctx := t.Context()
	exec := func(ctrl compCtrl.ComputeController, n *compute.Node) error {
		id, err := ctrl.CreateNode(ctx, n)
		if err != nil {
			t.Errorf("Failed to Create Node '%s' => %s", n.GetName(), err)
		}

		t.Log(n)

		if n.GetId() == "" {
			t.Errorf("Node ID is empty after creation for Node '%s'", n.GetName())
		}

		if n.GetId() != id {
			t.Errorf("Node ID mismatch after creation for Node '%s': expected '%s', got '%s'", n.GetName(), id, n.GetId())
		}

		if n.GetStatus() != compute.Created {
			t.Errorf("Node status is not 'Created' after creation for Node '%s': got '%s'", n.GetName(), n.GetStatus())
		}

		err = ctrl.StartNode(ctx, n)
		if err != nil {
			t.Errorf("Failed to Start Node '%s' => %s", n.GetName(), err)
		}

		if n.GetStatus() != compute.Up {
			t.Errorf("Node status is not 'Up' after starting for Node '%s': got '%s'", n.GetName(), n.GetStatus())
		}

		time.Sleep(20 * time.Second)
		ls, err := ctrl.ListNodes(ctx)
		if err != nil {
			t.Errorf("Failed to List Services => %s", err)
		}

		if len(ls) == 0 {
			t.Errorf("Expected at least 1 service but got 0")
		}

		for _, ctr := range ls {
			t.Logf("Service: %s (ID: %s)", ctr.GetName(), ctr.GetId())
		}

		err = ctrl.StopNode(ctx, n)
		if err != nil {
			t.Errorf("Failed to Stop Node '%s' => %s", n.GetName(), err)
		}

		if n.GetStatus() != compute.Down {
			t.Errorf("Node status is not 'Down' after stopping for Node '%s': got '%s'", n.GetName(), n.GetStatus())
		}

		err = ctrl.RemoveNode(ctx, n)
		if err != nil {
			t.Errorf("Failed to Remove Node '%s' => %s", n.GetName(), err)
		}

		if n.GetStatus() != compute.Destroyed {
			t.Errorf("Node status is not 'Destroyed' after removal for Node '%s': got '%s'", n.GetName(), n.GetStatus())
		}

		return nil
	}

	for _, svc := range []*compute.Node{
		compute.NewNode(
			compute.WithName("test-1"),
			compute.WithImage("alpine:latest"),
			compute.WithPorts(map[string]string{
				"22": "22",
			}),
		),
	} {
		func() {
			t.Log("Testing service create")
			ctrl, err := compCtrl.NewComputeController(t.Context(), compCtrl.ControllerArgs{})
			if err != nil {
				t.Errorf("Failed to create ComputeController => %s", err)
			}

			exec(ctrl, svc)
			resp, _ := ctrl.ListNodes(t.Context())
			for _, n := range resp {
				ctrl.StopNode(t.Context(), n)
				ctrl.RemoveNode(t.Context(), n)
			}
		}()
	}
}
