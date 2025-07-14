package core

import (
	"fmt"
	"log"

	"github.com/KyloRilo/helios/pkg/controller/docker"
	"github.com/KyloRilo/helios/pkg/model"

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
	dockerCtrl docker.DockerController
}

func (serv CoreService) Receive(ctx actor.Context) {
	var err error
	msg := ctx.Message()
	log.Println("CoreService.MsgHandler() => Receive: ", msg)

	switch req := msg.(type) {
	case SpinupNode:
		err = serv.spinupNode(req.conf)
	case TeardownNode:
		err = serv.teardownNode(req.instanceId)
	default:
		err = fmt.Errorf("CoreService.MsgHandler() => Unhandled Message case")
	}

	ctx.Respond(err)
}

func (serv CoreService) spinupNode(_ model.InstanceConfig) error {
	fmt.Printf("Received spinup event")
	return nil
}

func (serv CoreService) teardownNode(_ string) error {
	return nil
}

func (serv CoreService) healthcheck() {}

func InitCoreService() CoreService {
	return CoreService{
		dockerCtrl: docker.InitDockerController(),
	}
}
