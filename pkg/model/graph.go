package model

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type NodeStatus int

const (
	NodeStateUndefined NodeStatus = iota
	NodeStateCreating
	NodeStateRunning
	NodeStateStopped
	NodeStateError
)

type NodeState struct {
	Status NodeStatus
}

type NodeMeta HService
type ExecFunc func(ctx context.Context, n *Node) (interface{}, error)
type ExecCtx struct {
	Create ExecFunc
	Start  ExecFunc
	Read   ExecFunc
	Update ExecFunc
	Stop   ExecFunc
	Delete ExecFunc
}

type Node struct {
	Meta    NodeMeta
	State   NodeState
	ExecCtx ExecCtx
}

func NewNode(meta NodeMeta) *Node {
	return &Node{
		Meta: meta,
	}
}

type Graph struct {
	nodes        map[string]*Node
	dependencies map[string][]string
	dependents   map[string][]string
}

func NewGraph() Graph {
	return Graph{
		nodes:        map[string]*Node{},
		dependencies: map[string][]string{},
		dependents:   map[string][]string{},
	}
}

func (g *Graph) IsEmpty() bool {
	return len(g.nodes) == 0
}

func (g *Graph) AddNode(n *Node) {
	name := n.Meta.Name
	g.nodes[name] = n
	if _, ok := g.dependencies[name]; !ok {
		g.dependencies[name] = []string{}
	}
	if _, ok := g.dependents[name]; !ok {
		g.dependents[name] = []string{}
	}
}

func (g *Graph) GetNode(nodeKey string) *Node {
	return g.nodes[nodeKey]
}

func (g *Graph) AddDependency(nodeKey, dependsKey string) {
	g.dependencies[nodeKey] = append(g.dependencies[nodeKey], dependsKey)
	g.dependents[dependsKey] = append(g.dependents[dependsKey], nodeKey)
}

func (g *Graph) Levels() ([][]string, error) {
	// indegree counts how many dependencies remain for each node
	indegree := make(map[string]int)
	for key := range g.nodes {
		indegree[key] = len(g.dependencies[key])
	}

	// initialize queue with nodes that have indegree 0
	var zero []string
	for key, d := range indegree {
		if d == 0 {
			zero = append(zero, key)
		}
	}

	var levels [][]string
	processed := 0
	for len(zero) > 0 {
		// snapshot current zero-degree nodes as a level (they can run in parallel)
		level := make([]string, len(zero))
		copy(level, zero)
		levels = append(levels, level)

		nextZero := []string{}
		// for each node in this level, decrement indegree of its dependents
		for _, n := range zero {
			processed++
			for _, dep := range g.dependents[n] {
				indegree[dep]--
				if indegree[dep] == 0 {
					nextZero = append(nextZero, dep)
				}
			}
		}

		zero = nextZero
	}

	if processed != len(g.nodes) {
		return nil, errors.New("dependency cycle detected")
	}
	return levels, nil
}

func (g *Graph) ExecLevels(ctx context.Context, genFunc func(ExecCtx) ExecFunc) error {
	var levelErr error
	var group sync.WaitGroup
	defer group.Done()

	levels, err := g.Levels()
	if err != nil {
		return err
	}

	for _, level := range levels {
		errs := make(chan error, len(level))
		for _, id := range level {
			node := g.GetNode(id)
			execFunc := genFunc(node.ExecCtx)
			group.Add(1)
			if _, err := execFunc(ctx, node); err != nil {
				return fmt.Errorf("Node '%s': %w", node.Meta.Name, err)
			}
		}

		group.Wait()
		close(errs)
		for e := range errs {
			levelErr = fmt.Errorf("%v; %w", levelErr, e)
		}

		if levelErr != nil {
			return levelErr
		}
	}

	return nil
}
