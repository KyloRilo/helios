package leader

import (
	"context"
	"log"
	"os"
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

type LeaderService struct {
	ctx     context.Context
	mtx     sync.RWMutex
	cluster *cluster.Cluster
	consul  *api.Client
	core    core.CoreService
	state   state.StateService
}

func (l *LeaderService) Shutdown() {
	l.cluster.Shutdown(true)
}

func (l *LeaderService) RegisterMembers() {
	members := l.cluster.MemberList.Members()
	log.Printf("RegisterMembers: %v", members)

	l.cluster.ActorSystem.EventStream.Subscribe(func(evt interface{}) {
		switch e := evt.(type) {
		case *cluster.MemberJoinedEvent:
			log.Printf("Node joined: %s (%s)", e.Name(), e.Host)
			log.Printf("Member Kids: %v", e.GetKinds())
		case *cluster.MemberLeftEvent:
			log.Printf("Node left: %s", e.Name())
			// Optionally remove from Raft cluster
		}
	})
	// Optionally store sub for later Unsubscribe
}

func (l *LeaderService) ListMembers() {
	for {
		services, _, err := l.consul.Catalog().Services(nil)
		if err != nil {
			log.Printf("Error querying services: %v", err)
		} else {
			log.Println("Services in catalog:")
			for name := range services {
				log.Printf("- %s", name)
			}
		}

		nodes, _, err := l.consul.Catalog().Nodes(nil)
		if err != nil {
			log.Printf("Error querying services: %v", err)
		} else {
			log.Println("Nodes in catalog:")
			for i := 0; i < len(nodes); i++ {
				log.Printf("- %s", nodes[i].Node)
			}
		}
		time.Sleep(15 * time.Second)
	}
}

func (l *LeaderService) DiscoverPeers() {
	for {
		members := l.cluster.MemberList.Members().Members()
		log.Printf("RegisterMembers: %v", members)
		time.Sleep(15 * time.Second)
	}
}

func InitLeaderService(ctx context.Context, host string, port int) LeaderService {
	consulConfig := api.DefaultConfig()
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr != "" {
		consulConfig.Address = consulAddr
	}

	client, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	system := actor.NewActorSystem()
	provider, err := consul.NewWithConfig(consulConfig)
	if err != nil {
		log.Panic(err)
	}

	lookup := disthash.New()
	config := remote.Configure(host, port)
	leader := LeaderService{
		ctx:    ctx,
		consul: client,
		//core:   core.InitCoreService(mngr),
		state: state.InitStateService(),
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
	leader.RegisterMembers()
	return leader
}
