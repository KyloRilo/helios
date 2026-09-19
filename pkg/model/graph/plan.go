package graph

import (
	"fmt"
	"strings"

	"github.com/KyloRilo/helios/pkg/model/compute"
)

type Op string

const (
	OpCreate  Op = "create"
	OpUpdate  Op = "update"
	OpDestroy Op = "destroy"
	OpNoop    Op = "noop"
)

type PlanAction struct {
	Name   string
	Node   *compute.Node
	Op     Op
	Reason string
}

type Plan struct {
	Actions []PlanAction
	levels  [][]PlanAction
}

func (p *Plan) Levels() [][]PlanAction {
	return p.levels
}

func (p *Plan) HasChanges() bool {
	for _, a := range p.Actions {
		if a.Op != OpNoop {
			return true
		}
	}
	return false
}

func (p *Plan) Summary() string {
	counts := map[Op]int{}
	for _, a := range p.Actions {
		counts[a.Op]++
	}
	return fmt.Sprintf("create %d, update %d, destroy %d, noop %d",
		counts[OpCreate], counts[OpUpdate], counts[OpDestroy], counts[OpNoop])
}

type DiffTarget struct {
	Name       string
	Image      string
	Command    string
	Entrypoint string
	Restart    string
	Replicas   int
	DependsOn  []string
}

func Diff(desired []DiffTarget, desiredDeps map[string][]string, current *ClusterGraph) (*Plan, error) {
	topo, err := New(desiredDeps)
	if err != nil {
		return nil, fmt.Errorf("invalid desired topology: %w", err)
	}
	if err := topo.Validate(); err != nil {
		return nil, fmt.Errorf("desired topology has cycle: %w", err)
	}

	desiredMap := make(map[string]DiffTarget, len(desired))
	for _, d := range desired {
		desiredMap[d.Name] = d
	}

	actionMap := make(map[string]PlanAction)

	for _, d := range desired {
		existing := current.Get(d.Name)
		if existing == nil {
			actionMap[d.Name] = PlanAction{Name: d.Name, Op: OpCreate, Reason: "new service"}
			continue
		}

		reasons := diffNode(d, existing)
		if len(reasons) > 0 {
			actionMap[d.Name] = PlanAction{
				Name:   d.Name,
				Node:   existing,
				Op:     OpUpdate,
				Reason: strings.Join(reasons, ", "),
			}
		} else {
			actionMap[d.Name] = PlanAction{Name: d.Name, Node: existing, Op: OpNoop}
		}
	}

	for _, name := range current.Topology().NodeNames() {
		if _, ok := desiredMap[name]; !ok {
			actionMap[name] = PlanAction{
				Name:   name,
				Node:   current.Get(name),
				Op:     OpDestroy,
				Reason: "service removed",
			}
		}
	}

	levels, err := topo.Levels()
	if err != nil {
		return nil, err
	}

	var allActions []PlanAction
	var planLevels [][]PlanAction

	for _, level := range levels {
		var levelActions []PlanAction
		for _, name := range level {
			if a, ok := actionMap[name]; ok {
				allActions = append(allActions, a)
				levelActions = append(levelActions, a)
			}
		}
		if len(levelActions) > 0 {
			planLevels = append(planLevels, levelActions)
		}
	}

	// Destroyed services aren't in the desired topology, append them at the end
	for name, a := range actionMap {
		if a.Op == OpDestroy {
			allActions = append(allActions, a)
			planLevels = append(planLevels, []PlanAction{a})
			_ = name
		}
	}

	return &Plan{Actions: allActions, levels: planLevels}, nil
}

func diffNode(desired DiffTarget, current *compute.Node) []string {
	var reasons []string
	if desired.Image != "" && desired.Image != current.Image {
		reasons = append(reasons, fmt.Sprintf("image: %s → %s", current.Image, desired.Image))
	}
	if desired.Command != "" && desired.Command != current.Cmd {
		reasons = append(reasons, fmt.Sprintf("command: %s → %s", current.Cmd, desired.Command))
	}
	if desired.Entrypoint != "" && desired.Entrypoint != current.Entrypoint {
		reasons = append(reasons, fmt.Sprintf("entrypoint: %s → %s", current.Entrypoint, desired.Entrypoint))
	}
	if desired.Restart != "" && desired.Restart != current.Restart {
		reasons = append(reasons, fmt.Sprintf("restart: %s → %s", current.Restart, desired.Restart))
	}
	if desired.Replicas != 0 && desired.Replicas != current.Replicas {
		reasons = append(reasons, fmt.Sprintf("replicas: %d → %d", current.Replicas, desired.Replicas))
	}
	return reasons
}
