package graph

import (
	"testing"

	"github.com/KyloRilo/helios/pkg/model/compute"
)

func makeNodes(names ...string) []*compute.Node {
	nodes := make([]*compute.Node, len(names))
	for i, name := range names {
		nodes[i] = compute.NewNode(compute.WithName(name))
	}
	return nodes
}

func TestNewClusterGraph(t *testing.T) {
	topo, err := New(map[string][]string{
		"db":  {},
		"api": {"db"},
	})
	if err != nil {
		t.Fatal(err)
	}

	cg, err := NewClusterGraph(topo, makeNodes("db", "api"))
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if cg.NodeCount() != 2 {
		t.Errorf("expected 2 nodes but got %d", cg.NodeCount())
	}
}

func TestNewClusterGraphNodeNotInTopology(t *testing.T) {
	topo, _ := New(map[string][]string{"db": {}})
	_, err := NewClusterGraph(topo, makeNodes("db", "extra"))
	if err == nil {
		t.Error("expected error for node not in topology")
	}
}

func TestNewClusterGraphTopologyNodeMissing(t *testing.T) {
	topo, _ := New(map[string][]string{"db": {}, "api": {"db"}})
	_, err := NewClusterGraph(topo, makeNodes("db"))
	if err == nil {
		t.Error("expected error for topology node with no compute node")
	}
}

func TestClusterGraphGet(t *testing.T) {
	topo, _ := New(map[string][]string{"db": {}, "api": {"db"}})
	cg, _ := NewClusterGraph(topo, makeNodes("db", "api"))

	n := cg.Get("db")
	if n == nil || n.Name != "db" {
		t.Errorf("expected db node but got %v", n)
	}

	if cg.Get("nonexistent") != nil {
		t.Error("expected nil for nonexistent node")
	}
}

func TestClusterGraphNodes(t *testing.T) {
	topo, _ := New(map[string][]string{"db": {}, "api": {"db"}, "web": {"api"}})
	cg, _ := NewClusterGraph(topo, makeNodes("db", "api", "web"))

	nodes := cg.Nodes()
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes but got %d", len(nodes))
	}
	// NodeNames returns sorted, so order is api, db, web
	if nodes[0].Name != "api" || nodes[1].Name != "db" || nodes[2].Name != "web" {
		t.Errorf("unexpected node order: %s, %s, %s", nodes[0].Name, nodes[1].Name, nodes[2].Name)
	}
}

func TestClusterGraphLevels(t *testing.T) {
	topo, _ := New(map[string][]string{
		"db":    {},
		"cache": {},
		"api":   {"db", "cache"},
		"web":   {"api"},
	})
	cg, _ := NewClusterGraph(topo, makeNodes("db", "cache", "api", "web"))

	levels, err := cg.Levels()
	if err != nil {
		t.Fatal(err)
	}

	if len(levels) != 3 {
		t.Fatalf("expected 3 levels but got %d", len(levels))
	}

	// Level 0: db, cache (no deps)
	if len(levels[0]) != 2 {
		t.Fatalf("expected 2 nodes in level 0 but got %d", len(levels[0]))
	}
	if levels[0][0].Name != "cache" || levels[0][1].Name != "db" {
		t.Errorf("level 0: expected [cache, db] but got [%s, %s]", levels[0][0].Name, levels[0][1].Name)
	}

	// Level 1: api
	if len(levels[1]) != 1 || levels[1][0].Name != "api" {
		t.Errorf("level 1: expected [api] but got %v", levels[1])
	}

	// Level 2: web
	if len(levels[2]) != 1 || levels[2][0].Name != "web" {
		t.Errorf("level 2: expected [web] but got %v", levels[2])
	}
}

func TestClusterGraphReverseLevels(t *testing.T) {
	topo, _ := New(map[string][]string{
		"db":  {},
		"api": {"db"},
		"web": {"api"},
	})
	cg, _ := NewClusterGraph(topo, makeNodes("db", "api", "web"))

	levels, err := cg.ReverseLevels()
	if err != nil {
		t.Fatal(err)
	}

	if len(levels) != 3 {
		t.Fatalf("expected 3 levels but got %d", len(levels))
	}

	// Reverse: web first (depends on nothing in reverse), then api, then db
	if levels[0][0].Name != "web" {
		t.Errorf("reverse level 0: expected web but got %s", levels[0][0].Name)
	}
	if levels[1][0].Name != "api" {
		t.Errorf("reverse level 1: expected api but got %s", levels[1][0].Name)
	}
	if levels[2][0].Name != "db" {
		t.Errorf("reverse level 2: expected db but got %s", levels[2][0].Name)
	}
}

func TestClusterGraphRoots(t *testing.T) {
	topo, _ := New(map[string][]string{
		"db":    {},
		"cache": {},
		"api":   {"db"},
	})
	cg, _ := NewClusterGraph(topo, makeNodes("db", "cache", "api"))

	roots := cg.Roots()
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots but got %d", len(roots))
	}
	if roots[0].Name != "cache" || roots[1].Name != "db" {
		t.Errorf("expected roots [cache, db] but got [%s, %s]", roots[0].Name, roots[1].Name)
	}
}

func TestClusterGraphDependents(t *testing.T) {
	topo, _ := New(map[string][]string{
		"db":  {},
		"api": {"db"},
		"web": {"db"},
	})
	cg, _ := NewClusterGraph(topo, makeNodes("db", "api", "web"))

	deps := cg.Dependents("db")
	if len(deps) != 2 {
		t.Fatalf("expected 2 dependents but got %d", len(deps))
	}
	if deps[0].Name != "api" || deps[1].Name != "web" {
		t.Errorf("expected dependents [api, web] but got [%s, %s]", deps[0].Name, deps[1].Name)
	}
}

func TestClusterGraphTopology(t *testing.T) {
	topo, _ := New(map[string][]string{"db": {}})
	cg, _ := NewClusterGraph(topo, makeNodes("db"))

	if cg.Topology() != topo {
		t.Error("expected Topology() to return the original graph")
	}
}

func TestClusterGraphSingleNode(t *testing.T) {
	topo, _ := New(map[string][]string{"solo": {}})
	cg, _ := NewClusterGraph(topo, makeNodes("solo"))

	levels, err := cg.Levels()
	if err != nil {
		t.Fatal(err)
	}
	if len(levels) != 1 || len(levels[0]) != 1 || levels[0][0].Name != "solo" {
		t.Errorf("expected single level with solo node")
	}

	roots := cg.Roots()
	if len(roots) != 1 || roots[0].Name != "solo" {
		t.Error("expected solo as only root")
	}

	deps := cg.Dependents("solo")
	if len(deps) != 0 {
		t.Error("expected no dependents for solo")
	}
}
