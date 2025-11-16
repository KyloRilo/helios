package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/KyloRilo/helios/pkg/model"
	"github.com/KyloRilo/helios/pkg/service/core"
)

type DockerMock struct{}

func (d DockerMock) Create(ctx context.Context, conf model.ContainerConfig) (*model.ContainerMeta, error) {
	return nil, nil
}

func (d DockerMock) StartContainer(ctx context.Context, meta *model.ContainerMeta) error {
	return nil
}

func (d DockerMock) StopContainer(ctx context.Context, meta *model.ContainerMeta) error {
	return nil
}

func (d DockerMock) RemoveContainer(ctx context.Context, meta *model.ContainerMeta) error {
	return nil
}

func (d DockerMock) Log(ctx context.Context, meta *model.ContainerMeta) error {
	return nil
}

func TestCoreInit(t *testing.T) {
	_ = core.CoreService{Docker: DockerMock{}}
}

type CreateSuccess struct{ DockerMock }

func (d CreateSuccess) Create(ctx context.Context, conf model.ContainerConfig) (*model.ContainerMeta, error) {
	return &model.ContainerMeta{
		ContainerConfig: conf,
		Id:              "test-id",
	}, nil
}

func genServcConfig() model.Service {
	return model.Service{
		Name:  "test-service",
		Image: "test-image",
		Build: &model.Build{
			Context:    ".",
			Dockerfile: "./test/config/helios.hcl",
		},
	}
}

func TestCreateServicePasses(t *testing.T) {
	ctx := context.Background()
	coreService := core.CoreService{Docker: CreateSuccess{}}
	coreService.CreateService(ctx, genServcConfig())
}

type CreateFailed struct{ DockerMock }

func (d CreateFailed) Create(ctx context.Context, conf model.ContainerConfig) (*model.ContainerMeta, error) {
	return nil, fmt.Errorf("failed to create container")
}

func TestCreateServiceCreateFails(t *testing.T) {
	ctx := context.Background()
	coreService := core.CoreService{Docker: CreateFailed{}}
	err := coreService.CreateService(ctx, genServcConfig())
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}
