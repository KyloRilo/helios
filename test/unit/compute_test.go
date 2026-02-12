package unit

import (
	"context"
	"fmt"

	"github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
)

type stub interface {
	compute.ComputeController
}

type CompStub struct {
	auth       func(context.Context) (interface{}, error)
	createNode func(context.Context, *model.Node) (interface{}, error)
	startNode  func(context.Context, *model.Node) (interface{}, error)
	stopNode   func(context.Context, *model.Node) (interface{}, error)
	removeNode func(context.Context, *model.Node) (interface{}, error)
}

func (c CompStub) Authenticate(ctx context.Context) (interface{}, error) {
	return c.auth(ctx)
}

func (c CompStub) CreateNode(ctx context.Context, node *model.Node) (interface{}, error) {
	return c.createNode(ctx, node)
}

func (c CompStub) StartNode(ctx context.Context, node *model.Node) (interface{}, error) {
	return c.startNode(ctx, node)
}

func (c CompStub) StopNode(ctx context.Context, node *model.Node) (interface{}, error) {
	return c.stopNode(ctx, node)
}

func (c CompStub) RemoveNode(ctx context.Context, node *model.Node) (interface{}, error) {
	return c.removeNode(ctx, node)
}

type CreateSuccess struct{ CompStub }

func (c CreateSuccess) CreateNode(_ context.Context, _ *model.Node) (interface{}, error) {
	return nil, nil
}

type CreateFailed struct{ CompStub }

func (c CreateFailed) CreateNode(_ context.Context, _ *model.Node) (interface{}, error) {
	return nil, fmt.Errorf("failed to create container")
}

type StartFailed struct{ CompStub }

func (c StartFailed) StartNode(_ context.Context, _ *model.Node) (interface{}, error) {
	return nil, fmt.Errorf("failed to start container")
}

type StopFailed struct{ CompStub }

func (c StopFailed) StopNode(_ context.Context, _ *model.Node) (interface{}, error) {
	return nil, fmt.Errorf("failed to stop container")
}

type RemoveFailed struct{ CompStub }

func (c RemoveFailed) RemoveNode(ctx context.Context, _ *model.Node) (interface{}, error) {
	return nil, fmt.Errorf("failed to remove container")
}
