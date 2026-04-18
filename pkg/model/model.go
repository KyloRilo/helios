package model

import (
	"github.com/asynkron/protoactor-go/actor"
)

type ActorService interface {
	GetServiceName() string
	Receive(actor.Context)
}

type BaseService struct {
	name string
	ctx  actor.Context
}

func (bs BaseService) GetServiceName() string {
	return bs.name
}

func (bs BaseService) Receive(actx actor.Context) {
	// Default no-op
}

func NewBaseActorService(name string) ActorService {
	return BaseService{
		name: name,
	}
}

type CloudProvider string

const (
	GCP CloudProvider = "gcp"
	AWS CloudProvider = "aws"
)
