package core

import (
	"context"
	"fmt"
	"log"

	"github.com/KyloRilo/helios/pkg/controller/docker"
	"github.com/KyloRilo/helios/pkg/model"

	"github.com/KyloRilo/helios/pkg/service/raft"
	"github.com/asynkron/protoactor-go/actor"
)

type SpinupNode struct {
	conf model.InstanceConfig
}

type TeardownNode struct {
	instanceId string
}

type CoreService struct {
	actor.Actor
	raft   raft.RaftService
	Docker docker.DockerCtrl
}

func (core CoreService) GetServiceName() string {
	return "core-service"
}

func (core CoreService) Receive(ctx actor.Context) {
	var err error
	if core.raft.IsLeader() {
		msg := ctx.Message()
		log.Println("CoreService.MsgHandler() => Receive: ", msg)

		switch req := msg.(type) {
		case SpinupNode:
			err = core.spinupNode(req.conf)
		case TeardownNode:
			err = core.teardownNode(req.instanceId)
		default:
			err = fmt.Errorf("CoreService.MsgHandler() => Unhandled Message case")
		}

		ctx.Respond(err)
	}
}

func (core CoreService) spinupNode(_ model.InstanceConfig) error {
	fmt.Printf("Received spinup event")
	return nil
}

func (core CoreService) teardownNode(_ string) error {
	return nil
}

func (core CoreService) healthcheck() {}

func (core *CoreService) CreateService(ctx context.Context, srv model.Service) error {
	meta, err := core.Docker.Create(ctx, model.ContainerConfig{
		Name:  srv.Name,
		Image: srv.Image,
		Ports: srv.Ports,
	})
	if err != nil {
		return fmt.Errorf("Failed to Create container for service '%s' => %s", srv.Name, err)
	}

	fmt.Printf("Created Container '%s' running with ID '%s':", meta.Name, meta.Id)
	fmt.Printf("Running at '%s' with ports '%v'", meta.Hostname, meta.Ports)
	return nil
}

func (core *CoreService) DestroyService(srvName string) error {
	return nil
}

func InitCoreService(cfg *model.ClusterConfig) model.ActorService {
	core := CoreService{
		Docker: docker.InitDockerController(),
	}

	return core
}
