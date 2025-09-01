# Helios (WIP)

A POC of a docker orchestrator.

**Functional Requirements**:
* User should be able to define config for projects and supply them to helios

**NonFunctional Requirements**:
* System should be resilient to single points of failure
* System should be able to self heal in the event of a node's failure

## Build
Spin up the cluster with make via `make cluster-up`.