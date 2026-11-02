package leader

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/KyloRilo/helios/pkg/controller/actor"
	"github.com/KyloRilo/helios/pkg/controller/consul"
	heliosRaft "github.com/KyloRilo/helios/pkg/controller/raft"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/model/graph"
	"github.com/KyloRilo/helios/pkg/service/core"
	protoActor "github.com/asynkron/protoactor-go/actor"
	"github.com/hashicorp/raft"
)

var ErrNotLeader = fmt.Errorf("this node is not the raft leader")

type LeaderService struct {
	model.ActorService
	mtx         sync.RWMutex
	ctx         context.Context
	core        *core.CoreService
	consulCtrl  *consul.ConsulController
	clusterCtrl *actor.ActorSysController
	raft        *heliosRaft.RaftService
}

func (l *LeaderService) IsLeader() bool {
	if l.raft == nil {
		return false
	}
	return l.raft.IsLeader()
}

func (l *LeaderService) requireLeader() error {
	if !l.IsLeader() {
		return ErrNotLeader
	}
	return nil
}

func (l *LeaderService) SpinUp() {
	if l.clusterCtrl != nil {
		l.clusterCtrl.Start()
	}
}

func (l *LeaderService) SpinDown() {
	if l.clusterCtrl != nil {
		l.clusterCtrl.Shutdown()
	}
}

func (l *LeaderService) GetCore() *core.CoreService {
	return l.core
}

func (l *LeaderService) GetRaft() *heliosRaft.RaftService {
	return l.raft
}

func (l *LeaderService) CreateCluster(ctx context.Context) error {
	if err := l.requireLeader(); err != nil {
		return err
	}
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.core.CreateCluster(ctx)
}

func (l *LeaderService) StartCluster(ctx context.Context) error {
	if err := l.requireLeader(); err != nil {
		return err
	}
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.core.StartCluster(ctx)
}

func (l *LeaderService) StopCluster(ctx context.Context) error {
	if err := l.requireLeader(); err != nil {
		return err
	}
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.core.StopCluster(ctx)
}

func (l *LeaderService) TeardownCluster(ctx context.Context) error {
	if err := l.requireLeader(); err != nil {
		return err
	}
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.core.TeardownCluster(ctx)
}

func (l *LeaderService) PlanUpdate(desired *model.HCluster) (*graph.Plan, error) {
	if err := l.requireLeader(); err != nil {
		return nil, err
	}
	l.mtx.RLock()
	defer l.mtx.RUnlock()
	return l.core.PlanUpdate(desired)
}

func (l *LeaderService) BootstrapRaft(peers []raft.Server) error {
	if l.raft == nil {
		return fmt.Errorf("raft not initialized")
	}
	return l.raft.BootstrapCluster(peers)
}

func (l *LeaderService) Receive(actx protoActor.Context) {
	switch msg := actx.Message().(type) {
	case *protoActor.Started:
		log.Printf("Leader actor started: %s", l.GetServiceName())
	case *protoActor.Stopping:
		log.Printf("Leader actor stopping: %s", l.GetServiceName())
	case *protoActor.Stopped:
		log.Printf("Leader actor stopped: %s", l.GetServiceName())
	default:
		_ = msg
	}
}

func GenServiceMap(l *LeaderService) func() []model.ActorService {
	return func() []model.ActorService {
		return []model.ActorService{l}
	}
}

type LeaderArgs struct {
	Conf     *model.LeaderConfig
	CoreArgs core.CoreArgs
	RaftCfg  *heliosRaft.RaftConfig
	Stub     bool
}

func NewLeader(ctx context.Context, args LeaderArgs) (*LeaderService, error) {
	cs := core.NewCoreService(ctx, args.CoreArgs)

	l := &LeaderService{
		ActorService: model.NewBaseActorService("Leader"),
		ctx:          ctx,
		core:         &cs,
	}

	if args.Stub {
		return l, nil
	}

	consulCtrl := consul.NewConsulController()
	l.consulCtrl = consulCtrl

	if args.RaftCfg != nil {
		raftSvc, err := heliosRaft.InitRaftService(*args.RaftCfg)
		if err != nil {
			return nil, fmt.Errorf("leader: raft init: %w", err)
		}
		l.raft = raftSvc
	}

	clusterName := args.Conf.ClusterName
	if clusterName == "" {
		clusterName = "helios"
	}

	l.clusterCtrl = actor.NewActorSysController(&actor.ClusterConfig{
		Consul:   consulCtrl.GetConfig(),
		Name:     clusterName,
		Host:     args.Conf.Host,
		Port:     args.Conf.Port,
		SrvcGens: GenServiceMap(l),
	})

	return l, nil
}
