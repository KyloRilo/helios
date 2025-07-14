package state

import (
	"log"

	"github.com/KyloRilo/helios/pkg/controller/cloud"
	"github.com/asynkron/protoactor-go/actor"
)

type GetState struct{}

type StateService struct {
	actor.Actor
	storageCtrl cloud.ICloudStoreCtrl
	state       map[string]interface{}
}

func (serv StateService) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {
	case GetState:
		break
	}
}

func InitStateService() StateService {
	ctrl, err := cloud.InitCloudStorageCtrl(cloud.CloudConfig{})
	if err != nil {
		log.Fatal("Unable to ini[t state service => ", err)
	}

	return StateService{
		storageCtrl: ctrl,
	}
}
