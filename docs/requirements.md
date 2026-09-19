# Project Requirements

## Functional Requirements
* Core Requirements:
  * User should be able to define config for clusters and supply them to helios as hcl
    * Similar idea to a docker-compose.yml
  * Single pane of glass cluster management
    * The user can register/manage multiple helios clusters
    * High level health metrics for each cluster
  * Deployable as both local and remote compute
  * Can both run local docker builds for images as well as remote images
* Config Requirements:
  * Internal helios cluster representation (hidden to user):
    * Core services required to run helios in remote/local environments
    * Generatable based on selected type of cluster deployment
  * External user cluster config:
    * The cluster's services as the user sees it
* AI Requirements:
  * Currently read-only access to cluster metrics
  * User should provide endpoint / credentials required to hit AI
* TUI Requirements:
  * Init/register a cluster to helios
  * View various health metrics and stats of clusters
  * Exposes management operations of cluster and individual services
  * Engage with an AI agent which is fed with data regarding the cluster

## NonFunctional Requirements
* System should be resilient to single points of failure
* System should be able to self heal in the event of a node's failure
* System will both support single and mutli-leader deployments
