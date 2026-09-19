package graph

import "fmt"

type CycleError struct {
	Nodes []string
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("dependency cycle detected involving nodes: %v", e.Nodes)
}

type MissingDepError struct {
	Node       string
	Dependency string
}

func (e *MissingDepError) Error() string {
	return fmt.Sprintf("service %q depends on %q, which is not defined", e.Node, e.Dependency)
}
