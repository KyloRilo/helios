package core

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	compCtrl "github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/model/compute"
	"github.com/KyloRilo/helios/pkg/model/graph"
	"github.com/google/uuid"
)

type CompStub struct{ compCtrl.ComputeController }

func (c CompStub) CreateNode(_ context.Context, _ *compute.Node) (string, error) {
	return "", nil
}

func (c CompStub) StartNode(_ context.Context, _ *compute.Node) error {
	return nil
}

func (c CompStub) StopNode(_ context.Context, _ *compute.Node) error {
	return nil
}

func (c CompStub) RemoveNode(_ context.Context, _ *compute.Node) error {
	return nil
}

var dockerPath = flag.String("docker-path", ".", "Path to dockerfile")
var dockerFile = flag.String("docker-file", "Dockerfile", "Name of the dockerfile")

func genClusterConfig() *model.HCluster {
	return &model.HCluster{
		Name: "test-cluster",
		Services: []model.HService{
			{
				Name: "db",
				Build: &model.Build{
					Context:    *dockerPath,
					Dockerfile: *dockerFile,
				},
			},
			{
				Name:      "api",
				Image:     "myapp:latest",
				DependsOn: []string{"db"},
			},
		},
	}
}

func setGraph(svc *CoreService) {
	cg, err := svc.GenGraph(svc.GetConfig().Services)
	if err != nil {
		panic(fmt.Sprintf("setGraph failed: %v", err))
	}
	for _, n := range cg.Nodes() {
		n.Id = uuid.New().String()
	}
	svc.SetGraph(cg)
}

func initCluster(ctx context.Context, stub compCtrl.ComputeController) CoreService {
	svc := NewCoreService(ctx, CoreArgs{
		stub: &stub,
		Conf: genClusterConfig(),
	})

	setGraph(&svc)
	return svc
}

func TestMain(m *testing.M) {
	flag.Parse()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestInitCluster(t *testing.T) {
	svc := initCluster(t.Context(), CompStub{})
	nodes := svc.GetNodes()
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes but got %d", len(nodes))
	}
}

func TestGetGraph(t *testing.T) {
	svc := initCluster(t.Context(), CompStub{})
	g := svc.GetGraph()
	if g == nil {
		t.Fatal("expected graph but got nil")
	}
	if g.NodeCount() != 2 {
		t.Errorf("expected 2 nodes in graph but got %d", g.NodeCount())
	}
}

func TestGraphLevels(t *testing.T) {
	svc := initCluster(t.Context(), CompStub{})
	levels, err := svc.GetGraph().Levels()
	if err != nil {
		t.Fatal(err)
	}
	if len(levels) != 2 {
		t.Fatalf("expected 2 levels but got %d", len(levels))
	}
	if levels[0][0].Name != "db" {
		t.Errorf("expected db in level 0 but got %s", levels[0][0].Name)
	}
	if levels[1][0].Name != "api" {
		t.Errorf("expected api in level 1 but got %s", levels[1][0].Name)
	}
}

func TestGetNodesNilGraph(t *testing.T) {
	svc := NewCoreService(t.Context(), CoreArgs{
		stub: func() *compCtrl.ComputeController {
			var s compCtrl.ComputeController = CompStub{}
			return &s
		}(),
		Conf: genClusterConfig(),
	})
	if svc.GetNodes() != nil {
		t.Error("expected nil nodes when graph is nil")
	}
}

type CreatePasses CompStub

func (c CreatePasses) CreateNode(_ context.Context, n *compute.Node) (string, error) {
	return "some-id", nil
}

func TestCreateClusterPasses(t *testing.T) {
	svc := initCluster(t.Context(), CreatePasses{})
	err := svc.CreateCluster(t.Context())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
	if svc.GetGraph() == nil {
		t.Error("expected graph to be set after CreateCluster")
	}
}

type CreateFailed CompStub

func (c CreateFailed) CreateNode(_ context.Context, _ *compute.Node) (string, error) {
	return "", fmt.Errorf("failed to create container")
}

func TestCreateClusterCreateFails(t *testing.T) {
	svc := initCluster(t.Context(), CreateFailed{})
	err := svc.CreateCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

type StartFailed CompStub

func (c StartFailed) StartNode(_ context.Context, _ *compute.Node) error {
	return fmt.Errorf("failed to start container")
}

func TestStartClusterFails(t *testing.T) {
	svc := initCluster(t.Context(), StartFailed{})
	err := svc.StartCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

type StopFailed CompStub

func (c StopFailed) StopNode(_ context.Context, _ *compute.Node) error {
	return fmt.Errorf("failed to stop container")
}

func TestStopClusterFails(t *testing.T) {
	svc := initCluster(t.Context(), StopFailed{})
	err := svc.StopCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

type TeardownPasses CompStub

func (t TeardownPasses) RemoveNode(_ context.Context, _ *compute.Node) error {
	return nil
}

func TestTeardownCluster(t *testing.T) {
	svc := initCluster(t.Context(), TeardownPasses{})
	err := svc.TeardownCluster(t.Context())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

type RemoveFailed CompStub

func (c RemoveFailed) RemoveNode(ctx context.Context, _ *compute.Node) error {
	return fmt.Errorf("failed to remove container")
}

func TestTeardownClusterRemoveFails(t *testing.T) {
	svc := initCluster(t.Context(), RemoveFailed{})
	err := svc.TeardownCluster(t.Context())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

func TestPlanUpdateNewService(t *testing.T) {
	svc := initCluster(t.Context(), CompStub{})

	desired := &model.HCluster{
		Name: "test-cluster",
		Services: []model.HService{
			{Name: "db", Image: "postgres"},
			{Name: "api", Image: "myapp:latest", DependsOn: []string{"db"}},
			{Name: "web", Image: "nginx", DependsOn: []string{"api"}},
		},
	}

	plan, err := svc.PlanUpdate(desired)
	if err != nil {
		t.Fatal(err)
	}

	if !plan.HasChanges() {
		t.Error("expected plan to have changes")
	}

	var foundCreate bool
	for _, a := range plan.Actions {
		if a.Name == "web" && a.Op == graph.OpCreate {
			foundCreate = true
		}
	}
	if !foundCreate {
		t.Error("expected OpCreate for web service")
	}
}

func TestPlanUpdateImageChange(t *testing.T) {
	svc := initCluster(t.Context(), CompStub{})

	desired := &model.HCluster{
		Name: "test-cluster",
		Services: []model.HService{
			{Name: "db", Image: "postgres:15"},
			{Name: "api", Image: "myapp:v2", DependsOn: []string{"db"}},
		},
	}

	plan, err := svc.PlanUpdate(desired)
	if err != nil {
		t.Fatal(err)
	}

	if !plan.HasChanges() {
		t.Error("expected plan to have changes")
	}

	var foundUpdate bool
	for _, a := range plan.Actions {
		if a.Name == "api" && a.Op == graph.OpUpdate {
			foundUpdate = true
		}
	}
	if !foundUpdate {
		t.Error("expected OpUpdate for api service")
	}
}

func TestPlanUpdateNoGraph(t *testing.T) {
	svc := NewCoreService(t.Context(), CoreArgs{
		stub: func() *compCtrl.ComputeController {
			var s compCtrl.ComputeController = CompStub{}
			return &s
		}(),
		Conf: genClusterConfig(),
	})

	_, err := svc.PlanUpdate(genClusterConfig())
	if err == nil {
		t.Error("expected error when no current graph exists")
	}
}
