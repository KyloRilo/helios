# Known Issues

Bugs, dead code, missing functionality, and design concerns. Items removed when fixed.

## Bugs

- **AwsCtrl embeds `CompImpl` instead of `CtrlShim`** — any shim call panics (`compute/aws.go`)
- **`AddTags` value receiver** — tag mutations silently lost (`compute/node.go`)
- **`Receive` value receiver on CoreService** — actor state mutations lost (`service/core/core.go`)
- **api.dockerfile references nonexistent `cmd/api/`** — build fails (`bin/docker/api.dockerfile`)
- **Unreachable switch case in Docker `createNode`** — `Image != "" && Context != nil` branch never hits (`compute/docker.go`)

## Dead Code

- `provider` field never set in `NewComputeController`
- `stateMgrRef *actor.PID` unused in CoreService
- `BaseError.error()` and `ErrorCode` field unused outside tests
- `HService.ID` never populated
- `isECR`/`isDockerHub` helpers unused in production
- `platform`/`authToken` on DockerCtrl never set
- `colorBg`/`colorBgLight` unused in TUI styles
- `OptionsApplier` interface never satisfied

## Missing Functionality

- No `ApplyPlan` method — plans generate but can't execute
- No `Error` status recovery — nodes stuck permanently after error
- No external variable injection in HCL eval
- No interfaces for actor/consul controllers — untestable
- No tests for actor or consul controllers
- Worker service and binary are empty stubs
- GCP compute implementation empty
- Storage controller fully commented out
- Plan diff only compares 5 of 20+ node fields
- TUI entirely mock-driven — no integration with pkg/ types
- `cmd/api/` referenced but never created
- Proto codegen not run — no `.pb.go` files

## Design Concerns

- Consul controller uses `log.Fatalf` instead of returning errors
- No shared service interface across CoreService/LeaderService
- `BuildFileEvalContext` nearly identical to `BuildEvalContext` — redundant
- `WorkerConfig` has no fields beyond embedded `HConfig`
