package pkg

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func formatPort(port int) string {
	return fmt.Sprintf(":%d", port)
}

type CoreService struct {
	CloudRouter  *grpc.ClientConn
	DockerRouter *grpc.ClientConn
	ApiRouter    *grpc.ClientConn
	StateRouter  *grpc.ClientConn
}

func InitCoreService() CoreService {
	dockerConn, err := grpc.NewClient(formatPort(dockerServerPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	cloudConn, err := grpc.NewClient(formatPort(cloudServerPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	return CoreService{
		DockerRouter: dockerConn,
		CloudRouter:  cloudConn,
	}
}
