package main

import (
	"context"
	"fmt"
	"os"

	"github.com/KyloRilo/helios/pkg/controller/docker"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/service/core"
)

type Launchpad struct {
	core   core.CoreService
	config *model.HeliosConfig
}

func (l *Launchpad) InitCluster(ctx context.Context) error {
	cluster := l.config.Clusters[0]
	fmt.Println("Creating Cluster: '", cluster.Name, "'")

	for _, service := range cluster.Services {
		fmt.Println("Creating Service: ", service.Name)
		err := l.core.CreateService(ctx, service)
		if err != nil {
			return fmt.Errorf("Failed to create service '%s' => %s", service.Name, err)
		}
	}

	return nil
}

func NewLaunchpad(config *model.HeliosConfig) *Launchpad {
	return &Launchpad{
		config: config,
		core: core.CoreService{
			Docker: docker.InitDockerController(),
		},
	}
}

func main() {
	ctx := context.Background()
	path := os.Getenv("HELIOS_CONFIG_FILE")
	if path == "" {
		path = "/helios/config/cluster.hcl"
	}

	conf, err := model.ReadConfigFile(path)
	if err != nil {
		panic(err)
	}

	if ok, err := conf.IsValid(); !ok {
		panic(err)
	}

	pad := NewLaunchpad(conf)
	pad.InitCluster(ctx)
}
