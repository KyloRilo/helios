package test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/KyloRilo/helios/pkg/model"
)

type BadMsgTest struct {
	model.Message
}

type GoodMsgTest struct {
	model.Message
}

type SuccessResp struct {
	model.Message
}

type TestService struct {
	model.ChannelService
}

func (serv TestService) MsgHandler(ctx context.Context, msg model.IMessage) model.IMessage {
	var resp model.IMessage
	var err error
	log.Print("TestService.MsgHandler() => Received: ", msg)
	switch msg.(type) {
	case GoodMsgTest:
		resp, err = SuccessResp{}, nil
	default:
		err = fmt.Errorf("TestService.MsgHandler() => Unhandled message type")
	}

	if err != nil {
		resp = model.ErrorMsgResp{
			BaseErr: err,
		}
	}

	return resp
}

var testService *TestService

func initTestService() model.IChannel {
	if testService == nil {
		testService = &TestService{
			ChannelService: *model.NewChannelService(),
		}

		go testService.Listen(context.Background(), testService.MsgHandler)
	}

	return testService
}

func TestInitDockerService(t *testing.T) {
	initTestService()
}

func TestBadMsgType(t *testing.T) {
	serv := initTestService()
	err := serv.MsgHandler(context.Background(), BadMsgTest{})
	if err == nil {
		t.Errorf("Expected error throw")
	}
}

func TestCreateMsg(t *testing.T) {
	serv := initTestService()
	err := serv.MsgHandler(context.Background(), GoodMsgTest{})
	if err != nil {
		t.Errorf("Error Raised during CreateContainer => %s", err)
	}
}
