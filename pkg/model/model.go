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
	Authenticate(context.Context) error
	CreateNode(ctx context.Context, n *Node) error
	StartNode(ctx context.Context, n *Node) error
	StopNode(ctx context.Context, n *Node) error
	RemoveNode(ctx context.Context, n *Node) error
}
