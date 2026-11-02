package leader

import (
	"context"
	"testing"

	compCtrl "github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/model/compute"
	"github.com/KyloRilo/helios/pkg/service/core"
)

type CompStub struct{ compCtrl.ComputeController }

func (c CompStub) CreateNode(_ context.Context, _ *compute.Node) (string, error) {
	return "test-id", nil
}

func (c CompStub) StartNode(_ context.Context, _ *compute.Node) error { return nil }
func (c CompStub) StopNode(_ context.Context, _ *compute.Node) error  { return nil }
func (c CompStub) RemoveNode(_ context.Context, _ *compute.Node) error { return nil }

func stubLeader(t *testing.T, conf *model.HCluster) *LeaderService {
	t.Helper()
	stub := compCtrl.ComputeController(CompStub{})
	l, err := NewLeader(t.Context(), LeaderArgs{
		Stub: true,
		CoreArgs: core.CoreArgs{
			Stub: &stub,
			Conf: conf,
		},
	})
	if err != nil {
		t.Fatalf("NewLeader: %v", err)
	}
	return l
}

func TestParseLeaderConfig(t *testing.T) {
	conf, err := model.ParseLeaderConfig(`
		name = "test-leader"
		host = "127.0.0.1"
		port = 6330
	`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if conf.Name != "test-leader" {
		t.Errorf("expected name 'test-leader' but got '%s'", conf.Name)
	}

	if conf.Host != "127.0.0.1" {
		t.Errorf("expected host '127.0.0.1' but got '%s'", conf.Host)
	}

	if conf.Port != 6330 {
		t.Errorf("expected port 6330 but got %d", conf.Port)
	}
}

func TestParseLeaderConfigWithRaft(t *testing.T) {
	conf, err := model.ParseLeaderConfig(`
		name      = "test-leader"
		host      = "127.0.0.1"
		port      = 6330
		node_id   = "node-1"
		raft_dir  = "/tmp/raft/node-1"
		raft_bind = "127.0.0.1:6330"
	`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if conf.NodeID != "node-1" {
		t.Errorf("expected node_id 'node-1' but got '%s'", conf.NodeID)
	}

	if conf.RaftDir != "/tmp/raft/node-1" {
		t.Errorf("expected raft_dir '/tmp/raft/node-1' but got '%s'", conf.RaftDir)
	}

	if conf.RaftBind != "127.0.0.1:6330" {
		t.Errorf("expected raft_bind '127.0.0.1:6330' but got '%s'", conf.RaftBind)
	}
}

func TestParseLeaderConfigMissingHost(t *testing.T) {
	_, err := model.ParseLeaderConfig(`
		name = "test-leader"
		port = 6330
	`)
	if err == nil {
		t.Error("expected error for missing host but got nil")
	}
}

func TestParseLeaderConfigMissingPort(t *testing.T) {
	_, err := model.ParseLeaderConfig(`
		name = "test-leader"
		host = "127.0.0.1"
	`)
	if err == nil {
		t.Error("expected error for missing port but got nil")
	}
}

func TestLeaderNotLeaderByDefault(t *testing.T) {
	l := stubLeader(t, genClusterConfig())
	if l.IsLeader() {
		t.Error("expected IsLeader() false when raft is nil")
	}
}

func TestLeaderRejectsOperationsWhenNotLeader(t *testing.T) {
	l := stubLeader(t, genClusterConfig())

	if err := l.CreateCluster(t.Context()); err != ErrNotLeader {
		t.Errorf("CreateCluster: expected ErrNotLeader, got %v", err)
	}

	if err := l.StartCluster(t.Context()); err != ErrNotLeader {
		t.Errorf("StartCluster: expected ErrNotLeader, got %v", err)
	}

	if err := l.StopCluster(t.Context()); err != ErrNotLeader {
		t.Errorf("StopCluster: expected ErrNotLeader, got %v", err)
	}

	if err := l.TeardownCluster(t.Context()); err != ErrNotLeader {
		t.Errorf("TeardownCluster: expected ErrNotLeader, got %v", err)
	}

	if _, err := l.PlanUpdate(&model.HCluster{}); err != ErrNotLeader {
		t.Errorf("PlanUpdate: expected ErrNotLeader, got %v", err)
	}
}

func TestLeaderBootstrapRaftNilReturnsError(t *testing.T) {
	l := stubLeader(t, genClusterConfig())
	err := l.BootstrapRaft(nil)
	if err == nil {
		t.Error("expected error bootstrapping nil raft")
	}
}

func TestLeaderStubInit(t *testing.T) {
	l := stubLeader(t, genClusterConfig())
	if l.core == nil {
		t.Error("expected core service to be initialized")
	}
	if l.GetServiceName() != "Leader" {
		t.Errorf("expected service name 'Leader', got '%s'", l.GetServiceName())
	}
}

func genClusterConfig() *model.HCluster {
	return &model.HCluster{
		Name: "test-cluster",
		Services: []model.HService{
			{Name: "db", Image: "postgres:16"},
			{Name: "api", Image: "nginx:latest", DependsOn: []string{"db"}},
		},
	}
}
