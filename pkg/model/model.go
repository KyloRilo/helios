package model

import "github.com/asynkron/protoactor-go/actor"

type InstanceConfig struct{}

type ContainerConfig struct {
	Name     string
	Image    string
	Hostname string
	Ports    []string
	Build    *Build
}
type ContainerMeta struct {
	ContainerConfig
	Id string
}

type State struct{}

type ActorService interface {
	GetServiceName() string
	Receive(actor.Context)
}
