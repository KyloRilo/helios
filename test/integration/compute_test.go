package integration

import (
	"context"
	"testing"

	"github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
)

func initCompCtrl() model.ComputeController {
	return compute.NewComputeController(compute.DOCKER)
}

func TestInitComputeController(t *testing.T) {
	_ = initCompCtrl()
}

func TestComputeLifcycle(t *testing.T) {
	ctrl := initCompCtrl()
	exec := func(ctx context.Context, node *model.Node) error {
		err := ctrl.CreateNode(ctx, node)
		if err != nil {
			t.Errorf("Failed to Create Node '%s' => %s", node.Meta.Name, err)
		}

		if node.Meta.ID == "" {
			t.Errorf("Node ID is empty after creation for Node '%s'", node.Meta.Name)
		}

		err = ctrl.StartNode(ctx, node)
		if err != nil {
			t.Errorf("Failed to Start Node '%s' => %s", node.Meta.Name, err)
		}
		err = ctrl.StopNode(ctx, node)
		if err != nil {
			t.Errorf("Failed to Stop Node '%s' => %s", node.Meta.Name, err)
		}
		err = ctrl.RemoveNode(ctx, node)
		if err != nil {
			t.Errorf("Failed to Remove Node '%s' => %s", node.Meta.Name, err)
		}

		return nil
	}

	for _, node := range []*model.Node{
		{
			Exec: exec,
			Meta: model.NodeMeta{
				Name:  "comp-test-1",
				Image: "alpine:latest",
			},
		},
	} {
		node.Exec(t.Context(), node)
	}

}
