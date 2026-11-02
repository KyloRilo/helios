# CLAUDE.md

> **This file is a directory.** Detailed technical references live in `docs/`. When performing work in any area below, consult the relevant doc first and update it to reflect any changes made during that work.

## Project

Helios is a WIP proof-of-concept AI-first Docker/compute orchestrator written in Go. It reads HCL-based cluster manifests, manages container lifecycle across pluggable compute backends (Docker, AWS ECR, GCP GCR), and coordinates nodes using Proto.Actor for clustering with Consul for service discovery and Raft for consensus.

## Documentation Index

| Document | Summary |
|---|---|
| [Model Layer](docs/model.md) | Domain types, HCL config parsing (HManifest → HCluster → HService), compute Node with functional options and status state machine, error types, HCL eval engine (variables, locals, functions, dynamic blocks), DAG graph and plan/diff engine, validation rules. The most mature layer. |
| [Compute Controller](docs/compute-controller.md) | Backend abstraction via CtrlShim → CompImpl → ComputeController delegation pattern. Status-enforcement guardrails, Docker implementation (pull/build/lifecycle), AWS skeleton, GCP empty. Test stub embedding chain pattern. |
| [Service Orchestration](docs/service-orchestration.md) | CoreService composes controllers for cluster lifecycle — create/start/stop/teardown with topological ordering. Graph generation from HService configs. Plan/update diffing (no ApplyPlan yet). LeaderService (Consul + Actor setup). WorkerService is an empty stub. |
| [Clustering & Infrastructure](docs/clustering.md) | Proto.Actor cluster management, Consul service discovery, Raft consensus (all concrete structs, no interfaces, no tests). Storage controller fully commented out. Docker-compose topology, Dockerfiles, Consul configs, proto definition. |
| [HCL Configuration](docs/hcl-config.md) | Complete schema reference for the user-facing HCL config format. Variables, locals with topological sorting, 14+ functions, dynamic block expansion, dependency declaration. Annotated examples from local and QA cluster configs. |
| [CLI & TUI](docs/cli-tui.md) | CLI commands (validate, plan, graph, up, down) and their internal flows. Bubble Tea TUI with screens for cluster management and metrics dashboard (currently mock-driven). Leader/worker binary entrypoints. |
| [Development Guide](docs/development.md) | Build commands (make, make cli, make ui), test commands, testing patterns (stub embedding chains), CI pipeline (GitHub Actions on master), config loading (env vars, flags, file paths). |
| [Known Issues](docs/known-issues.md) | Consolidated index of 38 issues: 9 bugs (runtime failures), 14 dead code items, 14 missing functionality gaps, 7 design concerns. Each links back to the detailed doc where it was identified. |

## Maintenance Rule

When making changes to the codebase, update the corresponding doc in `docs/` to reflect those changes. This keeps the documentation accurate as the single source of truth for each area. If a change spans multiple areas, update all relevant docs.
