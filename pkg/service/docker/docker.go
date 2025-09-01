package docker

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/KyloRilo/helios/pkg/model"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type DockerService struct {
	*client.Client
	platform  string
	authToken string
}

func (serv *DockerService) Create(ctx context.Context, msg model.CreateContainer) (*container.CreateResponse, error) {
	reader, err := serv.ImagePull(
		ctx,
		msg.DockerImage,
		image.PullOptions{
			RegistryAuth: serv.authToken,
			Platform:     serv.platform,
		},
	)

	if err != nil {
		return nil, err
	}
	io.Copy(os.Stdout, reader)

	resp, err := serv.ContainerCreate(ctx, &container.Config{
		Image: msg.DockerImage,
		Cmd:   []string{"echo", "hello world"},
	}, nil, nil, nil, "")
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (serv *DockerService) Log(ctx context.Context, msg model.LogContainer) error {
	statusCh, errCh := serv.ContainerWait(ctx, msg.ContainerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := serv.ContainerLogs(ctx, msg.ContainerId, container.LogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return nil
}

func InitDockerService() DockerService {
	log.Print("Init InitDockerService...")
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	ctrl := DockerService{
		Client:    client,
		platform:  "",
		authToken: "",
	}

	return ctrl
}
