package pkg

import (
	"context"
	"io"
	"log"
	"net"
	"os"

	"github.com/KyloRilo/helios/proto"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type DockerService struct {
	*client.Client
	platform  string
	authToken string
}

func (cmgr *DockerService) init(client *client.Client, _ interface{}) {
	cmgr.Client = client
	cmgr.platform = ""
	cmgr.authToken = ""
}

func (cmgr *DockerService) Create(ctx context.Context, dockerImage string) (*container.CreateResponse, error) {
	reader, err := cmgr.ImagePull(
		ctx,
		dockerImage,
		image.PullOptions{
			RegistryAuth: cmgr.authToken,
			Platform:     cmgr.platform,
		},
	)

	if err != nil {
		return nil, err
	}
	io.Copy(os.Stdout, reader)

	resp, err := cmgr.ContainerCreate(ctx, &container.Config{
		Image: dockerImage,
		Cmd:   []string{"echo", "hello world"},
	}, nil, nil, nil, "")
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (cmgr *DockerService) Start(ctx context.Context, containerId string) {
	if err := cmgr.ContainerStart(ctx, containerId, container.StartOptions{}); err != nil {
		panic(err)
	}
}

func (cmgr *DockerService) Listen(ctx context.Context, containerId string) {
	statusCh, errCh := cmgr.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cmgr.ContainerLogs(ctx, containerId, container.LogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

//
// gRPC impl
//

var dockerService *DockerService

const dockerServerPort int = 50501

type DockerServer struct {
	proto.UnimplementedDockerServer
}

func (s *DockerServer) BuildImage(ctx context.Context, in *proto.BuildReq) (*proto.BuildResp, error) {
	// dockerService.ImageBuild(ctx)
	return nil, nil
}

func (s *DockerServer) CreateContainer(ctx context.Context, in *proto.CreateReq) (*proto.CreateResp, error) {
	dockerService.Create(ctx, in.Image)
	return nil, nil
}

func (s *DockerServer) StartContainer(_ context.Context, in *proto.StartReq) (*proto.StartResp, error) {
	return nil, nil
}

func InitDockerService() {
	log.Print("Running Docker gRPC server")
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	dockerService.init(client, nil)
	lis, err := net.Listen("tcp", formatPort(dockerServerPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterDockerServer(s, &DockerServer{})
	reflection.Register(s)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
