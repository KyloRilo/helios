package docker

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/KyloRilo/helios/pkg/model"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type DockerCtrl interface {
	Create(context.Context, model.ContainerConfig) (*model.ContainerMeta, error)
	StartContainer(context.Context, *model.ContainerMeta) error
	StopContainer(context.Context, *model.ContainerMeta) error
	RemoveContainer(context.Context, *model.ContainerMeta) error
	Log(context.Context, *model.ContainerMeta) error
}

type DockerController struct {
	client    *client.Client
	platform  string
	authToken string
}

func (d *DockerController) pullImage(ctx context.Context, img string) error {
	reader, err := d.client.ImagePull(
		ctx,
		img,
		image.PullOptions{
			RegistryAuth: d.authToken,
			Platform:     d.platform,
		},
	)

	if err != nil {
		return err
	}
	defer reader.Close()

	return nil
}

func (d *DockerController) buildContainer(ctx context.Context, dockerfile string) error {
	resp, err := d.client.ImageBuild(ctx, nil, types.ImageBuildOptions{
		Dockerfile: dockerfile,
		Remove:     true,
	})

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (d DockerController) Create(ctx context.Context, conf model.ContainerConfig) (*model.ContainerMeta, error) {
	if conf.Build != nil {
		// TODO: Implement build logic here
		if err := d.buildContainer(ctx, conf.Build.Dockerfile); err != nil {
			return nil, fmt.Errorf("[DockerController.Create()] Unable to build docker image => %s", err)
		}
	}

	if err := d.pullImage(ctx, conf.Image); err != nil {
		return nil, fmt.Errorf("[DockerController.Create()] Unable to pull docker image => %s", err)
	}

	name := conf.Name
	if name == "" {
		parts := strings.Split(conf.Image, ":") // Split image name and tag
		name = fmt.Sprintf("%s-helios", parts[0])
	}

	resp, err := d.client.ContainerCreate(ctx, &container.Config{
		Image: conf.Image,
		Cmd:   []string{"sleep", "infinity"},
	}, nil, &network.NetworkingConfig{}, nil, name)
	if err != nil {
		return nil, err
	}

	log.Println("Found warnings on create => ", resp.Warnings)
	conf.Name = name
	return &model.ContainerMeta{
		Id:              resp.ID,
		ContainerConfig: conf,
	}, nil
}

func (d DockerController) StartContainer(ctx context.Context, meta *model.ContainerMeta) error {
	return d.client.ContainerStart(ctx, meta.Id, container.StartOptions{})
}

func (d DockerController) Log(ctx context.Context, meta *model.ContainerMeta) error {
	statusCh, errCh := d.client.ContainerWait(ctx, meta.Id, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := d.client.ContainerLogs(ctx, meta.Id, container.LogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return nil
}

func (d DockerController) StopContainer(ctx context.Context, meta *model.ContainerMeta) error {
	return d.client.ContainerStop(ctx, meta.Id, container.StopOptions{})
}

func (d DockerController) RemoveContainer(ctx context.Context, meta *model.ContainerMeta) error {
	return d.client.ContainerRemove(ctx, meta.Id, container.RemoveOptions{Force: true})
}

func InitDockerController() DockerCtrl {
	log.Print("Init InitDockerController...")
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	ctrl := DockerController{
		client:    client,
		platform:  "",
		authToken: "",
	}

	return ctrl
}
