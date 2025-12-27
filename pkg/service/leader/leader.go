package leader

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/KyloRilo/helios/pkg/controller/actor"
	"github.com/KyloRilo/helios/pkg/controller/consul"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/service/raft"
)

type LeaderService struct {
	model.ActorService
	ctx         context.Context
	mtx         sync.RWMutex
	consulCtrl  *consul.ConsulController
	clusterCtrl *actor.ActorSysController
	raft        raft.RaftService
}

func (l *LeaderService) SpinDown() {
	l.clusterCtrl.Shutdown()
}

func (l *LeaderService) SpinUp() {
	l.clusterCtrl.Start()
}

func (l *LeaderService) Hello() {
	for {
		time.Sleep(15 * time.Second)
		fmt.Print("Hello")
	}
}

// func GenServiceMap() func() []model.ActorService {
// 	return func() []model.ActorService {
// 		return []model.ActorService{
// 			core.InitCoreService(),
// 		}
// 	}
// }

func NewLeader(ctx context.Context, configDir string) *LeaderService {
	consulCtrl := consul.NewConsulController()
	conf, err := model.ReadLeaderConfigFile(configDir)
	if err != nil {
		log.Fatal(err)
	}

	if ok, err := conf.IsValid(); !ok {
		log.Fatal(err)
	}

	return &LeaderService{
		consulCtrl: consulCtrl,
		clusterCtrl: actor.NewActorSysController(&actor.ClusterConfig{
			Consul: consulCtrl.GetConfig(),
			Name:   conf.Name,
			Host:   conf.Host,
			Port:   conf.Port,
			// SrvcGens: GenServiceMap(),
		}),
	}
}
