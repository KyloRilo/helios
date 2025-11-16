package model

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// type HeliosConfig struct {
// 	Cluster *ClusterConfig `hcl:"cluster,optional,block"`
// 	Leader  *LeaderConfig  `hcl:"leader,optional,block"`
// 	Worker  *WorkerConfig  `hcl:"worker,optional,block"`
// }

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

type HeliosConfig struct {
	Clusters []ClusterConfig `hcl:"cluster,block"`
}

func (cfg HeliosConfig) IsValid() (bool, error) {
	if len(cfg.Clusters) > 1 {
		return false, fmt.Errorf("Config invalid... Multi-cluster is unsupported")
	}

	return true, nil
}

func (cfg HeliosConfig) ReadFile(path string) error {
	return readConfigFile(path, &cfg)
}

type ClusterConfig struct {
	HeliosConfig
	Name     string    `hcl:"name,label"`
	Host     string    `hcl:"host"`
	Port     int       `hcl:"port"`
	Services []Service `hcl:"service,block"`
}

type Service struct {
	Name        string   `hcl:"name,label"`
	Image       string   `hcl:"image,optional"`
	Build       *Build   `hcl:"build,block"`
	Command     string   `hcl:"command,optional"`
	Volumes     []string `hcl:"volumes,optional"`
	Environment []string `hcl:"environment,optional"`
	Ports       []string `hcl:"ports,optional"`
}

type Build struct {
	Context    string `hcl:"context"`
	Dockerfile string `hcl:"dockerfile"`
}

type LeaderConfig struct {
	HeliosConfig
}
type WorkerConfig struct {
	HeliosConfig
}

func ParseConfig(confStr string) (*HeliosConfig, error) {
	cfg := &HeliosConfig{}
	err := parseConfig(confStr, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func ReadConfigFile(path string) (*HeliosConfig, error) {
	cfg := &HeliosConfig{}
	err := readConfigFile(path, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
