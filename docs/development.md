# Development Guide

Technical reference for building, testing, and CI workflows.

---

## Build Commands

```bash
# Build everything (tidy, build, run unit tests)
make

# Build CLI binary to bin/helios/helios
make cli

# Run TUI (sets HELIOS_CONFIG_FILE, builds, then runs cmd/ui/)
make ui

# Regenerate gRPC protobuf code
make gen-proto
```

## Test Commands

```bash
# Run unit tests only (no Docker required — tests use stub controllers)
go test -v ./pkg/...

# Run a single test
go test -v -run TestCreateNodeSuccess ./pkg/controller/compute/

# Run integration tests (requires Docker, 5min timeout)
make integration

# Default make target passes -docker-path and -docker-file flags to go test
```

## Testing Patterns

Tests use Go's standard `testing` package — no testify or assertion libraries.

### Compute Controller Tests

Stub implementations use an embedding chain pattern:

1. `TestCompCtrl` — base stub implementing all `CtrlShim` methods as no-ops
2. Test-specific stubs embed `TestCompCtrl` and override individual methods
3. `CreatePasses` → `StartPasses` → `StopPasses` chain builds up passing lifecycle
4. `*Failed` variants return errors for the target method
5. `newTestCtrl()` helper calls `NewComputeController` with the stub arg

Tests drive through lifecycle stages (create → start → stop → remove) verifying node status at each step. Invalid-status tests verify rejections.

See `pkg/controller/compute/compute_test.go` for the full pattern.

### Service Tests

Similar approach — define small named stub types that embed `CompStub` (which itself embeds `ComputeController` interface) and override individual methods. Tests verify lifecycle success paths and error propagation.

See `pkg/service/core/core_test.go`.

### Model Tests

Standard table-driven tests for config parsing, node construction, graph operations, eval engine, and plan diffing. No stubs needed — pure domain logic.

## CI

GitHub Actions (`.github/workflows/ci.yml`):

- **Triggers:** push/PR to `master`
- **Environment:** Go 1.24.1
- **Steps:** `go build -v ./...` → `go test -v ./...`
- **Not in CI:** integration tests, proto generation, linting

## Config Loading

- `HELIOS_CONFIG_FILE` env var overrides the default config path for the CLI and TUI
- Leader binary uses `--config` flag (default `/helios/config/config.hcl`)
- Config files live in `bin/helios/` (e.g., `local.cluster.hcl`, `qa.cluster.hcl`, `integration.cluster.hcl`)

## Integration Testing with the CLI

The `integration.cluster.hcl` manifest defines a minimal image-only cluster (redis, consul, nginx) for testing the full compute graph lifecycle against Docker without needing local Dockerfiles:

```bash
# Full lifecycle: pull images, create containers in topological order, start them
go run ./cmd/cli/ up bin/helios/integration.cluster.hcl

# Verify containers are running
docker ps

# Teardown: stop and remove all containers
go run ./cmd/cli/ down bin/helios/integration.cluster.hcl
```

This tests that the dependency graph correctly orders operations (level 0: redis + consul, level 1: api) and that the Docker shim can pull, create, start, stop, and remove containers.
