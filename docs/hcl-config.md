# HCL Configuration

Reference for the Helios HCL config format.

**Source:** `pkg/model/config.go`, `pkg/model/eval/eval.go`
**Examples:** `bin/helios/local.cluster.hcl`, `bin/helios/qa.cluster.hcl`

---

## Schema

Configs follow a `manifest > cluster > service` hierarchy. A manifest may contain one cluster (multi-cluster is not yet supported). Each cluster contains one or more services.

```hcl
cluster "my-cluster" {
  service "db"  { image = "postgres:16" }
  service "api" { image = "nginx:latest"; depends_on = ["db"] }
}
```

### Service Attributes

A service must have either `image` or a `build` block — not both.

| Attribute | Type | Notes |
|---|---|---|
| `image` | string | Container image (required if no `build`) |
| `build` | block | Has `context`, `dockerfile`, optional `args` and `target` |
| `command`, `entrypoint` | string | Override container defaults |
| `ports` | map | `"host" = "container"` |
| `environment` | map | Env vars |
| `volumes` | map | `"host_path" = "container_path"` |
| `depends_on` | list | Service names this depends on — drives topological ordering |
| `restart` | string | One of: `""`, `no`, `always`, `on-failure`, `unless-stopped` |
| `replicas` | int | Must be >= 0 |
| `healthcheck` | block | `test`, `interval`, `timeout`, `retries`, `start_period` |
| `resources` | block | `cpu_limit`, `memory_limit`, `cpu_reservation`, `memory_reservation` |
| `hostname`, `working_dir`, `user`, `labels`, `networks`, `expose`, `env_file` | various | See `config.go` for types |

## Variables

Declare with `variable` blocks at the top level. Access via `var.<name>`. Currently only default values are supported — no external injection.

```hcl
variable "consul_count" { default = 3 }
```

## Locals

Computed values in a `locals` block. Can reference `var.*` and other locals (topologically sorted before evaluation). Access via `local.<name>`.

```hcl
locals {
  names    = [for i in range(var.consul_count) : format("consul-%d", i + 1)]
  join_cmd = join(" -retry-join=", local.names)
}
```

## Functions

The eval engine provides built-in functions in three categories — **string** (`upper`, `lower`, `format`, `join`, `split`, `replace`, `trimspace`, `substr`, `strlen`), **collection** (`concat`, `flatten`, `lookup`), and **numeric** (`range`). Standard HCL expressions (`for`, conditionals, interpolation `${}`) also work. See `pkg/model/eval/eval.go` for the full list and implementations.

## Dynamic Blocks

Generate multiple service blocks from an iterable. Use `each.key` and `each.value` inside `content`. The `labels` attribute controls generated block names.

```hcl
dynamic "service" {
  for_each = range(var.consul_count)
  labels   = [local.names[each.value]]
  content {
    image = "hashicorp/consul:latest"
    ports = { "850${each.value}" = "8500" }
  }
}
```

Supports lists, maps, and numbers as `for_each` values. See `qa.cluster.hcl` for a full example.

## Dependencies

Services declare ordering via `depends_on`. Values can be string literals, resource refs (`service.db`), or computed expressions (`local.names`). The dependency graph is validated for missing targets and cycles during `HCluster.IsValid()`.

## Validation Summary

- **Manifest:** max 1 cluster
- **Cluster:** name required, >= 1 service, acyclic dependency graph with no missing targets
- **Service:** exactly one of `image`/`build`, valid restart policy, non-negative replicas

## Config Loading

| Function | Input | Validates |
|---|---|---|
| `ParseManifest` / `ReadManifestFile` | string / path → `*HManifest` | No |
| `ParseClusterConfig` / `ReadClusterConfigFile` | string / path → `*HCluster` | Yes (ParseClusterConfig only) |
| `ParseLeaderConfig` / `ReadLeaderConfigFile` | string / path → `*LeaderConfig` | No |
| `ParseWorkerConfig` / `ReadWorkerConfigFile` | string / path → `*WorkerConfig` | No |

The parse pipeline runs: parse HCL → resolve variables → evaluate locals → expand dynamic blocks → extract resource refs → decode into Go structs. See `eval.BuildEvalContext` for details.
