package cluster

import (
	"log"
	"time"

	"github.com/KyloRilo/helios/pkg/model"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/consul"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/hashicorp/consul/api"
)

type ClusterConfig struct {
	Consul   *api.Config
	Host     string
	Port     int
	Name     string
	SrvcGens func() []model.ActorService
}

type ClusterController struct {
	id      cluster.IdentityLookup
	cluster *cluster.Cluster
}

func (cc ClusterController) Shutdown() {
	cc.cluster.Shutdown(true)
}

func (cc ClusterController) Start() {
	cc.cluster.StartMember()
}

func (cc ClusterController) LogClusterEvents() {
	cc.cluster.ActorSystem.EventStream.Subscribe(func(evt interface{}) {
		switch e := evt.(type) {
		case *cluster.MemberJoinedEvent:
			log.Printf("Node joined: %s (%s)", e.Name(), e.Host)
			log.Printf("Member Kids: %v", e.GetKinds())
		case *cluster.MemberLeftEvent:
			log.Printf("Node left: %s", e.Name())
			// Optionally remove from Raft cluster
		}
	})
}

func (ac ClusterController) LogMembers() {
	for {
		for _, member := range ac.cluster.MemberList.Members().Members() {
			log.Printf("Discovered member: %s", member.Id)
			log.Printf("Member Host: %s (%s)", member.Host, member.Address())
			log.Printf("Member Kinds: %v", member.GetKinds())
			log.Printf(member.String())
		}
		time.Sleep(15 * time.Second)
	}
}

func NewClusterController(cfg *ClusterConfig) *ClusterController {
	id := disthash.New()

	config := remote.Configure(cfg.Host, cfg.Port)
	prvdr, err := consul.NewWithConfig(cfg.Consul)
	if err != nil {
		log.Panic(err)
	}

	sys := actor.NewActorSystem()
	kinds := []*cluster.Kind{}
	for _, srvc := range cfg.SrvcGens() {
		kinds = append(kinds, cluster.NewKind(srvc.GetServiceName(), actor.PropsFromProducer(func() actor.Actor {
			return srvc
		})))
	}

	clusterConf := cluster.Configure(cfg.Name, prvdr, id, config, cluster.WithKinds(kinds...))
	cluster := cluster.New(sys, clusterConf)
	return &ClusterController{
		id:      id,
		cluster: cluster,
	}
}
