package main

import (
	"context"
	"fmt"
	"os"

	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/service/core"
)

type Launchpad struct {
	core.CoreService
	manifest *model.HManifest
}

func (l *Launchpad) InitCluster(ctx context.Context) error {
	cluster := l.manifest.Clusters[0]
	fmt.Println("Creating Cluster: '", cluster.Name, "'")
	return l.CreateCluster(ctx)
}

func (l *Launchpad) DestroyCluster(ctx context.Context) {
	cluster := l.manifest.Clusters[0]
	fmt.Println("Tearing Down Cluster: '", cluster.Name, "'")
	err := l.TeardownCluster(ctx)
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()
	path := os.Getenv("HELIOS_CONFIG_FILE")
	if path == "" {
		path = "/helios/config/cluster.hcl"
	}

	conf, err := model.ReadManifestFile(path)
	if err != nil {
		panic(err)
	}

	if ok, err := conf.IsValid(); !ok {
		panic(err)
	}

	lp := &Launchpad{
		CoreService: core.NewCoreService(ctx, core.CoreArgs{}),
		manifest:    conf,
	}
	err = lp.InitCluster(ctx)
	defer lp.DestroyCluster(ctx)

	if err != nil {
		panic(err)
	}
}
