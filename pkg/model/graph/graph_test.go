package graph

import (
	"errors"
	"reflect"
	"sort"
	"testing"
)

func realWorldDeps() map[string][]string {
	return map[string][]string{
		"consul":     {},
		"localstack": {},
		"core":       {"consul", "localstack"},
		"worker":     {"consul", "core"},
		"api":        {"consul", "core"},
	}
}

func TestNewValidGraph(t *testing.T) {
	g, err := New(realWorldDeps())
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if g == nil {
		t.Fatal("expected graph but got nil")
	}
}

func TestNewMissingDependency(t *testing.T) {
	_, err := New(map[string][]string{
		"A": {"B"},
	})
	if err == nil {
		t.Fatal("expected error for missing dependency")
	}

	var missing *MissingDepError
	if !errors.As(err, &missing) {
		t.Fatalf("expected MissingDepError but got %T: %v", err, err)
	}
	if missing.Node != "A" || missing.Dependency != "B" {
		t.Errorf("expected A depends on B, got %+v", missing)
	}
}

func TestTopoSort(t *testing.T) {
	g, err := New(realWorldDeps())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	order, err := g.TopoSort()
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if len(order) != 5 {
		t.Fatalf("expected 5 nodes but got %d: %v", len(order), order)
	}

	pos := make(map[string]int, len(order))
	for i, name := range order {
		pos[name] = i
	}

	// consul and localstack must come before core
	if pos["consul"] >= pos["core"] {
		t.Errorf("consul must come before core: %v", order)
	}
	if pos["localstack"] >= pos["core"] {
		t.Errorf("localstack must come before core: %v", order)
	}
	// core must come before worker and api
	if pos["core"] >= pos["worker"] {
		t.Errorf("core must come before worker: %v", order)
	}
	if pos["core"] >= pos["api"] {
		t.Errorf("core must come before api: %v", order)
	}
	// consul must come before worker and api
	if pos["consul"] >= pos["worker"] {
		t.Errorf("consul must come before worker: %v", order)
	}
	if pos["consul"] >= pos["api"] {
		t.Errorf("consul must come before api: %v", order)
	}
}

func TestTopoSortCycle(t *testing.T) {
	g, err := New(map[string][]string{
		"A": {"B"},
		"B": {"A"},
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %v", err)
	}

	_, err = g.TopoSort()
	if err == nil {
		t.Fatal("expected cycle error")
	}

	var cycleErr *CycleError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("expected CycleError but got %T: %v", err, err)
	}
	sort.Strings(cycleErr.Nodes)
	if !reflect.DeepEqual(cycleErr.Nodes, []string{"A", "B"}) {
		t.Errorf("expected cycle nodes [A, B] but got %v", cycleErr.Nodes)
	}
}

func TestLevels(t *testing.T) {
	g, err := New(realWorldDeps())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	levels, err := g.Levels()
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if len(levels) != 3 {
		t.Fatalf("expected 3 levels but got %d: %v", len(levels), levels)
	}

	// Level 0: consul, localstack (sorted)
	if !reflect.DeepEqual(levels[0], []string{"consul", "localstack"}) {
		t.Errorf("expected level 0 [consul, localstack] but got %v", levels[0])
	}

	// Level 1: core
	if !reflect.DeepEqual(levels[1], []string{"core"}) {
		t.Errorf("expected level 1 [core] but got %v", levels[1])
	}

	// Level 2: api, worker (sorted)
	if !reflect.DeepEqual(levels[2], []string{"api", "worker"}) {
		t.Errorf("expected level 2 [api, worker] but got %v", levels[2])
	}
}

func TestLevelsCycle(t *testing.T) {
	g, _ := New(map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"A"},
	})

	_, err := g.Levels()
	if err == nil {
		t.Fatal("expected cycle error")
	}

	var cycleErr *CycleError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("expected CycleError but got %T: %v", err, err)
	}
}

func TestReverse(t *testing.T) {
	g, err := New(realWorldDeps())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rev := g.Reverse()

	// Teardown order: dependents before dependencies
	levels, err := rev.Levels()
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if len(levels) != 3 {
		t.Fatalf("expected 3 levels but got %d: %v", len(levels), levels)
	}

	// Level 0: api, worker (no dependents in reversed graph)
	if !reflect.DeepEqual(levels[0], []string{"api", "worker"}) {
		t.Errorf("expected reverse level 0 [api, worker] but got %v", levels[0])
	}

	// Level 1: core
	if !reflect.DeepEqual(levels[1], []string{"core"}) {
		t.Errorf("expected reverse level 1 [core] but got %v", levels[1])
	}

	// Level 2: consul, localstack
	if !reflect.DeepEqual(levels[2], []string{"consul", "localstack"}) {
		t.Errorf("expected reverse level 2 [consul, localstack] but got %v", levels[2])
	}
}

func TestRoots(t *testing.T) {
	g, _ := New(realWorldDeps())
	roots := g.Roots()
	if !reflect.DeepEqual(roots, []string{"consul", "localstack"}) {
		t.Errorf("expected roots [consul, localstack] but got %v", roots)
	}
}

func TestDependents(t *testing.T) {
	g, _ := New(realWorldDeps())

	deps := g.Dependents("consul")
	if !reflect.DeepEqual(deps, []string{"api", "core", "worker"}) {
		t.Errorf("expected dependents of consul [api, core, worker] but got %v", deps)
	}

	deps = g.Dependents("core")
	if !reflect.DeepEqual(deps, []string{"api", "worker"}) {
		t.Errorf("expected dependents of core [api, worker] but got %v", deps)
	}

	deps = g.Dependents("worker")
	if len(deps) != 0 {
		t.Errorf("expected no dependents of worker but got %v", deps)
	}
}

func TestSingleNode(t *testing.T) {
	g, err := New(map[string][]string{"solo": {}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	order, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(order, []string{"solo"}) {
		t.Errorf("expected [solo] but got %v", order)
	}

	levels, err := g.Levels()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(levels) != 1 || !reflect.DeepEqual(levels[0], []string{"solo"}) {
		t.Errorf("expected [[solo]] but got %v", levels)
	}
}

func TestEmptyGraph(t *testing.T) {
	g, err := New(map[string][]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	order, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 0 {
		t.Errorf("expected empty order but got %v", order)
	}

	levels, err := g.Levels()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(levels) != 0 {
		t.Errorf("expected empty levels but got %v", levels)
	}

	roots := g.Roots()
	if len(roots) != 0 {
		t.Errorf("expected empty roots but got %v", roots)
	}
}

func TestValidate(t *testing.T) {
	g, _ := New(map[string][]string{"A": {"B"}, "B": {"A"}})
	err := g.Validate()
	if err == nil {
		t.Error("expected validation error for cycle")
	}

	g, _ = New(realWorldDeps())
	err = g.Validate()
	if err != nil {
		t.Errorf("expected valid graph but got %v", err)
	}
}
