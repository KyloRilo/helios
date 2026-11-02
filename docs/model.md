# Model Layer

Domain types, HCL config parsing, compute node model, eval engine, and dependency graph/plan system. Everything here is pure data — no runtime dependencies on Docker, Consul, or Proto.Actor.

---

## Config Hierarchy

> `pkg/model/config.go`

HCL manifests are parsed into a `HManifest > HCluster > HService` hierarchy. All config types embed `HConfig` and implement the `Config` interface (`IsValid() (bool, error)`).

- **HManifest** — top-level, contains `[]HCluster`. Max 1 cluster allowed.
- **HCluster** — named cluster with `[]HService`. Validates name, services, and that the dependency graph is acyclic.
- **HService** — a container definition. Must have `Image` XOR `Build`. Supports optional sub-blocks `Healthcheck`, `Resources`, and `Build` (context, dockerfile, args, target). Valid restart policies: `""`, `no`, `always`, `on-failure`, `unless-stopped`.
- **LeaderConfig** / **WorkerConfig** — separate top-level configs for node binaries. Leader has `Name`, `ClusterName` (shared Proto.Actor cluster name, defaults to `"helios"`), `Host`, `Port`, `NodeID`, `RaftDir`, `RaftBind`, `Bootstrap`, `RaftPeers`. Worker is an empty stub.
- **ClusterState** (`state.go`) — serializable cluster state for CLI persistence. Contains cluster name, network name, manifest path, and `[]NodeState` (name, container ID, status, dependencies). Methods: `UpdateNodes`, `Save`, `LoadState`, `RemoveState`. State files live in `.helios/<cluster>.state.json`.

Each config type has `Parse*()` (from string) and `Read*File()` (from disk) constructors.

## Actor Service Interface

> `pkg/model/model.go`

`ActorService` defines `GetServiceName()` and `Receive(actor.Context)`. `BaseService` provides a default no-op implementation via `NewBaseActorService(name)`. Both `CoreService` and `LeaderService` embed this.

---

## Compute Node

> `pkg/model/compute/node.go`

The `Node` struct represents a container with all its configuration (image, ports, env, volumes, healthcheck, resources, etc.) plus runtime state (`Id`, `Status`).

**Construction** uses the functional options pattern: `NewNode(opts ...NodeOption)` where each `With*` function sets a field. `NewNode` defaults to `Status = Ready` only if no `WithStatus()` option was applied (i.e., status is still `Init` after all options run).

### Status State Machine

```
Init → Ready → Created → Updating → Up → Down → Destroyed
                                                    ↓
                             Error ← (any state)
```

Statuses are defined here as string constants, but **transitions are not enforced in the model layer** — that happens in `pkg/controller/compute/` via `CompImpl.isValidStatus()`.

---

## Error Types

> `pkg/model/errors/errors.go`

- **`InvalidNodeStatus`** — returned by the controller layer when a node action is attempted in a disallowed state. Contains `NodeName`, `NodeStatus`, `Expected` statuses, and `Action`.
- **`BaseError`** — has an `ErrorCode` and `Err`, but is essentially unused outside its own test.

---

## HCL Eval Engine

> `pkg/model/eval/eval.go`

Two entry points: `BuildEvalContext(src, filename)` returns an eval context + body; `BuildFileEvalContext(path, src)` returns just the context. Both run the same pipeline:

1. **Parse** HCL source
2. **Variables** — `variable` blocks resolved from their `default` attribute (no external injection). Populates `var.*` namespace.
3. **Locals** — topologically sorted by inter-local dependencies, then evaluated incrementally. Populates `local.*` namespace.
4. **Dynamic blocks** — `dynamic "type" { for_each = ... content { ... } }` expanded in-place. Supports list, map, and number iteration with `each.key`/`each.value`.
5. **Resource refs** — `service` and `cluster` block labels registered as string values (e.g., `service.db` → `"db"`).

**Functions available:** `format`, `join`, `upper`, `lower`, `replace`, `split`, `trimspace`, `substr`, `strlen`, `concat`, `coalesce`, `lookup` (map key with default), `range` (generate `[0..n-1]`).

---

## Graph & Plan Engine

### Graph (DAG)

> `pkg/model/graph/graph.go`

`Graph` is a directed acyclic graph built from a dependency map. Uses Kahn's algorithm for topological sort and level computation. Key methods: `TopoSort()`, `Levels()` (parallel-execution tiers), `Reverse()` (for teardown ordering), `Roots()`, `Dependents()`. Returns `CycleError` or `MissingDepError` on invalid input.

### ClusterGraph

> `pkg/model/graph/cluster.go`

Binds a `Graph` topology to `compute.Node` objects. Provides `Levels()` (startup order) and `ReverseLevels()` (teardown order) returning `[][]*compute.Node`. Construction validates that topology nodes and compute nodes match 1:1.

### Plan & Diff

> `pkg/model/graph/plan.go`

`Diff(desired, desiredDeps, current)` compares a desired `[]DiffTarget` against a `ClusterGraph` and produces a `Plan` with actions: `OpCreate`, `OpUpdate`, `OpDestroy`, `OpNoop`. Actions are organized into dependency-respecting levels.

Currently only diffs 5 fields: image, command, entrypoint, restart, replicas. Many node fields (ports, env, volumes, healthcheck, resources, etc.) are not compared.

---

## Known Issues

- **`AddTags` value receiver** (`compute/node.go`) — operates on a copy; tag mutations silently lost
- **`OptionsApplier` interface mismatch** (`compute/node.go`) — interface signature doesn't match `NodeOption.Apply`; never satisfied
- **Limited plan diff** (`graph/plan.go`) — only compares 5 of 20+ node fields
- **`WorkerConfig` empty** (`config.go`) — no fields, no real validation
- **`BaseError` unused** (`errors/errors.go`) — unexported `error()` method and `ErrorCode` never used
- **`HService.ID` unpopulated** (`config.go`) — no HCL tag, never set
- **No external variable injection** (`eval/eval.go`) — variables must have defaults; no runtime override mechanism
- **`BuildFileEvalContext` redundancy** (`eval/eval.go`) — nearly identical to `BuildEvalContext` but drops body return
