# CLI & TUI

Command-line interfaces to Helios.

---

## CLI (`cmd/cli/main.go`)

Build: `make cli` → `bin/helios/helios`. All commands take positional args (no flags).

| Command | Args | What it does |
|---------|------|--------------|
| `validate` | `<manifest>` | Parse and validate a manifest, print service table |
| `plan` | `<current> <desired>` | Diff two manifests, show execution plan with levels and ops |
| `graph` | `<manifest>` | Show dependency graph with topological levels |
| `up` | `<manifest>` | Create and start a cluster; write state file |
| `down` | `<manifest>` | Stop and teardown a cluster using persisted state |
| `help` | — | Print usage |

Internally, commands share a pipeline: `loadManifest` → `clusterFromManifest` → `buildClusterGraph` (converts HServices to compute Nodes, builds topology). `up`/`down` construct a `CoreService` with the Docker backend.

### State File (`.helios/<cluster>.state.json`)

`up` persists a `ClusterState` (defined in `pkg/model/state.go`) after every operation — both `CreateCluster` and `StartCluster` — even on partial failure. The state records the cluster name, network name, manifest path, and each node's name, container ID, status, and dependencies.

`down` reads this state to reconstruct the cluster graph with exact container IDs, so it only targets resources the orchestrator created. After teardown the state file is deleted. Guards: `up` refuses if state already exists (prevents orphaning); `down` refuses if no state is found.

State files are gitignored via `.helios` in `.gitignore`.

---

## TUI (`cmd/ui/`)

Build: `make ui`. Built with Bubble Tea + Lip Gloss.

Currently **mock-driven only** — no integration with `pkg/` types. Mock data in `mock.go` provides sample clusters and randomized metrics.

**Screens:** Splash → Init (animated) → Main Menu → Cluster List → Cluster Detail → Dashboard → Submit Config → Confirm Delete

**Files:** `ui.go` (entrypoint), `model.go` (state/update), `views.go` (rendering), `styles.go` (styles), `mock.go` (mock data)

---

## Leader Binary (`cmd/leader/leader.go`)

Built via `bin/docker/leader.Dockerfile` (Go 1.24.2 base, exposes ports 6330 for Raft and 6331 for Proto.Actor).

**Flags:** `--config` (default `/helios/config/config.hcl`)

**Flow:** Parse flags → read/validate `LeaderConfig` → derive Raft settings (bind, dir, node ID with defaults from config) → construct `LeaderService` (wires CoreService + Raft + Consul + Actor system) → `SpinUp` → bootstrap Raft if `config.Bootstrap == true` (reads peers from `config.RaftPeers`) → block on signal → `SpinDown`

---

## Worker Binary (`cmd/worker/worker.go`)

Empty stub — no implementation. Commented code indicates planned Proto.Actor cluster membership.

---

## Known Issues

- TUI is entirely mock-driven, no integration with `pkg/` types
- `cmd/api/` referenced in `bin/docker/api.dockerfile` but doesn't exist
- Worker binary is empty
