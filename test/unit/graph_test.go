package unit

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KyloRilo/helios/pkg/model"
)

// helper to make a node with a mock Create() function
func mockNode(id string, fn func() error) *model.Node {
	return &model.Node{
		Meta: model.NodeMeta{
			Name: id,
			ID:   id,
		},
		Exec: func(ctx context.Context, n *model.Node) error {
			if fn != nil {
				return fn()
			}
			return nil
		},
	}
}

func genExec() model.NodeExec {
	return func(ctx context.Context, n *model.Node) error {
		return n.Exec(ctx, n)
	}
}

func TestLevelsSimple(t *testing.T) {
	g := model.NewGraph(nil)

	// A → B → C
	g.AddNode(mockNode("A", nil))
	g.AddNode(mockNode("B", nil))
	g.AddNode(mockNode("C", nil))

	g.AddDependency("B", "A")
	g.AddDependency("C", "B")

	levels, err := g.Levels()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect: [[A], [B], [C]]
	if len(levels) != 3 {
		t.Fatalf("expected 3 levels, got %d: %v", len(levels), levels)
	}

	expected := [][]string{
		{"A"},
		{"B"},
		{"C"},
	}

	for i := range expected {
		if len(levels[i]) != 1 || levels[i][0] != expected[i][0] {
			t.Fatalf("expected %v, got %v", expected, levels)
		}
	}
}

func TestLevelsParallelBatch(t *testing.T) {
	g := model.NewGraph(nil)

	//   A
	//  / \
	// B   C
	//  \ /
	//   D

	g.AddNode(mockNode("A", nil))
	g.AddNode(mockNode("B", nil))
	g.AddNode(mockNode("C", nil))
	g.AddNode(mockNode("D", nil))

	g.AddDependency("B", "A")
	g.AddDependency("C", "A")
	g.AddDependency("D", "B")
	g.AddDependency("D", "C")

	levels, err := g.Levels()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect A first, then B and C in same batch, then D
	if len(levels) != 3 {
		t.Fatalf("expected 3 levels, got %d: %v", len(levels), levels)
	}

	if !(len(levels[1]) == 2 &&
		((levels[1][0] == "B" && levels[1][1] == "C") ||
			(levels[1][0] == "C" && levels[1][1] == "B"))) {
		t.Fatalf("expected B and C in same batch: got %v", levels)
	}
}

func TestLevelsDetectsCycle(t *testing.T) {
	g := model.NewGraph(nil)

	g.AddNode(mockNode("A", nil))
	g.AddNode(mockNode("B", nil))

	// A → B → A (cycle)
	g.AddDependency("B", "A")
	g.AddDependency("A", "B")

	_, err := g.Levels()
	if err == nil {
		t.Fatal("expected cycle detection error, got nil")
	}
}

func TestExecuteLevelsOrder(t *testing.T) {
	var order []string
	var idx int64

	g := model.NewGraph(nil)

	g.AddNode(&model.Node{
		ID: "A",
		Exec: func(ctx context.Context, n *model.Node) error {
			atomic.AddInt64(&idx, 1)
			order = append(order, "A")
			return nil
		},
	})

	g.AddNode(&model.Node{
		ID: "B",
		Exec: func(ctx context.Context, n *model.Node) error {
			atomic.AddInt64(&idx, 1)
			order = append(order, "B")
			return nil
		},
	})

	g.AddNode(&model.Node{
		ID: "C",
		Exec: func(ctx context.Context, n *model.Node) error {
			atomic.AddInt64(&idx, 1)
			order = append(order, "C")
			return nil
		},
	})

	// A → B → C
	g.AddDependency("B", "A")
	g.AddDependency("C", "B")

	if err := g.ExecLevels(t.Context(), genExec()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"A", "B", "C"}

	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("expected order %v, got %v", expected, order)
		}
	}
}

func TestExecuteLevelsStopsOnError(t *testing.T) {
	g := model.NewGraph(nil)

	g.AddNode(mockNode("A", func() error { return nil }))
	g.AddNode(mockNode("B", func() error { return errors.New("boom") }))
	g.AddNode(mockNode("C", func() error { return nil }))

	g.AddDependency("B", "A")
	g.AddDependency("C", "B") // should NOT run if B fails

	err := g.ExecLevels(t.Context(), genExec())
	if err == nil {
		t.Fatal("expected error from B, got nil")
	}
}

func TestExecuteParallelBehavior(t *testing.T) {
	var count int64

	g := model.NewGraph(nil)

	// A then B+C in parallel
	g.AddNode(mockNode("A", func() error {
		atomic.AddInt64(&count, 1)
		return nil
	}))

	g.AddNode(mockNode("B", func() error {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt64(&count, 1)
		return nil
	}))

	g.AddNode(mockNode("C", func() error {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt64(&count, 1)
		return nil
	}))

	g.AddDependency("B", "A")
	g.AddDependency("C", "A")

	start := time.Now()
	err := g.ExecLevels(t.Context(), genExec())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	elapsed := time.Since(start)

	// B and C each sleep 50ms but run in parallel,
	// so execution should NOT take ~100ms, but ~50ms.
	if elapsed > 80*time.Millisecond {
		t.Fatalf("expected parallel execution (≈50ms), got %v", elapsed)
	}

	if atomic.LoadInt64(&count) != 3 {
		t.Fatalf("expected all 3 Create() funcs executed")
	}
}
