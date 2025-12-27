package compute

import (
	"fmt"
	"strings"

	"github.com/KyloRilo/helios/pkg/model"
)

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

func NewComputeController(image string, stub *model.ComputeController) (model.ComputeController, error) {
	switch {
	// case isECR(image):
	// 	return newECRController()
	// case strings.HasPrefix(image, "gcr.io/"):
	// 	return newGCRController()
	case isDockerHub(image):
		return newDockerController()
	case stub != nil:
		return *stub, nil
	default:
		return nil, fmt.Errorf("Unable to map generate controller from uri '%s'", image)
	}
}
