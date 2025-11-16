package leader

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/KyloRilo/helios/pkg/controller/cluster"
	"github.com/KyloRilo/helios/pkg/controller/consul"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/service/core"
)

type LeaderService struct {
	ctx         context.Context
	mtx         sync.RWMutex
	consulCtrl  *consul.ConsulController
	clusterCtrl *cluster.ClusterController
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

func GenServiceMap(cfg *model.ClusterConfig) func() []model.ActorService {
	return func() []model.ActorService {
		return []model.ActorService{
			core.InitCoreService(cfg),
		}
	}
}

func NewLeader(ctx context.Context, configDir string) *LeaderService {
	conf := &model.ClusterConfig{}
	consulCtrl := consul.NewConsulController()
	err := conf.ReadFile(configDir)
	if err != nil {
		log.Fatal(err)
	}

	return &LeaderService{
		consulCtrl: consulCtrl,
		clusterCtrl: cluster.NewClusterController(&cluster.ClusterConfig{
			Consul:   consulCtrl.GetConfig(),
			Name:     conf.Name,
			Host:     conf.Host,
			Port:     conf.Port,
			SrvcGens: GenServiceMap(conf),
		}),
	}
}
