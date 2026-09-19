package graph

import (
	"strings"
	"testing"

	"github.com/KyloRilo/helios/pkg/model/compute"
)

func buildCurrentGraph(t *testing.T, nodes []*compute.Node, deps map[string][]string) *ClusterGraph {
	t.Helper()
	topo, err := New(deps)
	if err != nil {
		t.Fatal(err)
	}
	cg, err := NewClusterGraph(topo, nodes)
	if err != nil {
		t.Fatal(err)
	}
	return cg
}

func TestDiffNewService(t *testing.T) {
	current := buildCurrentGraph(t,
		[]*compute.Node{compute.NewNode(compute.WithName("db"), compute.WithImage("postgres"))},
		map[string][]string{"db": {}},
	)

	desired := []DiffTarget{
		{Name: "db", Image: "postgres"},
		{Name: "api", Image: "myapp"},
	}
	desiredDeps := map[string][]string{"db": {}, "api": {"db"}}

	plan, err := Diff(desired, desiredDeps, current)
	if err != nil {
		t.Fatal(err)
	}

	if !plan.HasChanges() {
		t.Error("expected plan to have changes")
	}

	var found bool
	for _, a := range plan.Actions {
		if a.Name == "api" && a.Op == OpCreate {
			found = true
		}
	}
	if !found {
		t.Error("expected OpCreate for api service")
	}
}

func TestDiffRemovedService(t *testing.T) {
	current := buildCurrentGraph(t,
		[]*compute.Node{
			compute.NewNode(compute.WithName("db"), compute.WithImage("postgres")),
			compute.NewNode(compute.WithName("old"), compute.WithImage("legacy")),
		},
		map[string][]string{"db": {}, "old": {}},
	)

	desired := []DiffTarget{{Name: "db", Image: "postgres"}}
	desiredDeps := map[string][]string{"db": {}}

	plan, err := Diff(desired, desiredDeps, current)
	if err != nil {
		t.Fatal(err)
	}

	var found bool
	for _, a := range plan.Actions {
		if a.Name == "old" && a.Op == OpDestroy {
			found = true
		}
	}
	if !found {
		t.Error("expected OpDestroy for old service")
	}
}

func TestDiffUpdatedService(t *testing.T) {
	current := buildCurrentGraph(t,
		[]*compute.Node{compute.NewNode(compute.WithName("api"), compute.WithImage("myapp:v1"))},
		map[string][]string{"api": {}},
	)

	desired := []DiffTarget{{Name: "api", Image: "myapp:v2"}}
	desiredDeps := map[string][]string{"api": {}}

	plan, err := Diff(desired, desiredDeps, current)
	if err != nil {
		t.Fatal(err)
	}

	if !plan.HasChanges() {
		t.Error("expected changes")
	}

	for _, a := range plan.Actions {
		if a.Name == "api" {
			if a.Op != OpUpdate {
				t.Errorf("expected OpUpdate but got %s", a.Op)
			}
			if !strings.Contains(a.Reason, "image") {
				t.Errorf("expected reason to mention image but got: %s", a.Reason)
			}
			return
		}
	}
	t.Error("api action not found")
}

func TestDiffNoChanges(t *testing.T) {
	current := buildCurrentGraph(t,
		[]*compute.Node{
			compute.NewNode(compute.WithName("db"), compute.WithImage("postgres")),
			compute.NewNode(compute.WithName("api"), compute.WithImage("myapp")),
		},
		map[string][]string{"db": {}, "api": {"db"}},
	)

	desired := []DiffTarget{
		{Name: "db", Image: "postgres"},
		{Name: "api", Image: "myapp"},
	}
	desiredDeps := map[string][]string{"db": {}, "api": {"db"}}

	plan, err := Diff(desired, desiredDeps, current)
	if err != nil {
		t.Fatal(err)
	}

	if plan.HasChanges() {
		t.Errorf("expected no changes but got: %s", plan.Summary())
	}
}

func TestDiffMultipleFieldChanges(t *testing.T) {
	current := buildCurrentGraph(t,
		[]*compute.Node{compute.NewNode(
			compute.WithName("api"),
			compute.WithImage("myapp:v1"),
			compute.WithCmd("serve"),
			compute.WithReplicas(1),
		)},
		map[string][]string{"api": {}},
	)

	desired := []DiffTarget{{Name: "api", Image: "myapp:v2", Command: "run", Replicas: 3}}
	desiredDeps := map[string][]string{"api": {}}

	plan, err := Diff(desired, desiredDeps, current)
	if err != nil {
		t.Fatal(err)
	}

	for _, a := range plan.Actions {
		if a.Name == "api" {
			if a.Op != OpUpdate {
				t.Errorf("expected OpUpdate but got %s", a.Op)
			}
			if !strings.Contains(a.Reason, "image") || !strings.Contains(a.Reason, "command") || !strings.Contains(a.Reason, "replicas") {
				t.Errorf("expected reason to mention image, command, and replicas but got: %s", a.Reason)
			}
			return
		}
	}
	t.Error("api action not found")
}

func TestPlanLevelsRespectOrder(t *testing.T) {
	current := buildCurrentGraph(t,
		[]*compute.Node{compute.NewNode(compute.WithName("db"), compute.WithImage("postgres"))},
		map[string][]string{"db": {}},
	)

	desired := []DiffTarget{
		{Name: "db", Image: "postgres"},
		{Name: "api", Image: "myapp"},
		{Name: "web", Image: "nginx"},
	}
	desiredDeps := map[string][]string{"db": {}, "api": {"db"}, "web": {"api"}}

	plan, err := Diff(desired, desiredDeps, current)
	if err != nil {
		t.Fatal(err)
	}

	levels := plan.Levels()
	if len(levels) < 2 {
		t.Fatalf("expected at least 2 levels but got %d", len(levels))
	}

	// db should be in an earlier level than api, api earlier than web
	levelOf := map[string]int{}
	for i, level := range levels {
		for _, a := range level {
			levelOf[a.Name] = i
		}
	}

	if levelOf["db"] >= levelOf["api"] {
		t.Errorf("db (level %d) should be before api (level %d)", levelOf["db"], levelOf["api"])
	}
	if levelOf["api"] >= levelOf["web"] {
		t.Errorf("api (level %d) should be before web (level %d)", levelOf["api"], levelOf["web"])
	}
}

func TestPlanSummary(t *testing.T) {
	plan := &Plan{
		Actions: []PlanAction{
			{Op: OpCreate},
			{Op: OpCreate},
			{Op: OpUpdate},
			{Op: OpNoop},
		},
	}

	summary := plan.Summary()
	if !strings.Contains(summary, "create 2") {
		t.Errorf("expected 'create 2' in summary: %s", summary)
	}
	if !strings.Contains(summary, "update 1") {
		t.Errorf("expected 'update 1' in summary: %s", summary)
	}
}

func TestPlanHasChangesEmpty(t *testing.T) {
	plan := &Plan{}
	if plan.HasChanges() {
		t.Error("expected empty plan to have no changes")
	}
}

func TestDiffInvalidTopology(t *testing.T) {
	current := buildCurrentGraph(t,
		[]*compute.Node{compute.NewNode(compute.WithName("db"), compute.WithImage("postgres"))},
		map[string][]string{"db": {}},
	)

	desired := []DiffTarget{{Name: "api", Image: "myapp"}}
	desiredDeps := map[string][]string{"api": {"nonexistent"}}

	_, err := Diff(desired, desiredDeps, current)
	if err == nil {
		t.Error("expected error for invalid desired topology")
	}
}
