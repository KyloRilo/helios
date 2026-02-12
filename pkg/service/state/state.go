package state

import (
	"log"

	"github.com/KyloRilo/helios/pkg/controller/storage"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/model/graph"
	"github.com/asynkron/protoactor-go/actor"
)

// Currently only supports services, may expand
const (
	ServiceNodeType graph.NodeType = "service"
)

type StateService struct {
	model.ActorService
	storageCtrl  storage.StorageService
	serviceGraph graph.GraphV2
	services     map[string]graph.NodeV2
	state        map[string]interface{}
}

// func (s *StateService) newServiceNode(svc model.HService) error {
// 	meta := map[string]interface{}{}
// 	if svc.Command != "" {
// 		meta["command"] = svc.Command
// 	}

// 	if svc.Image != "" {
// 		meta["image"] = svc.Image
// 	}

// 	if len(svc.Volumes) > 0 {
// 		meta["volumes"] = svc.Volumes
// 	}

// 	if len(svc.Ports) > 0 {
// 		meta["ports"] = svc.Ports
// 	}

// 	if len(svc.Environment) > 0 {
// 		meta["environment"] = svc.Environment
// 	}

// 	if svc.Build != nil {
// 		meta["build"] = map[string]string{
// 			"context":    svc.Build.Context,
// 			"dockerfile": svc.Build.Dockerfile,
// 		}
// 	}

// 	node := graph.NewBasicNode(ServiceNodeType, svc.Name, svc.Type, svc.DependsOn, meta)
// 	err := s.serviceGraph.AddNode(node)
// 	if err != nil {
// 		return err
// 	}

// 	s.services[node.GetName()] = node
// 	node.SetStatus(graph.NodeStatusReady)
// 	return nil
// }

// func (s *StateService) InitServiceGraph() error {
// 	var err error
// 	s.serviceGraph.SetStatus(graph.GraphStatusPending)
// 	for _, svc := range s.services {
// 		err := s.newServiceNode(svc)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	err = s.serviceGraph.BuildDependencyGraph()
// 	if err != nil {
// 		return fmt.Errorf("Unable to build dependency graph => %s", err)
// 	}

// 	err = s.serviceGraph.Validate()
// 	if err != nil {
// 		return fmt.Errorf("Cluster found in invalid state => %s")
// 	}

// 	s.serviceGraph.SetStatus(graph.GraphStatusReady)
// 	return nil
// }

type GetState struct{}

func (serv StateService) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {
	case GetState:
		break
	}
}

func InitStateService() StateService {
	ctrl, err := storage.InitStorageController()
	if err != nil {
		log.Fatal("Unable to init state service => ", err)
	}

	return StateService{
		ActorService: model.NewBaseActorService("State"),
		serviceGraph: graph.NewGraphV2(),
		storageCtrl:  ctrl,
	}
}
