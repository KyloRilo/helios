package core

import (
	"context"
	"fmt"

	compCtrl "github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/model/compute"
	"github.com/asynkron/protoactor-go/actor"
)

type CoreService struct {
	model.ActorService
	compCtrl    compCtrl.ComputeController
	conf        *model.HCluster
	stateMgrRef *actor.PID
	nodes       []*compute.Node
}

func (cs *CoreService) SetConfig(conf *model.HCluster) {
	cs.conf = conf
}

func (cs *CoreService) GetConfig() *model.HCluster {
	return cs.conf
}

func (cs *CoreService) SetNodes(nodes []*compute.Node) {
	cs.nodes = nodes
}

func (cs *CoreService) GetNodes() []*compute.Node {
	return cs.nodes
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

func (cs *CoreService) GenNodes(svcs []model.HService) []*compute.Node {
	fmt.Println("Generating Nodes...")
	nodes := []*compute.Node{}
	for _, svc := range svcs {
		fmt.Println("Generating: ", svc.Name)
		nodes = append(nodes, compute.NewNode(
			compute.WithName(svc.Name),
			compute.WithCmd(svc.Command),
			compute.WithPorts(svc.Ports),
			compute.WithVolumes(svc.Volumes),
			compute.WithEnv(svc.Environment),
			func() compute.NodeOption {
				switch {
				case svc.Build != nil:
					return compute.WithContext(&compute.Context{
						Path: svc.Build.Context,
						File: svc.Build.Dockerfile,
					})
				default:
					return compute.WithImage(svc.Image)
				}
			}(),
		))
	}

	fmt.Println("Final Nodes: ", nodes)
	return nodes
}

func (cs *CoreService) CreateCluster(ctx context.Context) error {
	failed := make(map[string]string)
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	ns := cs.GenNodes(cs.conf.Services)
	for _, n := range ns {
		var id string
		if id, err = cs.compCtrl.CreateNode(ctx, n); err != nil {
			failed[n.GetName()] = err.Error()
		}

		n.SetId(id)
	}

	if len(failed) != 0 {
		return fmt.Errorf("The following nodes failed to create => %s", failed)
	}

	cs.SetNodes(ns)
	return nil
}

func (cs *CoreService) StartCluster(ctx context.Context) error {
	failed := make(map[string]string)
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	for _, n := range cs.nodes {
		err := cs.compCtrl.StartNode(ctx, n)
		if err != nil {
			failed[n.GetId()] = err.Error()
		}
	}

	if len(failed) != 0 {
		return fmt.Errorf("The following nodes failed to start => %s", failed)
	}

	return nil
}

func (cs *CoreService) StopCluster(ctx context.Context) error {
	failed := make(map[string]string)
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	for _, n := range cs.nodes {
		err := cs.compCtrl.StopNode(ctx, n)
		if err != nil {
			failed[n.GetId()] = err.Error()
		}
	}

	print(failed)
	if len(failed) != 0 {
		return fmt.Errorf("The following nodes failed to stop => %s", failed)
	}

	return nil
}

func (cs *CoreService) TeardownCluster(ctx context.Context) error {
	failed := make(map[string]string)
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	for _, n := range cs.nodes {
		err := cs.compCtrl.RemoveNode(ctx, n)
		if err != nil {
			failed[n.GetId()] = err.Error()
		}
	}

	if len(failed) != 0 {
		return fmt.Errorf("Failed to destroy the following nodes => %s", failed)
	}

	return nil
}

func (cs CoreService) Receive(actx actor.Context) {
	switch actx.Message().(type) {
	case *actor.Started:
		fmt.Println("Started Core ActorService")

	}
}

type CoreArgs struct {
	Conf       *model.HCluster
	ScalerArgs compCtrl.ControllerArgs
}

func NewCoreService(ctx context.Context, args CoreArgs) CoreService {
	ctrl, err := compCtrl.NewComputeController(ctx, args.ScalerArgs)
	if err != nil {
		panic(err)
	}

	return CoreService{
		ActorService: model.NewBaseActorService("Core"),
		conf:         args.Conf,
		compCtrl:     ctrl,
	}
}
