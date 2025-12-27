package core

import (
	"context"
	"fmt"

	"github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
)

type CoreService struct {
	Cluster    *model.HeliosCluster
	SessionMap map[string]model.ComputeController
}

func (cs *CoreService) InitCluster(ctx context.Context, conf *model.HCluster, compStub model.ComputeController) error {
	cs.Cluster = model.NewHeliosCluster(*conf)
	cs.Cluster.Init(func(svc model.HService) model.ExecCtx {
		ctrl, err := compute.NewComputeController(svc.Image, &compStub)
		if err != nil {
			panic(err)
		}

		_, err = ctrl.Authenticate(ctx)
		if err != nil {
			panic(err)
		}

		cs.SessionMap[svc.Name] = ctrl
		return model.ExecCtx{
			Create: ctrl.CreateNode,
			Start:  ctrl.StartNode,
			Read:   nil,
			Update: nil,
			Stop:   ctrl.StopNode,
			Delete: ctrl.RemoveNode,
		}
	})

	return nil
}

func (cs *CoreService) ValidateCluster(_ *model.HCluster) error {
	if cs.Cluster == nil {
		return fmt.Errorf("Cluster found nil. Initialize cluster before proceeding")
	}

	return cs.Cluster.Validate()
}

func (cs *CoreService) CreateCluster(ctx context.Context, newConf *model.HCluster) error {
	err := cs.ValidateCluster(newConf)
	if err != nil {
		return err
	}

	return cs.Cluster.Create(ctx)
}

func (cs *CoreService) StartCluster(ctx context.Context, newConf *model.HCluster) error {
	err := cs.ValidateCluster(newConf)
	if err != nil {
		return err
	}

	return cs.Cluster.Start(ctx)
}

func (cs *CoreService) PlanCluster(ctx context.Context, _ *model.HCluster) error {
	return cs.Cluster.Validate()
}

func (cs *CoreService) StopCluster(ctx context.Context, newConf *model.HCluster) error {
	err := cs.ValidateCluster(newConf)
	if err != nil {
		return err
	}

	return cs.Cluster.Stop(ctx)
}

func (cs *CoreService) TeardownCluster(ctx context.Context, newConf *model.HCluster) error {
	err := cs.ValidateCluster(newConf)
	if err != nil {
		return err
	}

	return cs.Cluster.Teardown(ctx)
}

func NewCoreService() CoreService {
	return CoreService{
		SessionMap: make(map[string]model.ComputeController),
	}
}
