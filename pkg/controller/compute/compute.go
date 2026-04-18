package compute

import (
	"context"
	"strings"

	"github.com/KyloRilo/helios/pkg/model/node"
)

type ComputeController interface {
	CreateNode(context.Context, *node.Node) (*node.CreateNodeResp, error)
	StartNode(context.Context, *node.Node) (*node.StartNodeResp, error)
	ListNodes(context.Context) (*node.ListNodesResp, error)
	StopNode(context.Context, *node.Node) (*node.StopNodeResp, error)
	RemoveNode(context.Context, *node.Node) (*node.RmNodeResp, error)
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

type Provider string

const (
	ProviderDocker Provider = "docker"
	ProviderECR    Provider = "ecr"
)

type ControllerArgs struct {
	Stub *ComputeController
	// DockerCreds *DockerCreds
	AwsCreds *AwsCreds
}

func NewComputeController(ctx context.Context, args ControllerArgs, stub *ComputeController) (ComputeController, error) {
	switch {
	case args.Stub != nil:
		return *args.Stub, nil
	case args.AwsCreds != nil:
		return newAwsCtrl(ctx, *args.AwsCreds), nil
	default:
		return newDockerCtrl(ctx)
	}
}
