package graph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type NodeType string
type NodeStatus int

const (
	NodeStatusUndefined NodeStatus = iota
	NodeStatusReady
	NodeStatusPending
	NodeStatusFinished
	NodeStatusError
)

type ExecFunc func(ctx context.Context, meta interface{}) (map[string]interface{}, error)
type NodeId string
type NodeV2 interface {
	GetId() NodeId
	GetStatus() NodeStatus
	GetType() NodeType
	GetDependsOn() []string
	GetName() string
	GetMeta() map[string]interface{}
	GetOutputs() map[string]interface{}
	SetOutputs(outputs map[string]interface{})
	SetStatus(status NodeStatus)
}

type GraphStatus int

const (
	GraphStatusUndefined GraphStatus = iota
	GraphStatusReady
	GraphStatusPending
)

type GraphV2 struct {
	nodes  map[NodeId]NodeV2
	refMap map[string]NodeId
	edges  map[NodeId][]NodeId
	status GraphStatus
}

func NewGraphV2() GraphV2 {
	return GraphV2{
		nodes:  make(map[NodeId]NodeV2),
		edges:  make(map[NodeId][]NodeId),
		refMap: make(map[string]NodeId),
		status: GraphStatusUndefined,
	}
}

func (g *GraphV2) GetStatus() GraphStatus {
	return g.status
}

func (g *GraphV2) SetStatus(status GraphStatus) {
	g.status = status
}

func (g *GraphV2) AddNode(n NodeV2) error {
	if _, exists := g.refMap[n.GetName()]; !exists {
		g.nodes[n.GetId()] = n
		g.refMap[n.GetName()] = n.GetId()
	} else {
		return fmt.Errorf("Unable to add node => Node at key '%s' already exists in graph", n.GetName())
	}

	return nil
}

func (g *GraphV2) AddEdge(from, to NodeId) {
	g.edges[from] = append(g.edges[from], to)
}

func (g *GraphV2) BuildDependencyGraph() error {
	var errors []error = []error{}
	for _, n := range g.nodes {
		deps := n.GetDependsOn()
		for _, dep := range deps {
			var id NodeId
			if depId, exists := g.refMap[dep]; !exists {
				errors = append(errors, fmt.Errorf("Unable to build dependency graph => Node '%s' depends on unknown node '%s'", n.GetName(), dep))
			} else {
				id = depId
				g.AddEdge(n.GetId(), id)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("Errors occurred building dependency graph: %v", errors)
	}

	return nil
}

func (g *GraphV2) GetNode(nodeId NodeId) (NodeV2, error) {
	if node, exists := g.nodes[nodeId]; exists {
		return node, nil
	} else {
		return nil, fmt.Errorf("Node with id '%s' not found", nodeId)
	}
}

func (g *GraphV2) GetNodeByRef(ref string) (NodeV2, error) {
	if id, exists := g.refMap[ref]; exists {
		return g.GetNode(id)
	} else {
		return nil, fmt.Errorf("Node with ref '%s' not found", ref)
	}
}

func (g *GraphV2) GetDependencies(id NodeId) []NodeId {
	var deps []NodeId
	for from, tos := range g.edges {
		for _, to := range tos {
			if to == id {
				deps = append(deps, from)
			}
		}
	}

	return deps
}

func (g *GraphV2) GetDependents(id NodeId) []NodeId {
	return g.edges[id]
}

func (g *GraphV2) IsEmpty() bool {
	return len(g.nodes) == 0
}

func (g *GraphV2) Validate() error {
	if g.IsEmpty() {
		return fmt.Errorf("Graph is empty")
	}

	notReady := make(map[NodeId]NodeStatus, 0)
	for id, node := range g.nodes {
		if node.GetStatus() != NodeStatusReady {
			notReady[id] = node.GetStatus()
		}
	}

	if len(notReady) > 0 {
		return fmt.Errorf("Graph unready for execution: %v", notReady)
	}

	// TODO: Add cycle detection
	return nil
}

func (g *GraphV2) Levels() ([][]NodeId, error) {
	return nil, nil
}

type BaseNode struct {
	id        NodeId `json:"id"`
	name      string `json:"name"`
	nodeType  NodeType
	status    NodeStatus
	dependsOn []string               `json:"depends_on"`
	meta      map[string]interface{} `json:"meta"`
	outputs   map[string]interface{} `json:"outputs"`
}

func (n BaseNode) GetId() NodeId                             { return n.id }
func (n BaseNode) GetType() NodeType                         { return n.nodeType }
func (n BaseNode) GetDependsOn() []string                    { return n.dependsOn }
func (n BaseNode) GetName() string                           { return n.name }
func (n BaseNode) GetMeta() map[string]interface{}           { return n.meta }
func (n BaseNode) GetOutputs() map[string]interface{}        { return n.outputs }
func (n BaseNode) SetOutputs(outputs map[string]interface{}) { n.outputs = outputs }
func (n BaseNode) GetStatus() NodeStatus                     { return n.status }
func (n BaseNode) SetStatus(status NodeStatus)               { n.status = status }

func mergeMeta(metaMaps ...map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for _, m := range metaMaps {
		for k, v := range m {
			merged[k] = v
		}
	}

	return merged
}

func NewBasicNode(nodeType NodeType, name string, dependsOn []string, meta map[string]interface{}) NodeV2 {
	id := uuid.New().String()
	status := NodeStatusUndefined
	return BaseNode{
		id:        NodeId(id),
		name:      name,
		nodeType:  nodeType,
		status:    status,
		dependsOn: dependsOn,
		meta:      meta,
	}
}
