package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/KyloRilo/helios/pkg/service/leader"
)

func getHostInfo() (host string, port int) {
	host = os.Getenv("HELIOS_HOST")
	port, err := strconv.Atoi(os.Getenv("HELIOS_PORT"))
	if err != nil {
		log.Panic(err)
	}

	return host, port
}

func mainOld() {
	// ctx := context.Background()
	// host, port := getHostInfo()
	// nodeId := os.Getenv("NODE_ID")
	// raftDir := filepath.Join(os.Getenv("RAFT_DIR"), nodeId)
	// os.MkdirAll(raftDir, 0700)

	// mngr := heliosRaft.InitRaftManager(nodeId, raftDir, fmt.Sprintf(`%s:800%s`, host, nodeId))
	// if nodeId == "1" {
	// 	mngr.BootstrapCluster()
	// 	time.Sleep(60 * time.Second)
	// 	mngr.AddVoter()
	// } else {
	// 	// TODO: have nodes request to join the cluster via gRPC
	// 	// Consul service mesh could help here
	// }

	// leader := leader.InitLeaderService(ctx, host, port)
	// defer leader.Shutdown()

	// _, stop := signal.NotifyContext(ctx)
	// defer stop()

	// leader.DiscoverPeers()
	// leader.ListMembers()
}

func main() {
	ctx := context.Background()
	configDir := flag.String("config", "/helios/config/config.hcl", "Path to helios config hcl")
	leader := leader.NewLeader(ctx, *configDir)
	defer leader.SpinDown()

	_, stop := signal.NotifyContext(ctx)
	defer stop()

	leader.Hello()
}
