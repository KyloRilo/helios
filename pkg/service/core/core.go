package core

import (
	"context"
	"fmt"

	"github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/model/node"
	"github.com/asynkron/protoactor-go/actor"
)

type CoreService struct {
	model.ActorService
	compCtrl    compute.ComputeController
	conf        *model.HCluster
	stateMgrRef *actor.PID
	nodes       []*node.Node
}

func (cs *CoreService) SetConfig(conf *model.HCluster) {
	cs.conf = conf
}

func (cs *CoreService) GetConfig() *model.HCluster {
	return cs.conf
}

func (cs *CoreService) SetNodes(nodes []*node.Node) {
	cs.nodes = nodes
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
	nodes := []*node.Node{}
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	for _, svc := range cs.conf.Services {
		n := node.NewNode(
			node.WithImage(svc.Image),
			node.WithName(svc.Name),
			node.WithCmd(svc.Command),
			node.WithPorts(svc.Ports),
			node.WithVolumes(svc.Volumes),
		)

		resp, err := cs.compCtrl.CreateNode(ctx, n)
		if err != nil {
			return fmt.Errorf("Failed to create service: %s", err)
		}

		nodes = append(nodes, resp.Node)
	}

	cs.SetNodes(nodes)
	return nil
}

func (cs *CoreService) StartCluster(ctx context.Context) error {
	failed := map[string]interface{}{}
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	for _, node := range cs.nodes {
		_, err = cs.compCtrl.StartNode(ctx, node)
		if err != nil {
			failed[node.GetId()] = map[string]interface{}{
				"Error": err,
				"Node":  node,
			}
		}
	}

	if len(failed) != 0 {
		return fmt.Errorf("The following nodes failed to start => %s", failed)
	}

	return nil
}

func (cs *CoreService) StopCluster(ctx context.Context) error {
	failed := map[string]interface{}{}
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	for _, node := range cs.nodes {
		_, err = cs.compCtrl.StopNode(ctx, node)
		if err != nil {
			failed[node.GetId()] = map[string]interface{}{
				"Error": err,
				"Node":  node,
			}
		}
	}

	if len(failed) != 0 {
		return fmt.Errorf("The following nodes failed to stop => %s", failed)
	}

	return nil
}

func (cs *CoreService) TeardownCluster(ctx context.Context) error {
	failed := map[string]interface{}{}
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	for _, node := range cs.nodes {
		_, err = cs.compCtrl.RemoveNode(ctx, node)
		if err != nil {
			failed[node.GetId()] = map[string]interface{}{
				"Error": err,
				"Node":  node,
			}
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
	ScalerArgs compute.ControllerArgs
}

func NewCoreService(ctx context.Context, args CoreArgs) CoreService {
	ctrl, err := compute.NewComputeController(ctx, args.ScalerArgs, nil)
	if err != nil {
		panic(err)
	}

	return CoreService{
		ActorService: model.NewBaseActorService("Core"),
		conf:         args.Conf,
		compCtrl:     ctrl,
	}
}
