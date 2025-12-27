package model

import (
	"context"

	"github.com/asynkron/protoactor-go/actor"
)

type ActorService interface {
	GetServiceName() string
	Receive(actor.Context)
}

type ComputeController interface {
	Authenticate(context.Context) (interface{}, error)
	CreateNode(ctx context.Context, n *Node) (interface{}, error)
	StartNode(ctx context.Context, n *Node) (interface{}, error)
	StopNode(ctx context.Context, n *Node) (interface{}, error)
	RemoveNode(ctx context.Context, n *Node) (interface{}, error)
}
