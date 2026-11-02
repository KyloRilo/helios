package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/KyloRilo/helios/pkg/model/compute"
)

type NodeState struct {
	Name      string         `json:"name"`
	Id        string         `json:"id"`
	Status    compute.Status `json:"status"`
	DependsOn []string       `json:"depends_on,omitempty"`
}

type ClusterState struct {
	Cluster  string      `json:"cluster"`
	Network  string      `json:"network"`
	Manifest string      `json:"manifest"`
	Nodes    []NodeState `json:"nodes"`
}

func (s *ClusterState) UpdateNodes(nodes []*compute.Node) {
	s.Nodes = make([]NodeState, len(nodes))
	for i, n := range nodes {
		s.Nodes[i] = NodeState{
			Name:      n.Name,
			Id:        n.Id,
			Status:    n.Status,
			DependsOn: n.DependsOn,
		}
	}
}

func stateDir() string {
	return ".helios"
}

func StatePath(clusterName string) string {
	return filepath.Join(stateDir(), clusterName+".state.json")
}

func (s *ClusterState) Save() error {
	if err := os.MkdirAll(stateDir(), 0755); err != nil {
		return fmt.Errorf("creating state dir: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	path := StatePath(s.Cluster)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	return nil
}

func LoadState(clusterName string) (*ClusterState, error) {
	path := StatePath(clusterName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	state := &ClusterState{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}

	return state, nil
}

func RemoveState(clusterName string) error {
	return os.Remove(StatePath(clusterName))
}
