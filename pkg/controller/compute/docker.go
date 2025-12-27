package compute

import (
	"context"
	"log"
	"os"

	"github.com/KyloRilo/helios/pkg/model"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type DockerController struct {
	client    *client.Client
	platform  string
	authToken string
}

func (d DockerController) pullImage(ctx context.Context, img string) error {
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

func (d DockerController) buildImage(ctx context.Context, dockerfile string) error {
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

func (d DockerController) logContainer(ctx context.Context, node *model.Node) error {
	statusCh, errCh := d.client.ContainerWait(ctx, node.Meta.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := d.client.ContainerLogs(ctx, node.Meta.ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return nil
}

func (d DockerController) Authenticate(context.Context) error {
	return nil
}

func (d DockerController) CreateNode(ctx context.Context, node *model.Node) error {
	meta := node.Meta
	resp, err := d.client.ContainerCreate(ctx, &container.Config{
		Image: meta.Image,
		Cmd:   []string{"sleep", "infinity"},
	}, nil, &network.NetworkingConfig{}, nil, meta.Name)
	if err != nil {
		return err
	}

	if len(resp.Warnings) > 0 {
		log.Println("Found warnings on create => ", resp.Warnings)
	}

	node.Meta.ID = resp.ID
	return nil
}

func (d DockerController) StartNode(ctx context.Context, node *model.Node) error {
	return d.client.ContainerStart(ctx, node.Meta.ID, container.StartOptions{})
}

func (d DockerController) StopNode(ctx context.Context, node *model.Node) error {
	return d.client.ContainerStop(ctx, node.Meta.ID, container.StopOptions{})
}

func (d DockerController) RemoveNode(ctx context.Context, node *model.Node) error {
	return d.client.ContainerRemove(ctx, node.Meta.ID, container.RemoveOptions{Force: true})
}

func newDockerController() model.ComputeController {
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
