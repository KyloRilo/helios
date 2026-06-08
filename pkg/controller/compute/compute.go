package compute

import (
	"context"
	"slices"
	"strings"

	"github.com/KyloRilo/helios/pkg/model/compute"
	"github.com/KyloRilo/helios/pkg/model/errors"
)

type Provider string

const (
	ProviderTest   Provider = "test"
	ProviderDocker Provider = "docker"
	ProviderECR    Provider = "ecr"
	ProviderGCR    Provider = "gcr"
)

type NodeAction string

const (
	NodeActionCreate NodeAction = "Create"
	NodeActionStart  NodeAction = "Start"
	NodeActionStop   NodeAction = "Stop"
	NodeActionRemove NodeAction = "Remove"
)

func (n NodeAction) getExpectedStatuses() []compute.Status {
	switch n {
	case NodeActionCreate:
		return []compute.Status{compute.Ready, compute.Destroyed}
	case NodeActionStart:
		return []compute.Status{compute.Created, compute.Down}
	case NodeActionStop:
		return []compute.Status{compute.Up}
	case NodeActionRemove:
		return []compute.Status{compute.Created, compute.Down}
	default:
		return nil
	}
}

func (n NodeAction) isValidStatus(node *compute.Node) bool {
	return slices.Contains(n.getExpectedStatuses(), node.Status)
}

func IsValidProvider(p Provider) bool {
	switch p {
	case ProviderDocker, ProviderECR, ProviderGCR:
		return true
	default:
		return false
	}
}

func isValidStatus(n *compute.Node, act NodeAction) error {
	if act.isValidStatus(n) {
		return nil
	}

	return errors.InvalidNodeStatus{
		NodeName:   n.Name,
		NodeStatus: string(n.Status),
		Action:     string(act),
		Expected: func() []string {
			expected := act.getExpectedStatuses()
			strs := make([]string, len(expected))
			for i, s := range expected {
				strs[i] = string(s)
			}
			return strs
		}(),
	}
}

type CtrlShim interface {
	createNode(context.Context, *compute.Node) (string, error)
	startNode(context.Context, *compute.Node) error
	listNodes(context.Context) ([]*compute.Node, error)
	stopNode(context.Context, *compute.Node) error
	removeNode(context.Context, *compute.Node) error
}

type ComputeController interface {
	GetProvider() Provider
	CreateNode(context.Context, *compute.Node) (string, error)
	StartNode(context.Context, *compute.Node) error
	ListNodes(context.Context) ([]*compute.Node, error)
	StopNode(context.Context, *compute.Node) error
	RemoveNode(context.Context, *compute.Node) error
}

type CompImpl struct {
	CtrlShim
	provider Provider
}

func (c CompImpl) GetProvider() Provider {
	return c.provider
}

func (c CompImpl) CreateNode(ctx context.Context, n *compute.Node) (string, error) {
	var id string
	var err error

	if err := isValidStatus(n, NodeActionCreate); err != nil {
		return "", err
	}

	if id, err = c.CtrlShim.createNode(ctx, n); err != nil {
		n.Status = compute.Error
		return "", err
	}

	n.Id = id
	n.Status = compute.Created
	return id, nil
}

func (c CompImpl) StartNode(ctx context.Context, n *compute.Node) error {
	if err := isValidStatus(n, NodeActionStart); err != nil {
		return err
	}

	err := c.CtrlShim.startNode(ctx, n)
	if err != nil {
		n.Status = compute.Error
		return err
	}

	n.Status = compute.Up
	return nil
}

func (c CompImpl) ListNodes(ctx context.Context) ([]*compute.Node, error) {
	return c.CtrlShim.listNodes(ctx)
}

func (c CompImpl) StopNode(ctx context.Context, n *compute.Node) error {
	if err := isValidStatus(n, NodeActionStop); err != nil {
		return err
	}

	err := c.CtrlShim.stopNode(ctx, n)
	if err != nil {
		n.Status = compute.Error
		return err
	}

	n.Status = compute.Down
	return nil
}

func (c CompImpl) RemoveNode(ctx context.Context, n *compute.Node) error {
	if err := isValidStatus(n, NodeActionRemove); err != nil {
		return err
	}

	err := c.CtrlShim.removeNode(ctx, n)
	if err != nil {
		n.Status = compute.Error
		return err
	}

	n.Status = compute.Destroyed
	return nil
}

func isECR(image string) bool {
	return strings.Contains(image, ".dkr.ecr.") &&
		strings.Contains(image, ".amazonaws.com/")
}

func isDockerHub(image string) bool {
	if strings.HasPrefix(image, "docker.io/") {
		return true
	}

	firstSlash := strings.Index(image, "/")
	if firstSlash == -1 {
		return true
	}

	return !strings.Contains(image[:firstSlash], ".")
}

type ControllerArgs struct {
	stub *CtrlShim
	// DockerCreds *DockerCreds
	AwsCreds *AwsCreds
}

func NewComputeController(ctx context.Context, args ControllerArgs) (ComputeController, error) {
	var ctrl CtrlShim
	var err error

	switch {
	case args.stub != nil:
		ctrl, err = *args.stub, nil
	case args.AwsCreds != nil:
		ctrl, err = newAwsCtrl(ctx, *args.AwsCreds), nil
	default:
		ctrl, err = newDockerCtrl(ctx)
	}

	return CompImpl{CtrlShim: ctrl}, err
}
