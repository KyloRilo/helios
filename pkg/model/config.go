package model

import (
	"fmt"
	"os"

	"github.com/KyloRilo/helios/pkg/model/eval"
	"github.com/KyloRilo/helios/pkg/model/graph"
	"github.com/hashicorp/hcl/v2/gohcl"
)

func readConfigFile(path string, target interface{}) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Failed to read config file: %s", err)
	}

	ctx, body, err := eval.BuildEvalContext(src, path)
	if err != nil {
		return fmt.Errorf("Failed to build eval context: %s", err)
	}

	diags := gohcl.DecodeBody(body, ctx, target)
	if diags.HasErrors() {
		fmt.Printf("HCL error: %s", diags.Error())
		return fmt.Errorf("Failed to parse config: %s", diags.Error())
	}

	return nil
}

func parseConfig(conf string, target interface{}) error {
	ctx, body, err := eval.BuildEvalContext([]byte(conf), "config.hcl")
	if err != nil {
		return fmt.Errorf("Unable to Parse Config => %s", err)
	}

	diags := gohcl.DecodeBody(body, ctx, target)
	if diags.HasErrors() {
		return fmt.Errorf("Unable to Decode Body => %s", diags.Error())
	}

	return nil
}

type Config interface {
	IsValid() (bool, error)
}

type HConfig struct{}

func (cfg HConfig) IsValid() (bool, error) {
	return false, fmt.Errorf("Validation Unimplemented")
}

type HManifest struct {
	HConfig
	Clusters []HCluster `hcl:"cluster,block"`
}

func (cfg HManifest) IsValid() (bool, error) {
	if len(cfg.Clusters) > 1 {
		return false, fmt.Errorf("Config invalid... Multi-cluster is unsupported")
	}

	return true, nil
}

type HCluster struct {
	HConfig
	Name     string     `hcl:"name,label"`
	Services []HService `hcl:"service,block"`
}

func (cfg HCluster) IsValid() (bool, error) {
	if cfg.Name == "" {
		return false, fmt.Errorf("Cluster name is required")
	}

	if len(cfg.Services) == 0 {
		return false, fmt.Errorf("At least one service is required")
	}

	for _, svc := range cfg.Services {
		if ok, err := svc.IsValid(); !ok {
			return false, fmt.Errorf("Service %s is invalid => %s", svc.Name, err)
		}
	}

	deps := make(map[string][]string, len(cfg.Services))
	for _, svc := range cfg.Services {
		deps[svc.Name] = svc.DependsOn
	}
	g, err := graph.New(deps)
	if err != nil {
		return false, fmt.Errorf("Dependency graph is invalid: %s", err)
	}
	if err := g.Validate(); err != nil {
		return false, fmt.Errorf("Dependency graph is invalid: %s", err)
	}

	return true, nil
}

type HService struct {
	HConfig
	ID          string
	Name        string            `hcl:"name,label"`
	Image       string            `hcl:"image,optional"`
	Build       *Build            `hcl:"build,block"`
	Command     string            `hcl:"command,optional"`
	Volumes     map[string]string `hcl:"volumes,optional"`
	Environment map[string]string `hcl:"environment,optional"`
	Hostname    string            `hcl:"hostname,optional"`
	Ports       map[string]string `hcl:"ports,optional"`
	Expose      []string          `hcl:"expose,optional"`
	DependsOn   []string          `hcl:"depends_on,optional"`
	Entrypoint  string            `hcl:"entrypoint,optional"`
	WorkingDir  string            `hcl:"working_dir,optional"`
	User        string            `hcl:"user,optional"`
	Labels      map[string]string `hcl:"labels,optional"`
	EnvFile     string            `hcl:"env_file,optional"`
	Restart     string            `hcl:"restart,optional"`
	Replicas    int               `hcl:"replicas,optional"`
	Networks    []string          `hcl:"networks,optional"`
	Healthcheck *Healthcheck      `hcl:"healthcheck,block"`
	Resources   *Resources        `hcl:"resources,block"`
}

var validRestartPolicies = map[string]bool{
	"":                true,
	"no":              true,
	"always":          true,
	"on-failure":      true,
	"unless-stopped":  true,
}

func (svc HService) IsValid() (bool, error) {
	if svc.Image == "" && svc.Build == nil {
		return false, fmt.Errorf("Service %s must have either an image or build configuration", svc.Name)
	}

	if svc.Image != "" && svc.Build != nil {
		return false, fmt.Errorf("Service %s cannot have both image and build configuration", svc.Name)
	}

	if !validRestartPolicies[svc.Restart] {
		return false, fmt.Errorf("Service %s has invalid restart policy '%s'", svc.Name, svc.Restart)
	}

	if svc.Replicas < 0 {
		return false, fmt.Errorf("Service %s replicas must be non-negative", svc.Name)
	}

	return true, nil
}

type Build struct {
	Context    string            `hcl:"context"`
	Dockerfile string            `hcl:"dockerfile"`
	Args       map[string]string `hcl:"args,optional"`
	Target     string            `hcl:"target,optional"`
}

type Healthcheck struct {
	Test        string `hcl:"test"`
	Interval    string `hcl:"interval,optional"`
	Timeout     string `hcl:"timeout,optional"`
	Retries     int    `hcl:"retries,optional"`
	StartPeriod string `hcl:"start_period,optional"`
}

type Resources struct {
	CPULimit          string `hcl:"cpu_limit,optional"`
	MemoryLimit       string `hcl:"memory_limit,optional"`
	CPUReservation    string `hcl:"cpu_reservation,optional"`
	MemoryReservation string `hcl:"memory_reservation,optional"`
}

type LeaderConfig struct {
	HConfig
	Name string `hcl:"name"`
	Host string `hcl:"host"`
	Port int    `hcl:"port"`
}

type WorkerConfig struct {
	HConfig
}

func ParseManifest(confStr string) (*HManifest, error) {
	manifest := &HManifest{}
	err := parseConfig(confStr, manifest)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func ReadManifestFile(path string) (*HManifest, error) {
	manifest := &HManifest{}
	err := readConfigFile(path, manifest)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func ParseClusterConfig(confStr string) (*HCluster, error) {
	cluster := &HCluster{}
	err := parseConfig(confStr, cluster)
	if err != nil {
		return nil, err
	}

	if ok, err := cluster.IsValid(); !ok {
		return nil, fmt.Errorf("Cluster config is invalid => %s", err)
	}

	return cluster, nil
}

func ReadClusterConfigFile(path string) (*HCluster, error) {
	cluster := &HCluster{}
	err := readConfigFile(path, cluster)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func ParseLeaderConfig(confStr string) (*LeaderConfig, error) {
	leaderConf := &LeaderConfig{}
	err := parseConfig(confStr, leaderConf)
	if err != nil {
		return nil, err
	}
	return leaderConf, nil
}

func ReadLeaderConfigFile(path string) (*LeaderConfig, error) {
	leaderConf := &LeaderConfig{}
	err := readConfigFile(path, leaderConf)
	if err != nil {
		return nil, err
	}
	return leaderConf, nil
}

func ParseWorkerConfig(confStr string) (*WorkerConfig, error) {
	workerConf := &WorkerConfig{}
	err := parseConfig(confStr, workerConf)
	if err != nil {
		return nil, err
	}
	return workerConf, nil
}

func ReadWorkerConfigFile(path string) (*WorkerConfig, error) {
	workerConf := &WorkerConfig{}
	err := readConfigFile(path, workerConf)
	if err != nil {
		return nil, err
	}
	return workerConf, nil
}
