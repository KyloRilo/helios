package core

import (
	"context"
	"fmt"

	compCtrl "github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/model/compute"
	"github.com/KyloRilo/helios/pkg/model/graph"
	"github.com/asynkron/protoactor-go/actor"
)

type CoreService struct {
	model.ActorService
	compCtrl    compCtrl.ComputeController
	conf        *model.HCluster
	stateMgrRef *actor.PID
	graph       *graph.ClusterGraph
}

func (cs *CoreService) SetConfig(conf *model.HCluster) {
	cs.conf = conf
}

func (cs *CoreService) GetConfig() *model.HCluster {
	return cs.conf
}

func (cs *CoreService) SetGraph(g *graph.ClusterGraph) {
	cs.graph = g
}

func (cs *CoreService) GetGraph() *graph.ClusterGraph {
	return cs.graph
}

func (cs *CoreService) GetNodes() []*compute.Node {
	if cs.graph == nil {
		return nil
	}
	return cs.graph.Nodes()
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

func (cs *CoreService) GenGraph(svcs []model.HService) (*graph.ClusterGraph, error) {
	fmt.Println("Generating Nodes...")
	nodes := []*compute.Node{}
	deps := make(map[string][]string, len(svcs))

	for _, svc := range svcs {
		fmt.Println("Generating: ", svc.Name)
		deps[svc.Name] = svc.DependsOn

		opts := []compute.NodeOption{
			compute.WithName(svc.Name),
			compute.WithCmd(svc.Command),
			compute.WithEntrypoint(svc.Entrypoint),
			compute.WithPorts(svc.Ports),
			compute.WithExpose(svc.Expose),
			compute.WithVolumes(svc.Volumes),
			compute.WithEnv(svc.Environment),
			compute.WithEnvFile(svc.EnvFile),
			compute.WithHostname(svc.Hostname),
			compute.WithWorkingDir(svc.WorkingDir),
			compute.WithUser(svc.User),
			compute.WithLabels(svc.Labels),
			compute.WithDependsOn(svc.DependsOn),
			compute.WithRestart(svc.Restart),
			compute.WithReplicas(svc.Replicas),
			compute.WithNetworks(svc.Networks),
			func() compute.NodeOption {
				switch {
				case svc.Build != nil:
					return compute.WithContext(&compute.Context{
						Path:   svc.Build.Context,
						File:   svc.Build.Dockerfile,
						Args:   svc.Build.Args,
						Target: svc.Build.Target,
					})
				default:
					return compute.WithImage(svc.Image)
				}
			}(),
		}

		if svc.Healthcheck != nil {
			opts = append(opts, compute.WithHealthcheck(&compute.Healthcheck{
				Test:        svc.Healthcheck.Test,
				Interval:    svc.Healthcheck.Interval,
				Timeout:     svc.Healthcheck.Timeout,
				Retries:     svc.Healthcheck.Retries,
				StartPeriod: svc.Healthcheck.StartPeriod,
			}))
		}

		if svc.Resources != nil {
			opts = append(opts, compute.WithResources(&compute.Resources{
				CPULimit:          svc.Resources.CPULimit,
				MemoryLimit:       svc.Resources.MemoryLimit,
				CPUReservation:    svc.Resources.CPUReservation,
				MemoryReservation: svc.Resources.MemoryReservation,
			}))
		}

		nodes = append(nodes, compute.NewNode(opts...))
	}

	topo, err := graph.New(deps)
	if err != nil {
		return nil, fmt.Errorf("failed to build topology: %w", err)
	}

	cg, err := graph.NewClusterGraph(topo, nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to build cluster graph: %w", err)
	}

	fmt.Println("Final Nodes: ", nodes)
	return cg, nil
}

func (cs *CoreService) CreateCluster(ctx context.Context) error {
	failed := make(map[string]string)
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	cg, err := cs.GenGraph(cs.conf.Services)
	if err != nil {
		return err
	}

	levels, err := cg.Levels()
	if err != nil {
		return err
	}

	for _, level := range levels {
		for _, n := range level {
			var id string
			if id, err = cs.compCtrl.CreateNode(ctx, n); err != nil {
				failed[n.Name] = err.Error()
			}
			n.Id = id
		}
	}

	if len(failed) != 0 {
		return fmt.Errorf("The following nodes failed to create => %s", failed)
	}

	cs.SetGraph(cg)
	return nil
}

func (cs *CoreService) StartCluster(ctx context.Context) error {
	failed := make(map[string]string)
	err := cs.ValidateCluster()
	if err != nil {
		return err
	}

	levels, err := cs.graph.Levels()
	if err != nil {
		return err
	}

	for _, level := range levels {
		for _, n := range level {
			if err := cs.compCtrl.StartNode(ctx, n); err != nil {
				failed[n.Id] = err.Error()
			}
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

	levels, err := cs.graph.ReverseLevels()
	if err != nil {
		return err
	}

	for _, level := range levels {
		for _, n := range level {
			if err := cs.compCtrl.StopNode(ctx, n); err != nil {
				failed[n.Id] = err.Error()
			}
		}
	}

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

	levels, err := cs.graph.ReverseLevels()
	if err != nil {
		return err
	}

	for _, level := range levels {
		for _, n := range level {
			if err := cs.compCtrl.RemoveNode(ctx, n); err != nil {
				failed[n.Id] = err.Error()
			}
		}
	}

	if len(failed) != 0 {
		return fmt.Errorf("Failed to destroy the following nodes => %s", failed)
	}

	return nil
}

func (cs *CoreService) PlanUpdate(desired *model.HCluster) (*graph.Plan, error) {
	if cs.graph == nil {
		return nil, fmt.Errorf("no current graph to diff against")
	}

	targets := make([]graph.DiffTarget, len(desired.Services))
	deps := make(map[string][]string, len(desired.Services))
	for i, svc := range desired.Services {
		targets[i] = graph.DiffTarget{
			Name:       svc.Name,
			Image:      svc.Image,
			Command:    svc.Command,
			Entrypoint: svc.Entrypoint,
			Restart:    svc.Restart,
			Replicas:   svc.Replicas,
			DependsOn:  svc.DependsOn,
		}
		deps[svc.Name] = svc.DependsOn
	}

	return graph.Diff(targets, deps, cs.graph)
}

func (cs CoreService) Receive(actx actor.Context) {
	switch actx.Message().(type) {
	case *actor.Started:
		fmt.Println("Started Core ActorService")
	}
}

type CoreArgs struct {
	stub       *compCtrl.ComputeController
	Conf       *model.HCluster
	ScalerArgs compCtrl.ControllerArgs
}

func NewCoreService(ctx context.Context, args CoreArgs) CoreService {
	var ctrl compCtrl.ComputeController
	if args.stub != nil {
		ctrl = *args.stub
	} else {
		var err error
		ctrl, err = compCtrl.NewComputeController(ctx, args.ScalerArgs)
		if err != nil {
			panic(err)
		}
	}

	return CoreService{
		ActorService: model.NewBaseActorService("Core"),
		conf:         args.Conf,
		compCtrl:     ctrl,
	}
}
