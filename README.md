# Helios (WIP)

A POC of a docker orchestrator.

**Functional Requirements**:
* User should be able to define config for projects and supply them to helios

**NonFunctional Requirements**:
* System should be resilient to single points of failure
* System should be able to self heal in the event of a node's failure

## Build


## Technologies

[Proto.Actor](https://github.com/asynkron/protoactor-go)
[Raft](https://github.com/hashicorp/raft)
[Consul](https://developer.hashicorp.com/consul)
[gRPC](https://grpc.io/)