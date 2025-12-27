package core

import (
	"context"

	"github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
)

type CoreService struct {
	CompController model.ComputeController
	ServiceGraph   *model.Graph
}

func (core *CoreService) PlanCluster(ctx context.Context, cfg model.HCluster) {
	core.ServiceGraph = model.NewGraph(&cfg)
}

func (core *CoreService) CreateCluster(ctx context.Context, cfg model.HCluster) error {
	core.ServiceGraph.UpdateLevels(core.CompController.CreateNode)
	return core.ServiceGraph.ExecLevels(ctx)
}

func (core *CoreService) StartCluster(ctx context.Context, cfg model.HCluster) error {
	core.ServiceGraph.UpdateLevels(core.CompController.StartNode)
	return core.ServiceGraph.ExecLevels(ctx)
}

func (core *CoreService) StopCluster(ctx context.Context, cfg model.HCluster) error {
	core.ServiceGraph.UpdateLevels(core.CompController.StopNode)
	return core.ServiceGraph.ExecLevels(ctx)
}

func (core *CoreService) TeardownCluster(ctx context.Context, cfg model.HCluster) error {
	core.ServiceGraph.UpdateLevels(core.CompController.RemoveNode)
	return core.ServiceGraph.ExecLevels(ctx)
}

func NewCoreService() CoreService {
	return CoreService{
		CompController: compute.NewComputeController(compute.DOCKER),
	}
}
