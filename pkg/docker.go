package pkg

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/KyloRilo/helios/models"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type DockerService struct {
	*client.Client
	models.ChannelService
	platform  string
	authToken string
}

func (serv *DockerService) Create(ctx context.Context, msg models.CreateContainer) (models.IMessage, error) {
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

func (serv *DockerService) Start(ctx context.Context, msg models.StartContainer) (models.IMessage, error) {
	if err := serv.ContainerStart(ctx, msg.ContainerId, container.StartOptions{}); err != nil {
		panic(err)
	}
}

func (serv *DockerService) Log(ctx context.Context, msg models.LogContainer) (models.IMessage, error) {
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
}

func (serv *DockerService) List(ctx context.Context, msg models.ListContainers) models.IMessage {
	var resp models.IMessage
	ls, err := serv.ContainerList(ctx, container.ListOptions{
		Size:   msg.Size,
		All:    msg.All,
		Latest: msg.Latest,
		Since:  msg.Since,
	})

	if err != nil {
		resp = models.ErrorMessage{
			BaseErr: err,
		}
	} else {
		// containers := make([]models.Container, len(ls))
		resp = models.ListContainerResp{
			Containers: []models.Container{{}},
		}
	}

	return resp
}

func (serv DockerService) MsgHandler(ctx context.Context, msg models.IMessage) models.IMessage {
	var resp models.IMessage
	var err error
	log.Print("DockerService.Receive() => Received: ", msg)
	switch req := msg.(type) {
	case models.CreateContainer:
		resp, err = serv.Create(ctx, req)
	case models.StartContainer:
		resp, err = serv.Start(ctx, req)
	case models.LogContainer:
		resp, err = serv.Log(ctx, req)
	default:
		err = fmt.Errorf("DockerService.Receive() => Unhandled message type")
	}

	if err != nil {
		resp = models.ErrorMessage{
			BaseErr: err,
		}
	}

	return resp
}

func InitDockerService() models.IChannel {
	log.Print("Init DockerService...")
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	serv := DockerService{
		Client:         client,
		platform:       "",
		authToken:      "",
		ChannelService: *models.NewChannelService(),
	}

	return serv
}
