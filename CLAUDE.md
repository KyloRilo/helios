# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Helios is a WIP proof-of-concept AI-first Docker/compute orchestrator written in Go. It reads HCL-based cluster manifests, manages container lifecycle (create/start/stop/remove) across pluggable compute backends (Docker, AWS ECR, GCP GCR), and coordinates nodes using Proto.Actor for clustering with Consul for service discovery and Raft for consensus.

## Build & Test Commands

```bash
# Build everything (tidy, build, run unit tests)
make

# Run unit tests only (no Docker required — tests use stub controllers)
go test -v ./pkg/...

# Run a single test
go test -v -run TestCreateNodeSuccess ./pkg/controller/compute/

# Run integration tests (requires Docker, 5min timeout)
make integration

# Regenerate gRPC protobuf code
make gen-proto

# Run the launchpad local dev cluster (reads bin/helios/local.cluster.hcl)
make launchpad
```

The default `make` target passes `-docker-path` and `-docker-file` flags to `go test`. Unit tests in `pkg/` use stub/shim implementations of `CtrlShim` and don't require a running Docker daemon.

## Architecture

### Three-layer design: Model → Controller → Service

- **`pkg/model/`** — Domain types and HCL config parsing. `HManifest > HCluster > HService` is the config hierarchy. `compute.Node` represents a container with a status state machine (`Init → Ready → Created → Up → Down → Destroyed`, with `Error` from any state). Node construction uses functional options (`WithName`, `WithImage`, etc.).

- **`pkg/controller/`** — Backend integrations, each behind an interface:
  - `compute/` — `ComputeController` interface with `CompImpl` base that delegates to a `CtrlShim` (Docker, AWS, GCP). `CompImpl` enforces node status transitions before calling the underlying shim. Tests inject stubs via the `CtrlShim` interface and the `stub` field in `ControllerArgs`.
  - `actor/` — Proto.Actor cluster management via Consul-backed provider.
  - `consul/` — Consul client wrapper for service discovery.
  - `raft/` — Hashicorp Raft consensus with BoltDB storage.

- **`pkg/service/`** — High-level orchestration services that compose controllers:
  - `core/` — `CoreService` is the main orchestrator: validates cluster config, generates nodes from services, manages full cluster lifecycle (create → start → stop → teardown).
  - `leader/` — `LeaderService` sets up the actor cluster and Consul integration.
  - `worker/` — Stub, not yet implemented.

### Entrypoints (`cmd/`)

- `launchpad/` — Local dev runner. Reads a manifest file, creates and tears down a cluster via `CoreService`.
- `leader/` — Leader node binary, reads config via `--config` flag (default `/helios/config/config.hcl`).
- `worker/` — Worker node binary (stub).

### Config format

Cluster configs are HCL files. A manifest contains `cluster` blocks, each containing `service` blocks. Services must have either `image` or `build` (not both). The `HELIOS_CONFIG_FILE` env var overrides the default config path for launchpad.

## Testing Patterns

Tests use Go's standard `testing` package (no testify assertions). Compute controller tests create stub implementations by embedding `TestCompCtrl` (which implements all `CtrlShim` methods as no-ops) and overriding specific methods. Tests verify both success paths and that node status is set to `Error` on failure.

## CI

GitHub Actions runs on pushes/PRs to `master`: Go 1.24.1, `go build -v ./...`, `go test -v ./...`.
