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

const NodeActionCreate = "Create"
const NodeActionStart = "Start"
const NodeActionStop = "Stop"
const NodeActionRemove = "Remove"

func IsValidProvider(p Provider) bool {
	switch p {
	case ProviderDocker, ProviderECR, ProviderGCR:
		return true
	default:
		return false
	}
}

func isValidStatus(n *compute.Node, action string, expected ...compute.Status) error {
	if slices.Contains(expected, n.GetStatus()) {
		return nil
	}

	return errors.InvalidNodeStatus{
		NodeName:   n.GetName(),
		NodeStatus: string(n.GetStatus()),
		Action:     action,
		Expected: func() []string {
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

	if err := isValidStatus(n, NodeActionCreate, compute.Ready, compute.Destroyed); err != nil {
		return "", err
	}
	print("Creating node")

	if id, err = c.CtrlShim.createNode(ctx, n); err != nil {
		print("Setting status to error")
		n.SetStatus(compute.Error)
		return "", err
	}

	print("Setting status to created")
	n.SetId(id)
	n.SetStatus(compute.Created)
	return id, nil
}

func (c CompImpl) StartNode(ctx context.Context, n *compute.Node) error {
	if err := isValidStatus(n, NodeActionStart, compute.Created, compute.Down); err != nil {
		return err
	}

	err := c.CtrlShim.startNode(ctx, n)
	if err != nil {
		n.SetStatus(compute.Error)
		return err
	}

	n.SetStatus(compute.Up)
	return nil
}

func (c CompImpl) ListNodes(ctx context.Context) ([]*compute.Node, error) {
	return c.CtrlShim.listNodes(ctx)
}

func (c CompImpl) StopNode(ctx context.Context, n *compute.Node) error {
	if err := isValidStatus(n, NodeActionStop, compute.Up); err != nil {
		return err
	}

	err := c.CtrlShim.stopNode(ctx, n)
	if err != nil {
		n.SetStatus(compute.Error)
		return err
	}

	n.SetStatus(compute.Down)
	return nil
}

func (c CompImpl) RemoveNode(ctx context.Context, n *compute.Node) error {
	if err := isValidStatus(n, NodeActionRemove, compute.Created, compute.Down); err != nil {
		return err
	}

	err := c.CtrlShim.removeNode(ctx, n)
	if err != nil {
		n.SetStatus(compute.Error)
		return err
	}

	n.SetStatus(compute.Destroyed)
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
	Stub *CtrlShim
	// DockerCreds *DockerCreds
	AwsCreds *AwsCreds
}

func NewComputeController(ctx context.Context, args ControllerArgs) (ComputeController, error) {
	var ctrl CtrlShim
	var err error

	switch {
	case args.Stub != nil:
		ctrl, err = *args.Stub, nil
	case args.AwsCreds != nil:
		ctrl, err = newAwsCtrl(ctx, *args.AwsCreds), nil
	default:
		ctrl, err = newDockerCtrl(ctx)
	}

	return CompImpl{CtrlShim: ctrl}, err
}
