package graph

import (
	"fmt"
	"sort"

	"github.com/KyloRilo/helios/pkg/model/compute"
)

type ClusterGraph struct {
	topology *Graph
	nodes    map[string]*compute.Node
}

func NewClusterGraph(topo *Graph, nodeList []*compute.Node) (*ClusterGraph, error) {
	nodes := make(map[string]*compute.Node, len(nodeList))
	for _, n := range nodeList {
		if !topo.HasNode(n.Name) {
			return nil, fmt.Errorf("node %q not found in topology", n.Name)
		}
		nodes[n.Name] = n
	}

	for _, name := range topo.NodeNames() {
		if _, ok := nodes[name]; !ok {
			return nil, fmt.Errorf("topology node %q has no matching compute node", name)
		}
	}

	return &ClusterGraph{topology: topo, nodes: nodes}, nil
}

func (cg *ClusterGraph) Get(name string) *compute.Node {
	return cg.nodes[name]
}

func (cg *ClusterGraph) Topology() *Graph {
	return cg.topology
}

func (cg *ClusterGraph) Nodes() []*compute.Node {
	names := cg.topology.NodeNames()
	result := make([]*compute.Node, 0, len(names))
	for _, name := range names {
		result = append(result, cg.nodes[name])
	}
	return result
}

func (cg *ClusterGraph) Levels() ([][]*compute.Node, error) {
	stringLevels, err := cg.topology.Levels()
	if err != nil {
		return nil, err
	}
	return cg.mapLevels(stringLevels), nil
}

func (cg *ClusterGraph) ReverseLevels() ([][]*compute.Node, error) {
	rev := cg.topology.Reverse()
	stringLevels, err := rev.Levels()
	if err != nil {
		return nil, err
	}
	return cg.mapLevels(stringLevels), nil
}

func (cg *ClusterGraph) Roots() []*compute.Node {
	names := cg.topology.Roots()
	result := make([]*compute.Node, 0, len(names))
	for _, name := range names {
		result = append(result, cg.nodes[name])
	}
	return result
}

func (cg *ClusterGraph) Dependents(name string) []*compute.Node {
	depNames := cg.topology.Dependents(name)
	result := make([]*compute.Node, 0, len(depNames))
	for _, n := range depNames {
		result = append(result, cg.nodes[n])
	}
	return result
}

func (cg *ClusterGraph) NodeCount() int {
	return len(cg.nodes)
}

func (cg *ClusterGraph) mapLevels(stringLevels [][]string) [][]*compute.Node {
	levels := make([][]*compute.Node, len(stringLevels))
	for i, sl := range stringLevels {
		level := make([]*compute.Node, len(sl))
		for j, name := range sl {
			level[j] = cg.nodes[name]
		}
		sort.Slice(level, func(a, b int) bool { return level[a].Name < level[b].Name })
		levels[i] = level
	}
	return levels
}
