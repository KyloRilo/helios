package model

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func readConfigFile(path string, target interface{}) error {
	err := hclsimple.DecodeFile(path, nil, target)
	if err != nil {
		// If there are HCL syntax errors, the diagnostics object will tell you
		if diags, ok := err.(hcl.Diagnostics); ok {
			fmt.Println("HCL error: %s", diags.Error())
		}
		return fmt.Errorf("Failed to parse config: %s", err)
	}

	return nil
}

func parseConfig(conf string, target interface{}) error {
	var diags hcl.Diagnostics
	var file *hcl.File

	file, diags = hclsyntax.ParseConfig([]byte(conf), "", hcl.InitialPos)
	if diags.HasErrors() {
		return fmt.Errorf("Unable to Parse Config => %s", diags.Error())
	}

	diags = gohcl.DecodeBody(file.Body, nil, target)
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
	Host     string     `hcl:"host"`
	Port     int        `hcl:"port"`
	Services []HService `hcl:"service,block"`
}

type HService struct {
	HConfig
	Name        string   `hcl:"name,label"`
	ID          string   `hcl:"id,optional"`
	Image       string   `hcl:"image,optional"`
	Build       *Build   `hcl:"build,block"`
	Command     string   `hcl:"command,optional"`
	Volumes     []string `hcl:"volumes,optional"`
	Environment []string `hcl:"environment,optional"`
	Ports       []string `hcl:"ports,optional"`
	DependsOn   []string `hcl:"depends_on,optional"`
}

type Build struct {
	Context    string `hcl:"context"`
	Dockerfile string `hcl:"dockerfile"`
}

type LeaderConfig struct {
	HConfig
	Name string `hcl:"name,label"`
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
