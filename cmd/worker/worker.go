package main

func main() {
	// TODO: worker node init method
	// c := startNode()
	// defer c.Shutdown(true)

	// _, stop := signal.NotifyContext(context.Background())
	// defer stop()
}

// func startNode() *cluster.Cluster {
// 	host, port := getHostInfo()
// 	system := actor.NewActorSystem()
// 	provider, err := k8s.New()
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	lookup := disthash.New()
// 	config := remote.Configure(host, port)

// 	c := cluster.New(system, cluster.Configure("worker-node", provider, lookup, config))
// 	c.StartMember()
// 	c.Remote.Register("worker", actor.PropsFromProducer(func() actor.Actor {
// 		return worker.InitWorkerService()
// 	}))

// 	return c
// }

// func getHostInfo() (string, int) {
// 	host := env.GetString("HELIOS_HOST", "127.0.0.1")
// 	port, err := env.GetInt("HELIOS_PORT", 6331)
// 	if err != nil {
// 		log.Panic(err)
// 	}

// 	return host, port
// }
