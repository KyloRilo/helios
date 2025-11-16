package worker

import (
	"fmt"
	"log"

	"github.com/KyloRilo/helios/pkg/controller/docker"
	"github.com/asynkron/protoactor-go/actor"
)

type SpinupContainer struct {
	image string
}

type GetContainer struct {
	containerId string
}

type DestroyContainer GetContainer

type WorkerService struct {
	actor.Actor
	dockerCtrl docker.DockerCtrl
}

func (serv WorkerService) Receive(ctx actor.Context) {
	var err error
	msg := ctx.Message()

	log.Println("WorkerService.MsgHandler() => Receive: ", msg)
	switch req := msg.(type) {
	case SpinupContainer:
		_, err = serv.spinupContainer(req.image)
	case DestroyContainer:
		err = serv.destroyContainer(req.containerId)
	case GetContainer:
		break
	default:
		err = fmt.Errorf("WorkerService.MsgHandler() => Unhandled Message case")
	}

	if err != nil {
		ctx.Respond(err)
	}

}

func (serv WorkerService) spinupContainer(image string) (string, error) {
	return "", nil
}

func (serv WorkerService) destroyContainer(containerId string) error {
	return nil
}

func InitWorkerService() actor.Actor {
	return &WorkerService{
		dockerCtrl: docker.InitDockerController(),
	}
}
