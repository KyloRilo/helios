package graph

import "sort"

type Graph struct {
	nodes map[string]bool
	edges map[string]map[string]bool // edges[A][B] = true means A depends on B
}

func New(deps map[string][]string) (*Graph, error) {
	g := &Graph{
		nodes: make(map[string]bool, len(deps)),
		edges: make(map[string]map[string]bool, len(deps)),
	}

	for name := range deps {
		g.nodes[name] = true
		g.edges[name] = make(map[string]bool)
	}

	for name, depList := range deps {
		for _, dep := range depList {
			if !g.nodes[dep] {
				return nil, &MissingDepError{Node: name, Dependency: dep}
			}
			g.edges[name][dep] = true
		}
	}

	return g, nil
}

func (g *Graph) Validate() error {
	_, err := g.TopoSort()
	return err
}

func (g *Graph) dependentsOf(name string) []string {
	var result []string
	for node, deps := range g.edges {
		if deps[name] {
			result = append(result, node)
		}
	}
	return result
}

func (g *Graph) TopoSort() ([]string, error) {
	deg := make(map[string]int, len(g.nodes))
	for name := range g.nodes {
		deg[name] = len(g.edges[name])
	}

	var queue []string
	for name, d := range deg {
		if d == 0 {
			queue = append(queue, name)
		}
	}
	sort.Strings(queue)

	var result []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		for _, dep := range g.dependentsOf(node) {
			deg[dep]--
			if deg[dep] == 0 {
				queue = append(queue, dep)
				sort.Strings(queue)
			}
		}
	}

	if len(result) != len(g.nodes) {
		var cycleNodes []string
		for name, d := range deg {
			if d > 0 {
				cycleNodes = append(cycleNodes, name)
			}
		}
		sort.Strings(cycleNodes)
		return nil, &CycleError{Nodes: cycleNodes}
	}

	return result, nil
}

func (g *Graph) Levels() ([][]string, error) {
	deg := make(map[string]int, len(g.nodes))
	for name := range g.nodes {
		deg[name] = len(g.edges[name])
	}

	var levels [][]string
	processed := 0

	for processed < len(g.nodes) {
		var level []string
		for name, d := range deg {
			if d == 0 {
				level = append(level, name)
			}
		}

		if len(level) == 0 {
			var cycleNodes []string
			for name, d := range deg {
				if d > 0 {
					cycleNodes = append(cycleNodes, name)
				}
			}
			sort.Strings(cycleNodes)
			return nil, &CycleError{Nodes: cycleNodes}
		}

		sort.Strings(level)
		levels = append(levels, level)

		for _, node := range level {
			delete(deg, node)
			for _, dep := range g.dependentsOf(node) {
				deg[dep]--
			}
		}

		processed += len(level)
	}

	return levels, nil
}

func (g *Graph) Reverse() *Graph {
	rev := &Graph{
		nodes: make(map[string]bool, len(g.nodes)),
		edges: make(map[string]map[string]bool, len(g.nodes)),
	}

	for name := range g.nodes {
		rev.nodes[name] = true
		rev.edges[name] = make(map[string]bool)
	}

	// edges[A][B] means A depends on B. Reverse: B depends on A.
	for name, deps := range g.edges {
		for dep := range deps {
			rev.edges[dep][name] = true
		}
	}

	return rev
}

func (g *Graph) Roots() []string {
	var roots []string
	for name := range g.nodes {
		if len(g.edges[name]) == 0 {
			roots = append(roots, name)
		}
	}
	sort.Strings(roots)
	return roots
}

func (g *Graph) Dependents(name string) []string {
	result := g.dependentsOf(name)
	sort.Strings(result)
	return result
}

func (g *Graph) NodeNames() []string {
	names := make([]string, 0, len(g.nodes))
	for name := range g.nodes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (g *Graph) HasNode(name string) bool {
	return g.nodes[name]
}
