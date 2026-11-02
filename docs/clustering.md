# Clustering & Infrastructure

Distributed systems layer: Proto.Actor clustering, Consul discovery, Raft consensus, deployment topology.

---

## Actor Controller

**`pkg/controller/actor/actor.go`**

Manages the Proto.Actor cluster. `ClusterConfig` defines the cluster name, host/port, Consul config, and a `SrvcGens` factory that produces `ActorService` implementations to register as cluster Kinds. All leaders must share the same `Name` so Consul groups them under one service for mutual discovery — `LeaderService` defaults this to `"helios"` via `LeaderConfig.ClusterName`. `ActorSysController` wraps the cluster instance and provides `Start()`, `Shutdown()`, `LogClusterEvents()`, and `LogMembers()`. Uses Consul as the cluster provider and distributed hash for identity lookup. No interface — concrete struct only.

## Consul Controller

**`pkg/controller/consul/consul.go`**

Thin wrapper around the HashiCorp Consul API client. `NewConsulController()` reads `CONSUL_HTTP_ADDR` from the environment and creates a client — calls `log.Fatalf` on failure (should return error). Provides `GetConfig()` and `GetServices()` (logs services but doesn't return them). No interface.

## Raft Controller

**`pkg/controller/raft/raft.go`**

Handles Raft consensus for leader election. `InitRaftService(RaftConfig)` returns `(*RaftService, error)` — sets up BoltDB log/stable stores, file snapshots, and TCP transport on port 6330. All errors are returned, not panicked. Key methods: `IsLeader()`, `LeaderAddress()`, `AddVoter(id, address)`, `RemoveServer(id)`, `BootstrapCluster(servers)`. The FSM (`RaftServiceState`) is mutex-protected and supports snapshot/restore. Currently used for election only — FSM will track cluster state in a future phase. No interface.

**Port allocation:** Raft TCP transport uses port 6330; Proto.Actor remote uses port 6331. These must not overlap — the original single-port setup caused `bind: address already in use` panics.

## Storage Controller

**`pkg/controller/storage/`** — fully commented out. Contains a planned `StorageService` interface and AWS S3 implementation as comments only.

## Deployment Topology

**QA cluster** (`bin/helios/qa.cluster.hcl`) defines the integration test topology: 3 Consul servers + 1 client, 3 Helios leaders. Brought up via `helios up`, which creates a `helios-qa` user-defined Docker network (enabling DNS between containers). Leaders mount per-node config from `bin/config/leader-{1,2,3}.hcl` and the Docker socket (`/var/run/docker.sock`). Leader-1 bootstraps the Raft cluster with peers 2 and 3. Consul configs live in `bin/consul/` (datacenter `dc1`, retry-join across all 3 servers). Legacy `docker-compose.yml` is no longer the primary deployment method.

## Dockerfiles

**`bin/docker/`** contains three Dockerfiles (Go 1.24.2 base):

- **leader.Dockerfile** — builds `cmd/leader`, creates empty Docker config for SDK init, exposes 6330 (Raft) and 6331 (Proto.Actor)
- **worker.Dockerfile** — builds `cmd/worker` (produces empty binary)
- **api.dockerfile** — **broken**: references `cmd/api/` which doesn't exist

Note: `docker-compose.yml` references `build/docker/` paths but files actually live in `bin/docker/`.

## Proto Definition

**`pkg/api/proto/helios.proto`** — minimal: one `Helios` service with a `HealthCheck` RPC. No generated `.pb.go` files exist.

## Known Issues

- No interfaces for actor/consul/raft controllers — prevents mocking and testing
- No test coverage for any clustering controller
- Consul controller uses `log.Fatalf` instead of returning errors
- Storage controller fully dormant (all commented out)
- `api.dockerfile` references nonexistent `cmd/api/`
- Dockerfile path mismatch: docker-compose references `build/docker/`, files are in `bin/docker/`
