# Compute Controller

Backend abstraction for container lifecycle management (`pkg/controller/compute/`).

---

## Architecture

The compute layer uses a two-tier interface pattern. `CtrlShim` (unexported) is the contract that backend implementations satisfy — it has seven methods: `createNetwork`, `removeNetwork`, `createNode`, `startNode`, `listNodes`, `stopNode`, `removeNode`. `ComputeController` (exported) is the public interface consumed by the service layer with the same operations plus `GetProvider()`.

`CompImpl` sits between them. It's the only type implementing `ComputeController`, and it wraps any `CtrlShim` to enforce node status transitions. Every mutating method follows the same flow: validate the node's current status against allowed statuses for that action → delegate to the shim → on success, transition the node to the next status; on failure, set `compute.Error`. See `compute.go` for the full interface definitions and `isValidStatus()` logic.

```
Service Layer → ComputeController → CompImpl (status guard) → CtrlShim → Docker/AWS/GCP
```

## Status Enforcement

Each action (Create, Start, Stop, Remove) defines which node statuses are valid entry points. For example, `Create` requires `Ready` or `Destroyed`; `Start` requires `Created` or `Down`. Invalid transitions return `errors.InvalidNodeStatus`. Any shim error sets the node to `Error`, which is currently a terminal state with no recovery path. See `getExpectedStatuses()` in `compute.go` for the full mapping.

## Factory

`NewComputeController(ctx, ControllerArgs)` selects a backend: test stub → AWS (if creds provided) → Docker (default). Returns `CompImpl{CtrlShim: ctrl}`.

## Backends

**Docker** (`docker.go`) — the only complete implementation. Handles image pull, build-from-context (tar archive), and container lifecycle via the Docker SDK (`github.com/docker/go-sdk`). Empty commands default to `sleep infinity`. Network management: `createNetwork` creates a user-defined bridge network via the SDK's `network.New()`; `removeNetwork` removes it. When a node has entries in `Networks`, `createNode` uses `container.WithNetworkName` (enabling DNS resolution between containers); otherwise falls back to the default bridge. Volume mounts use `container.WithHostConfigModifier` with absolute paths. See `DockerCtrl` and its methods.

**AWS** (`aws.go`) — stub only. `AwsCtrl` exists with an `AwsCreds` struct. Network methods are no-op stubs. Other shim methods are unimplemented — will panic at runtime if selected.

**GCP** (`gcp.go`) — empty file, package declaration only.

## Test Patterns

Tests use an embedding chain of stub types rooted in `TestCompCtrl` (all no-ops). Specific stubs like `CreatePasses`, `StartPasses`, `StopFailed` override individual methods to return success or error. This lets tests drive nodes through lifecycle stages and verify status transitions at each step. See `compute_test.go`.

## Known Issues

- `provider` field never set in factory — `GetProvider()` always returns `""`
- `AwsCtrl` embeds `CompImpl` instead of implementing `CtrlShim` — runtime panic
- GCP implementation empty
- Unreachable switch case in Docker `createNode` (`docker.go:159`)
- `isECR`/`isDockerHub` helpers unused in production code
- `platform`/`authToken` fields on `DockerCtrl` never populated
- No recovery path from `Error` status
