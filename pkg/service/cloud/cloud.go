package cloud

import (
	"github.com/asynkron/protoactor-go/actor"
)

type CloudComputeService struct {
	actor.Actor
}

func (serv CloudComputeService) Receive(actor.Actor) {}
