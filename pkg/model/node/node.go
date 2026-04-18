package node

import "fmt"

type Definition struct{}
type OptionsApplier interface {
	Apply(def Node) error
}

type NodeOption func(def *Node)

func (opt NodeOption) Apply(def *Node) {
	opt(def)
}

type Context struct {
	Path string
	File string
}

type Status string

const (
	Up   Status = "UP"
	Down Status = "DOWN"
)

type Ports map[string]string

func (p Ports) ToStringArray() []string {
	ports := []string{}
	for portPub, portPriv := range p {
		ports = append(ports, fmt.Sprintf("%s:%s/tcp", portPub, portPriv))
	}

	return ports
}

type Node struct {
	id      string
	image   string
	name    string
	tags    []string
	ports   Ports
	cmd     string
	env     map[string]string
	volumes map[string]string
	context *Context
	status  Status
}

func (n Node) GetId() string                 { return n.id }
func (n Node) SetId(id string)               { n.id = id }
func (n Node) GetImage() string              { return n.image }
func (n Node) SetImage(img string)           { n.image = img }
func (n Node) GetName() string               { return n.name }
func (n Node) GetCmd() string                { return n.cmd }
func (n Node) GetPorts() Ports               { return n.ports }
func (n Node) GetEnv() map[string]string     { return n.env }
func (n Node) GetVolumes() map[string]string { return n.volumes }
func (n Node) GetContext() *Context          { return n.context }
func (n Node) SetContext(ctx *Context)       { n.context = ctx }
func (n Node) AddTag(tag string) {
	if n.tags == nil {
		n.tags = []string{}
	}

	n.tags = append(n.tags, tag)
}

func WithId(id string) NodeOption {
	return func(node *Node) {
		node.id = id
	}
}

func WithImage(img string) NodeOption {
	return func(node *Node) {
		node.image = img
	}
}

func WithName(name string) NodeOption {
	return func(node *Node) {
		node.name = name
	}
}

func WithCmd(cmd string) NodeOption {
	return func(node *Node) {
		node.cmd = cmd
	}
}

func WithPorts(ports map[string]string) NodeOption {
	return func(node *Node) {
		node.ports = ports
	}
}
func WithEnv(env map[string]string) NodeOption {
	return func(node *Node) {
		node.env = env
	}
}

func WithVolumes(vols map[string]string) NodeOption {
	return func(node *Node) {
		node.volumes = vols
	}
}

func WithContext(ctx *Context) NodeOption {
	return func(node *Node) {
		node.context = ctx
	}
}

func WithStatus(stat Status) NodeOption {
	return func(node *Node) {
		node.status = stat
	}
}

func NewNode(opts ...NodeOption) *Node {
	node := &Node{}
	for _, opt := range opts {
		opt.Apply(node)
	}

	print(node)
	return node
}

type CreateNodeResp struct {
	Node *Node
}
type StartNodeResp struct {
	Node *Node
}
type ListNodesResp struct {
	Nodes []*Node
}
type StopNodeResp struct {
	Node *Node
}

type RmNodeResp struct {
	Node *Node
}
