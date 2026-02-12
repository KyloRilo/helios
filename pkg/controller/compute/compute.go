package compute

import (
	"context"
	"strings"

	"github.com/KyloRilo/helios/pkg/model"
)

type ComputeController interface {
	Authenticate(context.Context) (interface{}, error)
	CreateNode(ctx context.Context, n *model.Node) (interface{}, error)
	StartNode(ctx context.Context, n *model.Node) (interface{}, error)
	StopNode(ctx context.Context, n *model.Node) (interface{}, error)
	RemoveNode(ctx context.Context, n *model.Node) (interface{}, error)
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

func NewComputeController(stub *ComputeController) ComputeController {
	ctrl := *stub
	if ctrl == nil {
		ctrl = newDockerCtrl()
	}

	return ctrl
}
