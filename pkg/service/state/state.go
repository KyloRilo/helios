package state

import (
	"log"

	"github.com/KyloRilo/helios/pkg/service/cloud"
	"github.com/asynkron/protoactor-go/actor"
)

type GetState struct{}

type StateService struct {
	actor.Actor
	storageCtrl cloud.ICloudStoreService
	state       map[string]interface{}
}

func (serv StateService) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {
	case GetState:
		break
	}
}

func InitStateService() StateService {
	ctrl, err := cloud.InitCloudStorageService(cloud.NewCloudConfig())
	if err != nil {
		log.Fatal("Unable to init state service => ", err)
	}

	return StateService{
		storageCtrl: ctrl,
	}
}
