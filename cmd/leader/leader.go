package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	heliosRaft "github.com/KyloRilo/helios/pkg/controller/raft"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/service/core"
	"github.com/KyloRilo/helios/pkg/service/leader"
	"github.com/hashicorp/raft"
)

func main() {
	configPath := flag.String("config", "/helios/config/config.hcl", "Path to leader config HCL")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	conf, err := model.ReadLeaderConfigFile(*configPath)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	if ok, err := conf.IsValid(); !ok {
		log.Fatalf("invalid config: %v", err)
	}

	nodeID := conf.NodeID
	if nodeID == "" {
		nodeID = conf.Name
	}

	raftBind := conf.RaftBind
	if raftBind == "" {
		raftBind = conf.Host + ":6330"
	}

	raftDir := conf.RaftDir
	if raftDir == "" {
		raftDir = "/tmp/helios-raft/" + nodeID
	}

	l, err := leader.NewLeader(ctx, leader.LeaderArgs{
		Conf: conf,
		CoreArgs: core.CoreArgs{
			Conf: nil,
		},
		RaftCfg: &heliosRaft.RaftConfig{
			NodeID:  nodeID,
			RaftDir: raftDir,
			Bind:    raftBind,
		},
	})
	if err != nil {
		log.Fatalf("failed to create leader: %v", err)
	}

	defer l.SpinDown()
	l.SpinUp()

	if conf.Bootstrap {
		servers := []raft.Server{
			{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(nodeID),
				Address:  raft.ServerAddress(raftBind),
			},
		}
		for peerID, peerAddr := range conf.RaftPeers {
			servers = append(servers, raft.Server{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(peerID),
				Address:  raft.ServerAddress(peerAddr),
			})
		}

		if err := l.BootstrapRaft(servers); err != nil {
			log.Fatalf("failed to bootstrap raft: %v", err)
		}
		log.Printf("Raft cluster bootstrapped with %d servers", len(servers))
	}

	log.Printf("Leader %s started (node_id=%s, raft_bind=%s, is_bootstrap=%v)",
		conf.Name, nodeID, raftBind, conf.Bootstrap)

	<-ctx.Done()
	log.Println("Shutting down leader")
}
