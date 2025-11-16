package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/KyloRilo/helios/pkg/controller/docker"
	"github.com/KyloRilo/helios/pkg/model"
)

var CACHE map[string]string = map[string]string{}
var SERVC docker.DockerCtrl

func initService() docker.DockerCtrl {
	return docker.InitDockerController()
}

func TestInitDockerController(t *testing.T) {
	_ = initService()
}

func initBasicContainer(ctx context.Context, servc docker.DockerCtrl) (*model.ContainerMeta, error) {
	var err error
	var meta *model.ContainerMeta
	conf := model.ContainerConfig{
		Image:    "alpine:latest",
		Name:     "basic-test",
		Hostname: "basic-test",
	}

	if meta, err = servc.Create(ctx, conf); err != nil {
		return nil, fmt.Errorf("Failed to created container => %s", err)
	}

	return meta, nil
}

func TestContainerLifecycle(t *testing.T) {
	var meta *model.ContainerMeta
	var err error
	ctx := context.Background()
	servc := initService()

	if meta, err = initBasicContainer(ctx, servc); err != nil {
		t.Errorf("Failed to created container => %s", err)
	}

	if meta.Id == "" {
		t.Errorf("ID not found on init")
	}

	t.Logf("Created Container with ID '%s' and Name '%s'", meta.Id, meta.Name)

	if err = servc.StartContainer(ctx, meta); err != nil {
		t.Errorf("Failed to start container '%s' => %s", meta.Name, err)
	}

	if err = servc.StopContainer(ctx, meta); err != nil {
		t.Errorf("Failed to stop container '%s' => %s", meta.Name, err)
	}

	if err = servc.RemoveContainer(ctx, meta); err != nil {
		t.Errorf("Failed to stop container '%s' => %s", meta.Name, err)
	}
}
