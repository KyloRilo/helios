package core

import (
	"context"
	"fmt"

	"github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/asynkron/protoactor-go/actor"
)

type CoreService struct {
	model.ActorService
	compScaler  compute.ComputeController
	conf        *model.HCluster
	stateMgrRef *actor.PID
}

func (cs *CoreService) SetConfig(conf *model.HCluster) {
	cs.conf = conf
}

func (cs *CoreService) GetConfig() *model.HCluster {
	return cs.conf
}

func (cs *CoreService) ValidateCluster() error {
	if cs.conf == nil {
		return fmt.Errorf("No cluster config provided")
	}

	valid, err := cs.conf.IsValid()
	if err != nil {
		return fmt.Errorf("Error validating cluster config: %s", err)
	}

	if !valid {
		return fmt.Errorf("Cluster config is invalid")
	}

	return nil
}

func (cs *CoreService) CreateCluster(ctx context.Context) error {
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	return nil
}

func (cs *CoreService) StartCluster(ctx context.Context) error {
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	return nil
}

func (cs *CoreService) PlanCluster(ctx context.Context, _ *model.HCluster) error {
	return nil
}

func (cs *CoreService) StopCluster(ctx context.Context) error {
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	return nil
}

func (cs *CoreService) TeardownCluster(ctx context.Context) error {
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	return nil
}

func (cs CoreService) Receive(actx actor.Context) {
	switch actx.Message().(type) {
	case *actor.Started:
		fmt.Println("Started Core ActorService")

	}
}

func NewCoreService(conf *model.HCluster) CoreService {
	return CoreService{
		ActorService: model.NewBaseActorService("Core"),
		conf:         conf,
		compScaler:   compute.NewComputeController(nil),
	}
}
