package model

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
)

type IMessage interface {
	setId(string)
	getId() string
	setSrc(chan IMessage)
	getSrc() chan IMessage
	setTimestamp(time.Time)
	getTimestamp() time.Time
}

type Message struct {
	id        string
	timestamp time.Time
	src       chan IMessage
}

func (m Message) setId(id string) {
	m.id = id
}

func (m Message) getId() string {
	return m.id
}

func (m Message) setSrc(src chan IMessage) {
	m.src = src
}

func (m Message) getSrc() chan IMessage {
	return m.src
}

func (m Message) setTimestamp(ts time.Time) {
	m.timestamp = ts
}

func (m Message) getTimestamp() time.Time {
	return m.timestamp
}

type ErrorMsgResp struct {
	Message
	BaseErr error
}

type CreateContainer struct {
	Message
	DockerImage string
}

type StartContainer struct {
	Message
	ContainerId string
}

type LogContainer StartContainer

type ListContainers struct {
	Message
	Size   bool
	All    bool
	Latest bool
	Since  string
	Before string
	Limit  int
}

type IChannel interface {
	Send(IMessage, chan IMessage)
	MsgHandler(context.Context, IMessage) IMessage
	Listen(context.Context, func(context.Context, IMessage) IMessage)
	GetChannel() chan IMessage
}

type ChannelService struct {
	channel chan IMessage
}

func (c ChannelService) GetChannel() chan IMessage {
	return c.channel
}

func (c ChannelService) Send(req IMessage, target chan IMessage) {
	req.setId(uuid.NewString())
	req.setSrc(c.GetChannel())
	req.setTimestamp(time.Now())
	target <- req
}

func (c ChannelService) Listen(ctx context.Context, handler func(context.Context, IMessage) IMessage) {
	for {
		msg := <-c.GetChannel()
		log.Printf("Received Message => Message: %s", msg)
		resp := handler(ctx, msg)
		if resp != nil {
			c.Send(resp, msg.getSrc())
		}
	}
}

// ! Saved example of a MsgHandler() impl
// func (serv DockerController) MsgHandler(ctx context.Context, msg models.IMessage) models.IMessage {
// 	var resp models.IMessage
// 	var err error
// 	log.Print("DockerController.Receive() => Received: ", msg)
// 	switch req := msg.(type) {
// 	case models.CreateContainer:
// 		resp, err = serv.Create(ctx, req)
// 	case models.StartContainer:
// 		resp, err = serv.Start(ctx, req)
// 	case models.LogContainer:
// 		resp, err = serv.Log(ctx, req)
// 	default:
// 		err = fmt.Errorf("DockerController.Receive() => Unhandled message type")
// 	}

// 	if err != nil {
// 		resp = models.ErrorMessage{
// 			BaseErr: err,
// 		}
// 	}

// 	return resp
// }

func NewChannelService() *ChannelService {
	return &ChannelService{
		channel: make(chan IMessage),
	}
}
