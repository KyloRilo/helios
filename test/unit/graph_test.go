package unit

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KyloRilo/helios/pkg/model"
)

func mockExecCtx() model.ExecCtx {
	return model.ExecCtx{
		Create: func(ctx context.Context, n *model.Node) (interface{}, error) {
			return n, nil
		},
		Start: func(ctx context.Context, n *model.Node) (interface{}, error) {
			return n, nil
		},
		Read: func(ctx context.Context, n *model.Node) (interface{}, error) {
			return n, nil
		},
		Update: func(ctx context.Context, n *model.Node) (interface{}, error) {
			return n, nil
		},
		Stop: func(ctx context.Context, n *model.Node) (interface{}, error) {
			return n, nil
		},
		Delete: func(ctx context.Context, n *model.Node) (interface{}, error) {
			return n, nil
		},
	}
}

// helper to make a node with a mock Create() function
func mockNode(id string, execCtx *model.ExecCtx) *model.Node {
	var ctx model.ExecCtx = *execCtx
	if execCtx == nil {
		ctx = mockExecCtx()
	}

	return &model.Node{
		ExecCtx: ctx,
		Meta: model.NodeMeta{
			Name: id,
			ID:   id,
		},
	}
}

func TestLevelsSimple(t *testing.T) {
	g := model.NewGraph()

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
	g := model.NewGraph()

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
	g := model.NewGraph()

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

	g := model.NewGraph()
	aNode := mockNode("A", &model.ExecCtx{
		Create: func(ctx context.Context, n *model.Node) (interface{}, error) {
			atomic.AddInt64(&idx, 1)
			order = append(order, "A")
			return nil, nil
		},
	})

	bNode := mockNode("B", &model.ExecCtx{
		Create: func(ctx context.Context, n *model.Node) (interface{}, error) {
			atomic.AddInt64(&idx, 1)
			order = append(order, "B")
			return nil, nil
		},
	})

	cNode := mockNode("C", &model.ExecCtx{
		Create: func(ctx context.Context, n *model.Node) (interface{}, error) {
			atomic.AddInt64(&idx, 1)
			order = append(order, "C")
			return nil, nil
		},
	})

	g.AddNode(aNode)
	g.AddNode(bNode)
	g.AddNode(cNode)

	// A → B → C
	g.AddDependency("B", "A")
	g.AddDependency("C", "B")

	genExec := func(ctx model.ExecCtx) model.ExecFunc {
		return ctx.Create
	}

	if err := g.ExecLevels(t.Context(), genExec); err != nil {
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
	g := model.NewGraph()
	bNode := mockNode("B", &model.ExecCtx{
		Create: func(ctx context.Context, n *model.Node) (interface{}, error) {
			return nil, errors.New("boom")
		},
	})

	g.AddNode(mockNode("A", nil))
	g.AddNode(bNode)
	g.AddNode(mockNode("C", nil))

	g.AddDependency("B", "A")
	g.AddDependency("C", "B") // should NOT run if B fails

	genExec := func(ctx model.ExecCtx) model.ExecFunc {
		return ctx.Create
	}

	err := g.ExecLevels(t.Context(), genExec)
	if err == nil {
		t.Fatal("expected error from B, got nil")
	}
}

func TestExecuteParallelBehavior(t *testing.T) {
	var count int64

	g := model.NewGraph()

	// A then B+C in parallel
	aNode := mockNode("A", &model.ExecCtx{
		Create: func(ctx context.Context, n *model.Node) (interface{}, error) {
			atomic.AddInt64(&count, 1)
			time.Sleep(5 * time.Millisecond)
			return nil, nil
		},
	})

	bNode := mockNode("B", &model.ExecCtx{
		Create: func(ctx context.Context, n *model.Node) (interface{}, error) {
			atomic.AddInt64(&count, 1)
			time.Sleep(5 * time.Millisecond)
			return nil, nil
		},
	})

	cNode := mockNode("C", &model.ExecCtx{
		Create: func(ctx context.Context, n *model.Node) (interface{}, error) {
			atomic.AddInt64(&count, 1)
			time.Sleep(5 * time.Millisecond)
			return nil, nil
		},
	})

	g.AddNode(aNode)
	g.AddNode(bNode)
	g.AddNode(cNode)

	g.AddDependency("B", "A")
	g.AddDependency("C", "A")

	genExec := func(ctx model.ExecCtx) model.ExecFunc {
		return ctx.Create
	}

	start := time.Now()
	err := g.ExecLevels(t.Context(), genExec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed >= 15*time.Millisecond {
		t.Fatalf("expected parallel execution (up to 10ms), got %v", elapsed)
	}

	if atomic.LoadInt64(&count) != 3 {
		t.Fatalf("expected all 3 Create() funcs executed")
	}
}
