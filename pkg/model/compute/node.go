package compute

import "fmt"

type OptionsApplier interface {
	Apply(def Node) error
}

type NodeOption func(def *Node)

func (opt NodeOption) Apply(def *Node) {
	opt(def)
}

type Context struct {
	Path   string
	File   string
	Args   map[string]string
	Target string
}

type Status string

const (
	Init      Status = "INIT"
	Ready     Status = "READY"
	Created   Status = "CREATED"
	Updating  Status = "UPDATING"
	Up        Status = "UP"
	Down      Status = "DOWN"
	Destroyed Status = "DESTROYED"
	Error     Status = "ERROR"
)

type Ports map[string]string

func (p Ports) ToStringArray() []string {
	ports := []string{}
	for portPub, portPriv := range p {
		ports = append(ports, fmt.Sprintf("%s:%s/tcp", portPub, portPriv))
	}

	return ports
}

type Healthcheck struct {
	Test        string
	Interval    string
	Timeout     string
	Retries     int
	StartPeriod string
}

type Resources struct {
	CPULimit          string
	MemoryLimit       string
	CPUReservation    string
	MemoryReservation string
}

type Node struct {
	Id          string
	Image       string
	Name        string
	Tags        []string
	Ports       Ports
	Cmd         string
	Env         map[string]string
	Volumes     map[string]string
	Context     *Context
	Status      Status
	Hostname    string
	Restart     string
	Replicas    int
	Networks    []string
	Healthcheck *Healthcheck
	Resources   *Resources
	Entrypoint  string
	WorkingDir  string
	User        string
	Labels      map[string]string
	EnvFile     string
	Expose      []string
	DependsOn   []string
}

func (n Node) AddTags(tag ...string) {
	if n.Tags == nil {
		n.Tags = []string{}
	}

	n.Tags = append(n.Tags, tag...)
}

func WithId(id string) NodeOption {
	return func(node *Node) {
		node.Id = id
	}
}

func WithImage(img string) NodeOption {
	return func(node *Node) {
		node.Image = img
	}
}

func WithName(name string) NodeOption {
	return func(node *Node) {
		node.Name = name
	}
}

func WithCmd(cmd string) NodeOption {
	return func(node *Node) {
		node.Cmd = cmd
	}
}

func WithPorts(ports map[string]string) NodeOption {
	return func(node *Node) {
		node.Ports = ports
	}
}
func WithEnv(env map[string]string) NodeOption {
	return func(node *Node) {
		node.Env = env
	}
}

func WithVolumes(vols map[string]string) NodeOption {
	return func(node *Node) {
		node.Volumes = vols
	}
}

func WithContext(ctx *Context) NodeOption {
	return func(node *Node) {
		node.Context = ctx
	}
}

func WithTags(tags []string) NodeOption {
	return func(node *Node) {
		node.Tags = tags
	}
}

func WithStatus(stat Status) NodeOption {
	return func(node *Node) {
		node.Status = stat
	}
}

func WithHostname(hostname string) NodeOption {
	return func(node *Node) {
		node.Hostname = hostname
	}
}

func WithRestart(policy string) NodeOption {
	return func(node *Node) {
		node.Restart = policy
	}
}

func WithReplicas(n int) NodeOption {
	return func(node *Node) {
		node.Replicas = n
	}
}

func WithNetworks(networks []string) NodeOption {
	return func(node *Node) {
		node.Networks = networks
	}
}

func WithHealthcheck(hc *Healthcheck) NodeOption {
	return func(node *Node) {
		node.Healthcheck = hc
	}
}

func WithResources(res *Resources) NodeOption {
	return func(node *Node) {
		node.Resources = res
	}
}

func WithEntrypoint(ep string) NodeOption {
	return func(node *Node) {
		node.Entrypoint = ep
	}
}

func WithWorkingDir(dir string) NodeOption {
	return func(node *Node) {
		node.WorkingDir = dir
	}
}

func WithUser(user string) NodeOption {
	return func(node *Node) {
		node.User = user
	}
}

func WithLabels(labels map[string]string) NodeOption {
	return func(node *Node) {
		node.Labels = labels
	}
}

func WithEnvFile(path string) NodeOption {
	return func(node *Node) {
		node.EnvFile = path
	}
}

func WithExpose(ports []string) NodeOption {
	return func(node *Node) {
		node.Expose = ports
	}
}

func WithDependsOn(deps []string) NodeOption {
	return func(node *Node) {
		node.DependsOn = deps
	}
}

func NewNode(opts ...NodeOption) *Node {
	node := &Node{
		Status: Init,
	}

	for _, opt := range opts {
		opt.Apply(node)
	}

	node.Status = Ready
	return node
}
