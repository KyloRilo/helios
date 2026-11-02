package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	compCtrl "github.com/KyloRilo/helios/pkg/controller/compute"
	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/model/compute"
	"github.com/KyloRilo/helios/pkg/model/graph"
	"github.com/KyloRilo/helios/pkg/service/core"
)

const usage = `helios - AI-first Docker/compute orchestrator

Usage:
  helios <command> [options]

Commands:
  validate <manifest>           Validate a cluster manifest file
  plan <current> <desired>      Show execution plan between two manifests
  graph <manifest>              Show the service dependency graph
  up <manifest>                 Create and start a cluster
  down <manifest>               Stop and teardown a running cluster
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "validate":
		err = runValidate(args)
	case "plan":
		err = runPlan(args)
	case "graph":
		err = runGraph(args)
	case "up":
		err = runUp(args)
	case "down":
		err = runDown(args)
	case "help", "-h", "--help":
		fmt.Print(usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		fmt.Print(usage)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func loadManifest(path string) (*model.HManifest, error) {
	manifest, err := model.ReadManifestFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest %s: %w", path, err)
	}
	return manifest, nil
}

func clusterFromManifest(manifest *model.HManifest) (*model.HCluster, error) {
	if _, err := manifest.IsValid(); err != nil {
		return nil, err
	}
	if len(manifest.Clusters) == 0 {
		return nil, fmt.Errorf("manifest contains no clusters")
	}
	return &manifest.Clusters[0], nil
}

func runValidate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: helios validate <manifest>")
	}

	manifest, err := loadManifest(args[0])
	if err != nil {
		return err
	}

	cluster, err := clusterFromManifest(manifest)
	if err != nil {
		return err
	}

	if _, err := cluster.IsValid(); err != nil {
		return fmt.Errorf("cluster %q is invalid: %w", cluster.Name, err)
	}

	fmt.Printf("cluster %q is valid (%d services)\n", cluster.Name, len(cluster.Services))

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "SERVICE\tIMAGE\tBUILD\tDEPENDS ON")
	for _, svc := range cluster.Services {
		img := svc.Image
		build := "-"
		if svc.Build != nil {
			img = "-"
			build = svc.Build.Dockerfile
		}
		deps := "-"
		if len(svc.DependsOn) > 0 {
			deps = strings.Join(svc.DependsOn, ", ")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", svc.Name, img, build, deps)
	}
	w.Flush()

	return nil
}

func runGraph(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: helios graph <manifest>")
	}

	manifest, err := loadManifest(args[0])
	if err != nil {
		return err
	}

	cluster, err := clusterFromManifest(manifest)
	if err != nil {
		return err
	}

	if _, err := cluster.IsValid(); err != nil {
		return fmt.Errorf("cluster %q is invalid: %w", cluster.Name, err)
	}

	deps := make(map[string][]string, len(cluster.Services))
	for _, svc := range cluster.Services {
		deps[svc.Name] = svc.DependsOn
	}

	topo, err := graph.New(deps)
	if err != nil {
		return fmt.Errorf("building topology: %w", err)
	}

	levels, err := topo.Levels()
	if err != nil {
		return fmt.Errorf("computing levels: %w", err)
	}

	fmt.Printf("cluster %q dependency graph (%d services, %d levels):\n\n", cluster.Name, len(cluster.Services), len(levels))
	for i, level := range levels {
		fmt.Printf("  level %d: %s\n", i, strings.Join(level, ", "))
	}

	return nil
}

func buildClusterGraph(cluster *model.HCluster) (*graph.ClusterGraph, error) {
	nodes := make([]*compute.Node, 0, len(cluster.Services))
	deps := make(map[string][]string, len(cluster.Services))

	for _, svc := range cluster.Services {
		deps[svc.Name] = svc.DependsOn
		opts := []compute.NodeOption{
			compute.WithName(svc.Name),
			compute.WithCmd(svc.Command),
			compute.WithEntrypoint(svc.Entrypoint),
			compute.WithPorts(svc.Ports),
			compute.WithVolumes(svc.Volumes),
			compute.WithEnv(svc.Environment),
			compute.WithHostname(svc.Hostname),
			compute.WithDependsOn(svc.DependsOn),
		}
		if svc.Build != nil {
			opts = append(opts, compute.WithContext(&compute.Context{
				Path:   svc.Build.Context,
				File:   svc.Build.Dockerfile,
				Args:   svc.Build.Args,
				Target: svc.Build.Target,
			}))
		} else {
			opts = append(opts, compute.WithImage(svc.Image))
		}
		nodes = append(nodes, compute.NewNode(opts...))
	}

	topo, err := graph.New(deps)
	if err != nil {
		return nil, fmt.Errorf("building topology: %w", err)
	}

	return graph.NewClusterGraph(topo, nodes)
}

func runPlan(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: helios plan <current-manifest> <desired-manifest>")
	}

	currentManifest, err := loadManifest(args[0])
	if err != nil {
		return fmt.Errorf("current: %w", err)
	}
	currentCluster, err := clusterFromManifest(currentManifest)
	if err != nil {
		return fmt.Errorf("current: %w", err)
	}

	desiredManifest, err := loadManifest(args[1])
	if err != nil {
		return fmt.Errorf("desired: %w", err)
	}
	desiredCluster, err := clusterFromManifest(desiredManifest)
	if err != nil {
		return fmt.Errorf("desired: %w", err)
	}

	currentGraph, err := buildClusterGraph(currentCluster)
	if err != nil {
		return fmt.Errorf("building current graph: %w", err)
	}

	targets := make([]graph.DiffTarget, len(desiredCluster.Services))
	desiredDeps := make(map[string][]string, len(desiredCluster.Services))
	for i, svc := range desiredCluster.Services {
		targets[i] = graph.DiffTarget{
			Name:       svc.Name,
			Image:      svc.Image,
			Command:    svc.Command,
			Entrypoint: svc.Entrypoint,
			Restart:    svc.Restart,
			Replicas:   svc.Replicas,
			DependsOn:  svc.DependsOn,
		}
		desiredDeps[svc.Name] = svc.DependsOn
	}

	plan, err := graph.Diff(targets, desiredDeps, currentGraph)
	if err != nil {
		return fmt.Errorf("computing plan: %w", err)
	}

	if !plan.HasChanges() {
		fmt.Println("No changes detected.")
		return nil
	}

	fmt.Printf("Plan: %s\n\n", plan.Summary())

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "LEVEL\tSERVICE\tOP\tREASON")
	for i, level := range plan.Levels() {
		for _, action := range level {
			reason := action.Reason
			if reason == "" {
				reason = "-"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", i, action.Name, action.Op, reason)
		}
	}
	w.Flush()

	return nil
}

func newCoreService(cluster *model.HCluster) (core.CoreService, error) {
	ctx := context.Background()
	cs := core.NewCoreService(ctx, core.CoreArgs{
		Conf:       cluster,
		ScalerArgs: compCtrl.ControllerArgs{},
	})
	return cs, nil
}

func saveState(state *model.ClusterState, nodes []*compute.Node) {
	state.UpdateNodes(nodes)
	if err := state.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to save state: %s\n", err)
	}
}

func runUp(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: helios up <manifest>")
	}

	manifest, err := loadManifest(args[0])
	if err != nil {
		return err
	}

	cluster, err := clusterFromManifest(manifest)
	if err != nil {
		return err
	}

	if _, err := model.LoadState(cluster.Name); err == nil {
		return fmt.Errorf("cluster %q already has state — run 'helios down %s' first", cluster.Name, args[0])
	}

	cs, err := newCoreService(cluster)
	if err != nil {
		return err
	}

	ctx := context.Background()

	state := &model.ClusterState{
		Cluster:  cluster.Name,
		Network:  "helios-" + cluster.Name,
		Manifest: args[0],
	}

	fmt.Printf("Creating cluster %q...\n", cluster.Name)
	createErr := cs.CreateCluster(ctx)
	if cs.GetNodes() != nil {
		saveState(state, cs.GetNodes())
	}
	if createErr != nil {
		return fmt.Errorf("creating cluster: %w", createErr)
	}

	fmt.Printf("Starting cluster %q...\n", cluster.Name)
	startErr := cs.StartCluster(ctx)
	saveState(state, cs.GetNodes())
	if startErr != nil {
		return fmt.Errorf("starting cluster: %w", startErr)
	}

	nodes := cs.GetNodes()
	fmt.Printf("Cluster %q is up (%d nodes)\n", cluster.Name, len(nodes))

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NODE\tID\tSTATUS")
	for _, n := range nodes {
		id := n.Id
		if len(id) > 12 {
			id = id[:12]
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", n.Name, id, n.Status)
	}
	w.Flush()

	return nil
}

func runDown(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: helios down <manifest>")
	}

	manifest, err := loadManifest(args[0])
	if err != nil {
		return err
	}

	cluster, err := clusterFromManifest(manifest)
	if err != nil {
		return err
	}

	state, err := model.LoadState(cluster.Name)
	if err != nil {
		return fmt.Errorf("no state found for cluster %q — nothing to tear down", cluster.Name)
	}

	deps := make(map[string][]string, len(state.Nodes))
	nodes := make([]*compute.Node, 0, len(state.Nodes))
	for _, ns := range state.Nodes {
		deps[ns.Name] = ns.DependsOn
		nodes = append(nodes, compute.NewNode(
			compute.WithName(ns.Name),
			compute.WithId(ns.Id),
			compute.WithStatus(ns.Status),
			compute.WithDependsOn(ns.DependsOn),
		))
	}

	topo, err := graph.New(deps)
	if err != nil {
		return fmt.Errorf("rebuilding topology from state: %w", err)
	}

	cg, err := graph.NewClusterGraph(topo, nodes)
	if err != nil {
		return fmt.Errorf("rebuilding graph from state: %w", err)
	}

	cs, err := newCoreService(cluster)
	if err != nil {
		return err
	}
	cs.SetGraph(cg)

	ctx := context.Background()

	fmt.Printf("Stopping cluster %q...\n", cluster.Name)
	if err := cs.StopCluster(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning during stop: %s\n", err)
	}
	saveState(state, cs.GetNodes())

	fmt.Printf("Tearing down cluster %q...\n", cluster.Name)
	if err := cs.TeardownCluster(ctx); err != nil {
		saveState(state, cs.GetNodes())
		return fmt.Errorf("tearing down cluster: %w", err)
	}

	if err := model.RemoveState(cluster.Name); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to remove state file: %s\n", err)
	}

	fmt.Printf("Cluster %q is down\n", cluster.Name)
	return nil
}
