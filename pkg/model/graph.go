package model

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type NodeMeta HService
type NodeExec func(ctx context.Context, n *Node) error
type Node struct {
	Meta NodeMeta
	Exec NodeExec
}

type Graph struct {
	nodes        map[string]*Node
	dependencies map[string][]string
	dependents   map[string][]string
}

func NewGraph(cfg *HCluster) *Graph {
	graph := &Graph{
		nodes:        map[string]*Node{},
		dependencies: map[string][]string{},
		dependents:   map[string][]string{},
	}

	if cfg != nil {
		for _, srvc := range cfg.Services {
			graph.AddNode(&Node{
				Meta: NodeMeta(srvc),
			})

			for _, dep := range srvc.DependsOn {
				graph.AddDependency(srvc.Name, dep)
			}
		}
	}

	return graph
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

func (g *Graph) UpdateLevels(execFunc NodeExec) error {
	if g.IsEmpty() {
		return errors.New("Unable to update, graph found empty")
	}

	levels, err := g.Levels()
	if err != nil {
		return err
	}

	for _, level := range levels {
		for _, id := range level {
			node := g.nodes[id]
			node.Exec = execFunc
		}
	}

	return nil
}

func (g *Graph) ExecLevels(ctx context.Context) error {
	var group sync.WaitGroup
	var levelErr error
	levels, err := g.Levels()
	if err != nil {
		return err
	}

	for i, level := range levels {
		fmt.Printf("=== executing level %d (parallel batch size=%d): %v\n", i, len(level), level)
		errs := make(chan error, len(level))
		for _, id := range level {
			node := g.nodes[id]
			group.Add(1)
			go func() {
				defer group.Done()
				if err := node.Exec(ctx, node); err != nil {
					errs <- fmt.Errorf("Node '%s': %w", node.Meta.Name, err)
				}
			}()
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
