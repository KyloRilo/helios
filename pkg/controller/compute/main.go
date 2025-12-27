package compute

import (
	"github.com/KyloRilo/helios/pkg/model"
)

type CompCtrlType int

const (
	DOCKER CompCtrlType = iota
	AWS
	GCP
)

func NewComputeController(ctrlType CompCtrlType) model.ComputeController {
	switch ctrlType {
	case DOCKER:
		return newDockerController()
	// case AWS:
	// 	return NewAWSController()
	// case GCP:
	// 	return NewGCPController()
	default:
		return newDockerController()
	}
}
