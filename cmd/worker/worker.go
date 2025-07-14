package main

import (
	"context"
	"log"
	"os/signal"

	"github.com/KyloRilo/helios/pkg/service/worker"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/k8s"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"k8s.io/utils/env"
)

func main() {
	// TODO: worker node init method
	c := startNode()
	defer c.Shutdown(true)

	_, stop := signal.NotifyContext(context.Background())
	defer stop()
}

func startNode() *cluster.Cluster {
	host, port, advertHost := getHostInfo()
	system := actor.NewActorSystem()
	provider, err := k8s.New()
	if err != nil {
		log.Panic(err)
	}
	lookup := disthash.New()
	config := remote.Configure(host, port, remote.WithAdvertisedHost(advertHost))

	c := cluster.New(system, cluster.Configure("worker-node", provider, lookup, config))
	c.StartMember()
	c.Remote.Register("worker", actor.PropsFromProducer(func() actor.Actor {
		return worker.InitWorkerService()
	}))

	return c
}

func getHostInfo() (string, int, string) {
	host := env.GetString("HELIOS_HOST", "127.0.0.1")
	port, err := env.GetInt("HELIOS_PORT", 6331)
	if err != nil {
		log.Panic(err)
	}

	advertHost := env.GetString("HELIOS_ADVERT_HOST", "")

	return host, port, advertHost
}
