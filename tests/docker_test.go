package tests

import (
	"context"
	"testing"

	"github.com/KyloRilo/helios/models"
	"github.com/KyloRilo/helios/pkg"
)

type BadMsgTest struct {
	models.Message
}

func spinupDockerService() models.IChannel {
	serv := pkg.InitDockerService()
	go serv.Listen(context.Background(), serv.MsgHandler)

	return serv
}

func TestInitDockerService(t *testing.T) {
	spinupDockerService()
}

func TestBadMsgType(t *testing.T) {
	serv := spinupDockerService()
	err := serv.MsgHandler(context.Background(), BadMsgTest{})
	if err == nil {
		t.Errorf("Expected error throw")
	}
}

func TestCreateMsg(t *testing.T) {
	serv := spinupDockerService()
	err := serv.MsgHandler(context.Background(), models.CreateContainer{})
	if err != nil {
		t.Errorf("Error Raised during CreateContainer => %s", err)
	}
}
