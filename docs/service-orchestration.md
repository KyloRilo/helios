# Service Orchestration

High-level orchestration services that compose controllers into cluster lifecycle management.

**Source:** `pkg/service/core/`, `pkg/service/leader/`, `pkg/service/worker/`

---

## CoreService

`pkg/service/core/core.go`

The main orchestrator. Holds a `ComputeController` and a cluster config (`HCluster`), and manages the full container lifecycle through a dependency-aware graph.

**Lifecycle phases** — each operates on the cluster graph level-by-level, collecting all failures rather than short-circuiting:

| Phase | Direction | What it does |
|-------|-----------|-------------|
| `CreateCluster` | Forward (topological) | Creates `helios-<name>` network, generates graph, creates all nodes on that network |
| `StartCluster` | Forward | Starts nodes in dependency order |
| `StopCluster` | Reverse | Stops nodes in reverse dependency order |
| `TeardownCluster` | Reverse | Removes nodes in reverse dependency order, then removes the network |

**Graph generation:** `GenGraph` converts `[]HService` into `[]*compute.Node` using functional options, builds a DAG topology from `depends_on`, and wraps it in a `ClusterGraph`.

**Partial failure:** `CreateCluster` always sets the graph even when some nodes fail, so callers can inspect node statuses and persist state. All lifecycle methods collect failures rather than short-circuiting.

**Plan/update:** `PlanUpdate` diffs a desired `HCluster` against the current graph using `graph.Diff`. There is no `ApplyPlan` yet — plans are informational only.

**Construction:** `NewCoreService(ctx, CoreArgs)` — `CoreArgs.Stub` accepts a `*ComputeController` for test injection; otherwise creates a real controller via `NewComputeController`.

---

## LeaderService

`pkg/service/leader/leader.go`

The composition point for a leader node. Embeds `CoreService` and gates all cluster operations behind Raft leadership — only the elected leader can mutate cluster state.

**Key behaviors:**
- `requireLeader()` checks `raft.IsLeader()` before any operation; returns `ErrNotLeader` if not the active leader
- Mutation operations (create/start/stop/teardown) hold a write lock; read operations (plan) hold a read lock
- Registers itself as a Proto.Actor actor kind via `GenServiceMap`, handling `Started`/`Stopping`/`Stopped` messages
- `BootstrapRaft(peers)` bootstraps a new Raft cluster (first node only)

**Construction:** `NewLeader(ctx, LeaderArgs)` wires up Consul, Raft, and the actor cluster. The Proto.Actor cluster uses `LeaderConfig.ClusterName` (defaults to `"helios"`) so all leaders share a single Consul service name for discovery. Set `Stub: true` to skip external dependencies for testing.

---

## WorkerService

`pkg/service/worker/worker.go` — empty stub, package declaration only.

---

## Testing

Tests use stub `ComputeController` implementations that override individual methods. `CoreArgs.Stub` injects the stub directly. Leader tests use `LeaderArgs{Stub: true}` to bypass Consul/Raft/Actor setup. See `core/core_test.go` and `leader/leader_test.go`.

---

## Known Issues

- `stateMgrRef *actor.PID` in CoreService is declared but never used (`core.go:18`)
- `Receive` on CoreService uses a value receiver — state mutations are lost (`core.go:278`)
- No `ApplyPlan` method to execute generated plans
- No shared interface between CoreService and LeaderService
- Worker service is entirely unimplemented
