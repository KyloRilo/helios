package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/KyloRilo/helios/pkg/service/core"
	"github.com/KyloRilo/helios/pkg/service/state"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/consul"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/hashicorp/consul/api"
)

// Leader Raft Impl
type LeaderService struct {
	ctx          context.Context
	mtx          sync.RWMutex
	cluster      *cluster.Cluster
	consulClient *api.Client
	core         core.CoreService
	state        state.StateService
}

func (leader *LeaderService) Shutdown() {
	leader.cluster.Shutdown(true)
}

func (leader *LeaderService) ListMembers() {
	for {
		services, _, err := leader.consulClient.Catalog().Services(nil)
		if err != nil {
			log.Printf("Error querying services: %v", err)
		} else {
			log.Println("Services in catalog:")
			for name := range services {
				log.Printf("- %s", name)
			}
		}
		time.Sleep(15 * time.Second)
	}
}

func initLeaderService(ctx context.Context) LeaderService {
	consulConfig := api.DefaultConfig()
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr != "" {
		consulConfig.Address = consulAddr
	}

	client, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	host, port, _ := getHostInfo()
	system := actor.NewActorSystem()
	provider, err := consul.NewWithConfig(consulConfig)
	if err != nil {
		log.Panic(err)
	}

	lookup := disthash.New()
	config := remote.Configure(host, port)
	leader := LeaderService{
		ctx:          ctx,
		consulClient: client,
		core:         core.InitCoreService(),
		state:        state.InitStateService(),
	}

	cluster := cluster.New(system, cluster.Configure(
		"helios-leader", provider, lookup, config, cluster.WithKinds(
			cluster.NewKind("helios-core", actor.PropsFromProducer(func() actor.Actor {
				return leader.core
			})),
			cluster.NewKind("helios-state", actor.PropsFromProducer(func() actor.Actor {
				return leader.state
			})),
		),
	))

	leader.cluster = cluster
	cluster.StartMember()
	return leader
}

func getHostInfo() (string, int, string) {
	host := os.Getenv("HELIOS_HOST")
	advertHost := os.Getenv("HELIOS_ADVERT_HOST")
	port, err := strconv.Atoi(os.Getenv("HELIOS_PORT"))
	if err != nil {
		log.Panic(err)
	}

	return host, port, advertHost
}

func main() {
	ctx := context.Background()
	leader := initLeaderService(ctx)
	defer leader.Shutdown()

	_, stop := signal.NotifyContext(ctx)
	defer stop()

	go leader.ListMembers()
}
