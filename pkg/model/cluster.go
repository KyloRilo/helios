package model

import (
	"context"
	"fmt"
)

type HeliosCluster struct {
	conf  HCluster
	graph Graph
}

func NewHeliosCluster(conf HCluster) *HeliosCluster {
	return &HeliosCluster{
		conf:  conf,
		graph: NewGraph(),
	}
}

func (hc *HeliosCluster) Init(ctxGen func(svc HService) ExecCtx) {
	for _, svc := range hc.conf.Services {
		node := NewNode(NodeMeta(svc))
		node.ExecCtx = ctxGen(svc)
		hc.graph.AddNode(node)
		for _, dep := range node.Meta.DependsOn {
			hc.graph.AddDependency(node.Meta.Name, dep)
		}
	}
}

func (hc *HeliosCluster) Validate() error {
	if hc.graph.IsEmpty() {
		return fmt.Errorf("Cluster graph is empty. Initialize cluster before proceeding")
	}

	return nil
}

func (hc *HeliosCluster) Create(ctx context.Context) error {
	return hc.graph.ExecLevels(ctx, func(ctx ExecCtx) ExecFunc {
		return ctx.Create
	})
}

func (hc *HeliosCluster) Start(ctx context.Context) error {
	return hc.graph.ExecLevels(ctx, func(ctx ExecCtx) ExecFunc {
		return ctx.Start
	})
}

func (hc *HeliosCluster) Stop(ctx context.Context) error {
	return hc.graph.ExecLevels(ctx, func(ctx ExecCtx) ExecFunc {
		return ctx.Stop
	})
}

func (hc *HeliosCluster) Teardown(ctx context.Context) error {
	return hc.graph.ExecLevels(ctx, func(ctx ExecCtx) ExecFunc {
		return ctx.Delete
	})
}
