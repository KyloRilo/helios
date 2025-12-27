package integration

import (
	"testing"

	"github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
)

func TestComputeLifcycle(t *testing.T) {
	ctx := t.Context()
	exec := func(ctrl model.ComputeController, node *model.Node) error {
		_, err := ctrl.CreateNode(ctx, node)
		if err != nil {
			t.Errorf("Failed to Create Node '%s' => %s", node.Meta.Name, err)
		}

		if node.Meta.ID == "" {
			t.Errorf("Node ID is empty after creation for Node '%s'", node.Meta.Name)
		}

		_, err = ctrl.StartNode(ctx, node)
		if err != nil {
			t.Errorf("Failed to Start Node '%s' => %s", node.Meta.Name, err)
		}
		_, err = ctrl.StopNode(ctx, node)
		if err != nil {
			t.Errorf("Failed to Stop Node '%s' => %s", node.Meta.Name, err)
		}
		_, err = ctrl.RemoveNode(ctx, node)
		if err != nil {
			t.Errorf("Failed to Remove Node '%s' => %s", node.Meta.Name, err)
		}

		return nil
	}

	for _, node := range []*model.Node{
		{
			Meta: model.NodeMeta{
				Name:  "comp-test-1",
				Image: "alpine:latest",
			},
		},
	} {
		go func() {
			ctrl, err := compute.NewComputeController(node.Meta.Image, nil)
			if err != nil {
				t.Errorf("Failed to create compute controller for image '%s' => %s", node.Meta.Image, err)
			}

			exec(ctrl, node)
		}()
	}

}
