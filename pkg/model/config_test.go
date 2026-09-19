package model

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestHConfigIsValid(t *testing.T) {
	cfg := HConfig{}
	ok, err := cfg.IsValid()
	if ok {
		t.Error("expected HConfig.IsValid to return false")
	}
	if err == nil {
		t.Error("expected error for unimplemented validation")
	}
}

func TestHManifestIsValid(t *testing.T) {
	t.Run("single cluster", func(t *testing.T) {
		m := HManifest{Clusters: []HCluster{{Name: "c1"}}}
		ok, err := m.IsValid()
		if !ok || err != nil {
			t.Errorf("expected valid but got ok=%v err=%v", ok, err)
		}
	})

	t.Run("multi cluster rejected", func(t *testing.T) {
		m := HManifest{Clusters: []HCluster{{Name: "c1"}, {Name: "c2"}}}
		ok, _ := m.IsValid()
		if ok {
			t.Error("expected multi-cluster to be invalid")
		}
	})

	t.Run("empty clusters", func(t *testing.T) {
		m := HManifest{}
		ok, err := m.IsValid()
		if !ok || err != nil {
			t.Errorf("expected empty manifest to be valid but got ok=%v err=%v", ok, err)
		}
	})
}

func TestHClusterIsValid(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		c := HCluster{Name: "test", Services: []HService{{Name: "svc", Image: "nginx"}}}
		ok, err := c.IsValid()
		if !ok || err != nil {
			t.Errorf("expected valid but got ok=%v err=%v", ok, err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		c := HCluster{Services: []HService{{Name: "svc", Image: "nginx"}}}
		ok, _ := c.IsValid()
		if ok {
			t.Error("expected empty name to be invalid")
		}
	})

	t.Run("no services", func(t *testing.T) {
		c := HCluster{Name: "test"}
		ok, _ := c.IsValid()
		if ok {
			t.Error("expected no services to be invalid")
		}
	})

	t.Run("invalid service propagates", func(t *testing.T) {
		c := HCluster{Name: "test", Services: []HService{{Name: "bad"}}}
		ok, _ := c.IsValid()
		if ok {
			t.Error("expected invalid service to make cluster invalid")
		}
	})

	t.Run("missing dependency rejected", func(t *testing.T) {
		c := HCluster{
			Name: "test",
			Services: []HService{
				{Name: "web", Image: "nginx", DependsOn: []string{"db"}},
			},
		}
		ok, _ := c.IsValid()
		if ok {
			t.Error("expected missing dependency to be invalid")
		}
	})

	t.Run("cycle rejected", func(t *testing.T) {
		c := HCluster{
			Name: "test",
			Services: []HService{
				{Name: "a", Image: "nginx", DependsOn: []string{"b"}},
				{Name: "b", Image: "nginx", DependsOn: []string{"a"}},
			},
		}
		ok, _ := c.IsValid()
		if ok {
			t.Error("expected cycle to be invalid")
		}
	})

	t.Run("valid dependency graph", func(t *testing.T) {
		c := HCluster{
			Name: "test",
			Services: []HService{
				{Name: "db", Image: "postgres"},
				{Name: "api", Image: "nginx", DependsOn: []string{"db"}},
			},
		}
		ok, err := c.IsValid()
		if !ok || err != nil {
			t.Errorf("expected valid but got ok=%v err=%v", ok, err)
		}
	})
}

func TestHServiceIsValid(t *testing.T) {
	t.Run("image only", func(t *testing.T) {
		s := HService{Name: "svc", Image: "nginx"}
		ok, err := s.IsValid()
		if !ok || err != nil {
			t.Errorf("expected valid: %v", err)
		}
	})

	t.Run("build only", func(t *testing.T) {
		s := HService{Name: "svc", Build: &Build{Context: ".", Dockerfile: "Dockerfile"}}
		ok, err := s.IsValid()
		if !ok || err != nil {
			t.Errorf("expected valid: %v", err)
		}
	})

	t.Run("neither image nor build", func(t *testing.T) {
		s := HService{Name: "svc"}
		ok, _ := s.IsValid()
		if ok {
			t.Error("expected invalid with no image or build")
		}
	})

	t.Run("both image and build", func(t *testing.T) {
		s := HService{Name: "svc", Image: "nginx", Build: &Build{Context: ".", Dockerfile: "Dockerfile"}}
		ok, _ := s.IsValid()
		if ok {
			t.Error("expected invalid with both image and build")
		}
	})

	t.Run("invalid restart policy", func(t *testing.T) {
		s := HService{Name: "svc", Image: "nginx", Restart: "bogus"}
		ok, _ := s.IsValid()
		if ok {
			t.Error("expected invalid restart policy")
		}
	})

	t.Run("negative replicas", func(t *testing.T) {
		s := HService{Name: "svc", Image: "nginx", Replicas: -1}
		ok, _ := s.IsValid()
		if ok {
			t.Error("expected invalid with negative replicas")
		}
	})
}

func TestParseClusterConfig(t *testing.T) {
	t.Run("validation failure empty name", func(t *testing.T) {
		_, err := ParseClusterConfig(`
			service "web" {
				image = "nginx"
			}
		`)
		if err == nil {
			t.Error("expected validation error for empty cluster name")
		}
	})

	t.Run("parse error", func(t *testing.T) {
		_, err := ParseClusterConfig(`{{{invalid`)
		if err == nil {
			t.Error("expected parse error")
		}
	})
}

func TestParseManifestError(t *testing.T) {
	_, err := ParseManifest(`{{{invalid`)
	if err == nil {
		t.Error("expected parse error for invalid HCL")
	}
}

func TestParseLeaderConfigInModel(t *testing.T) {
	conf, err := ParseLeaderConfig(`
		name = "leader-1"
		host = "10.0.0.1"
		port = 6330
	`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if conf.Name != "leader-1" || conf.Host != "10.0.0.1" || conf.Port != 6330 {
		t.Errorf("unexpected config: %+v", conf)
	}
}

func TestParseLeaderConfigError(t *testing.T) {
	_, err := ParseLeaderConfig(`{{{invalid`)
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestParseWorkerConfigInModel(t *testing.T) {
	conf, err := ParseWorkerConfig(``)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if conf == nil {
		t.Error("expected config but got nil")
	}
}

func TestParseWorkerConfigError(t *testing.T) {
	_, err := ParseWorkerConfig(`{{{invalid`)
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestNewBaseActorService(t *testing.T) {
	svc := NewBaseActorService("TestService")
	if svc == nil {
		t.Fatal("expected service but got nil")
	}
	if svc.GetServiceName() != "TestService" {
		t.Errorf("expected name 'TestService' but got '%s'", svc.GetServiceName())
	}
}

func TestBaseServiceReceiveNoOp(t *testing.T) {
	svc := NewBaseActorService("Test")
	svc.Receive(nil)
}

func TestReadFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(cwd, "../../bin/helios/local.cluster.hcl")
	_, err = ReadManifestFile(path)
	if err != nil {
		t.Errorf("TestConfigRead() => %s", err)
	}
}

func TestReadQaClusterFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(cwd, "../../bin/helios/qa.cluster.hcl")
	conf, err := ReadManifestFile(path)
	if err != nil {
		t.Fatalf("TestReadQaClusterFile() => %s", err)
	}

	if len(conf.Clusters) != 1 {
		t.Fatalf("expected 1 cluster but got %d", len(conf.Clusters))
	}
	services := conf.Clusters[0].Services
	if len(services) != 8 {
		t.Fatalf("expected 8 services but got %d", len(services))
	}

	for i := 0; i < 3; i++ {
		expected := fmt.Sprintf("consul-%d", i+1)
		if services[i].Name != expected {
			t.Errorf("service[%d]: expected %q but got %q", i, expected, services[i].Name)
		}
	}

	worker := services[5]
	if worker.Name != "worker" {
		t.Errorf("expected last service 'worker' but got %q", worker.Name)
	}
	if len(worker.DependsOn) != 4 {
		t.Errorf("expected worker to have 4 deps but got %d: %v", len(worker.DependsOn), worker.DependsOn)
	}
}

func TestClusterConfig(t *testing.T) {
	tests := []string{
		`cluster "test" {
			service "test" {
				image = ""
				command = "echo hello"
				volumes = {
					"/data":"/data"
				}
				environment = {
					"ENV":"test"
				}
				ports = {
				    "8080":"80"
				}
			}
		}`,
		`cluster "test" {
			service "test" {
				build {
					context = "."
					dockerfile = "Dockerfile"
				}
			}
			service "test2" {
				image = "nginx:latest"
				depends_on = ["test"]
			}
		}`,
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Logf("\n%s", test)
			conf, err := ParseManifest(test)
			if err != nil {
				t.Error(err)
			}
			t.Logf("Parsed Config: %+v", conf)
		})
	}
}

func TestServiceOrchestrationConfig(t *testing.T) {
	conf, err := ParseManifest(`cluster "test" {
		service "web" {
			image = "nginx:latest"
			restart = "always"
			replicas = 3
			networks = ["frontend", "backend"]

			healthcheck {
				test = "curl -f http://localhost/"
				interval = "30s"
				timeout = "10s"
				retries = 3
				start_period = "5s"
			}

			resources {
				cpu_limit = "0.5"
				memory_limit = "512m"
				cpu_reservation = "0.25"
				memory_reservation = "256m"
			}
		}
	}`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	svc := conf.Clusters[0].Services[0]

	if svc.Restart != "always" {
		t.Errorf("expected restart 'always' but got '%s'", svc.Restart)
	}

	if svc.Replicas != 3 {
		t.Errorf("expected replicas 3 but got %d", svc.Replicas)
	}

	if len(svc.Networks) != 2 || svc.Networks[0] != "frontend" || svc.Networks[1] != "backend" {
		t.Errorf("expected networks [frontend, backend] but got %v", svc.Networks)
	}

	if svc.Healthcheck == nil {
		t.Fatal("expected healthcheck but got nil")
	}
	if svc.Healthcheck.Test != "curl -f http://localhost/" {
		t.Errorf("expected healthcheck test 'curl -f http://localhost/' but got '%s'", svc.Healthcheck.Test)
	}
	if svc.Healthcheck.Interval != "30s" {
		t.Errorf("expected healthcheck interval '30s' but got '%s'", svc.Healthcheck.Interval)
	}
	if svc.Healthcheck.Retries != 3 {
		t.Errorf("expected healthcheck retries 3 but got %d", svc.Healthcheck.Retries)
	}

	if svc.Resources == nil {
		t.Fatal("expected resources but got nil")
	}
	if svc.Resources.CPULimit != "0.5" {
		t.Errorf("expected cpu_limit '0.5' but got '%s'", svc.Resources.CPULimit)
	}
	if svc.Resources.MemoryLimit != "512m" {
		t.Errorf("expected memory_limit '512m' but got '%s'", svc.Resources.MemoryLimit)
	}
}

func TestServiceRestartValidation(t *testing.T) {
	validPolicies := []string{"", "no", "always", "on-failure", "unless-stopped"}
	for _, policy := range validPolicies {
		svc := HService{Name: "test", Image: "nginx", Restart: policy}
		if ok, err := svc.IsValid(); !ok {
			t.Errorf("expected restart policy '%s' to be valid but got: %v", policy, err)
		}
	}

	svc := HService{Name: "test", Image: "nginx", Restart: "invalid-policy"}
	if ok, _ := svc.IsValid(); ok {
		t.Error("expected restart policy 'invalid-policy' to be invalid")
	}
}

func TestServiceOptionalOrchestrationFields(t *testing.T) {
	conf, err := ParseManifest(`cluster "test" {
		service "minimal" {
			image = "nginx:latest"
		}
	}`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	svc := conf.Clusters[0].Services[0]

	if svc.Restart != "" {
		t.Errorf("expected empty restart but got '%s'", svc.Restart)
	}
	if svc.Replicas != 0 {
		t.Errorf("expected 0 replicas but got %d", svc.Replicas)
	}
	if svc.Healthcheck != nil {
		t.Errorf("expected nil healthcheck but got %+v", svc.Healthcheck)
	}
	if svc.Resources != nil {
		t.Errorf("expected nil resources but got %+v", svc.Resources)
	}
	if len(svc.Networks) != 0 {
		t.Errorf("expected empty networks but got %v", svc.Networks)
	}
	if svc.Entrypoint != "" {
		t.Errorf("expected empty entrypoint but got '%s'", svc.Entrypoint)
	}
	if svc.WorkingDir != "" {
		t.Errorf("expected empty working_dir but got '%s'", svc.WorkingDir)
	}
	if svc.User != "" {
		t.Errorf("expected empty user but got '%s'", svc.User)
	}
	if svc.EnvFile != "" {
		t.Errorf("expected empty env_file but got '%s'", svc.EnvFile)
	}
	if len(svc.Labels) != 0 {
		t.Errorf("expected empty labels but got %v", svc.Labels)
	}
	if len(svc.Expose) != 0 {
		t.Errorf("expected empty expose but got %v", svc.Expose)
	}
}

func TestServiceHighValueConfig(t *testing.T) {
	conf, err := ParseManifest(`cluster "test" {
		service "api" {
			image = "myapp:latest"
			entrypoint = "/docker-entrypoint.sh"
			command = "serve --port 8080"
			working_dir = "/app"
			user = "node:node"
			env_file = ".env.production"
			expose = ["3000", "9090"]
			labels = {
				"app" = "myapi"
				"tier" = "backend"
			}
		}
	}`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	svc := conf.Clusters[0].Services[0]

	if svc.Entrypoint != "/docker-entrypoint.sh" {
		t.Errorf("expected entrypoint '/docker-entrypoint.sh' but got '%s'", svc.Entrypoint)
	}
	if svc.WorkingDir != "/app" {
		t.Errorf("expected working_dir '/app' but got '%s'", svc.WorkingDir)
	}
	if svc.User != "node:node" {
		t.Errorf("expected user 'node:node' but got '%s'", svc.User)
	}
	if svc.EnvFile != ".env.production" {
		t.Errorf("expected env_file '.env.production' but got '%s'", svc.EnvFile)
	}
	if len(svc.Expose) != 2 || svc.Expose[0] != "3000" || svc.Expose[1] != "9090" {
		t.Errorf("expected expose [3000, 9090] but got %v", svc.Expose)
	}
	if len(svc.Labels) != 2 || svc.Labels["app"] != "myapi" || svc.Labels["tier"] != "backend" {
		t.Errorf("expected labels {app:myapi, tier:backend} but got %v", svc.Labels)
	}
}

func TestManifestWithVariables(t *testing.T) {
	conf, err := ParseManifest(`
variable "image_tag" {
  default = "1.21"
}

variable "app_port" {
  default = "8080"
}

cluster "test" {
  service "web" {
    image = "nginx:${var.image_tag}"
    ports = {
      "${var.app_port}" = "80"
    }
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	svc := conf.Clusters[0].Services[0]
	if svc.Image != "nginx:1.21" {
		t.Errorf("expected image 'nginx:1.21' but got %q", svc.Image)
	}
	if svc.Ports["8080"] != "80" {
		t.Errorf("expected port mapping 8080:80 but got %v", svc.Ports)
	}
}

func TestManifestWithLocals(t *testing.T) {
	conf, err := ParseManifest(`
variable "project" {
  default = "helios"
}

variable "env" {
  default = "staging"
}

locals {
  prefix   = "${var.project}-${var.env}"
  hostname = "${local.prefix}.example.com"
}

cluster "test" {
  service "api" {
    image    = "myapp:latest"
    hostname = local.hostname
    labels = {
      "app" = local.prefix
    }
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	svc := conf.Clusters[0].Services[0]
	if svc.Hostname != "helios-staging.example.com" {
		t.Errorf("expected hostname 'helios-staging.example.com' but got %q", svc.Hostname)
	}
	if svc.Labels["app"] != "helios-staging" {
		t.Errorf("expected label app='helios-staging' but got %q", svc.Labels["app"])
	}
}

func TestManifestWithFunctions(t *testing.T) {
	conf, err := ParseManifest(`
variable "name" {
  default = "helios"
}

locals {
  upper_name = upper(var.name)
  greeting   = format("Hello from %s!", local.upper_name)
}

cluster "test" {
  service "web" {
    image   = "nginx:latest"
    command = local.greeting
    hostname = lower(format("%s-HOST", var.name))
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	svc := conf.Clusters[0].Services[0]
	if svc.Command != "Hello from HELIOS!" {
		t.Errorf("expected command 'Hello from HELIOS!' but got %q", svc.Command)
	}
	if svc.Hostname != "helios-host" {
		t.Errorf("expected hostname 'helios-host' but got %q", svc.Hostname)
	}
}

func TestManifestVariablesInBuildBlock(t *testing.T) {
	conf, err := ParseManifest(`
variable "build_target" {
  default = "production"
}

variable "node_env" {
  default = "production"
}

cluster "test" {
  service "app" {
    build {
      context    = "."
      dockerfile = "Dockerfile"
      target     = var.build_target
      args = {
        "NODE_ENV" = var.node_env
      }
    }
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	build := conf.Clusters[0].Services[0].Build
	if build.Target != "production" {
		t.Errorf("expected target 'production' but got %q", build.Target)
	}
	if build.Args["NODE_ENV"] != "production" {
		t.Errorf("expected NODE_ENV='production' but got %q", build.Args["NODE_ENV"])
	}
}

func TestManifestVariablesInOrchestrationFields(t *testing.T) {
	conf, err := ParseManifest(`
variable "restart_policy" {
  default = "unless-stopped"
}

variable "replica_count" {
  default = 3
}

cluster "test" {
  service "web" {
    image    = "nginx:latest"
    restart  = var.restart_policy
    replicas = var.replica_count
    networks = ["frontend", "backend"]

    healthcheck {
      test     = format("curl -f http://localhost:%s/", "8080")
      interval = "30s"
    }
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	svc := conf.Clusters[0].Services[0]
	if svc.Restart != "unless-stopped" {
		t.Errorf("expected restart 'unless-stopped' but got %q", svc.Restart)
	}
	if svc.Replicas != 3 {
		t.Errorf("expected replicas 3 but got %d", svc.Replicas)
	}
	if svc.Healthcheck.Test != "curl -f http://localhost:8080/" {
		t.Errorf("expected healthcheck test with interpolated port but got %q", svc.Healthcheck.Test)
	}
}

func TestManifestMultiServiceWithSharedVariables(t *testing.T) {
	conf, err := ParseManifest(`
variable "env" {
  default = "prod"
}

variable "registry" {
  default = "gcr.io/myproject"
}

locals {
  api_image = "${var.registry}/api:${var.env}"
  web_image = "${var.registry}/web:${var.env}"
}

cluster "test" {
  service "db" {
    image = "postgres:16"
  }

  service "api" {
    image      = local.api_image
    depends_on = ["db"]
    environment = {
      "APP_ENV" = var.env
    }
  }

  service "web" {
    image      = local.web_image
    depends_on = ["api"]
    environment = {
      "API_URL" = format("http://api:%s", "3000")
    }
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	cluster := conf.Clusters[0]
	if len(cluster.Services) != 3 {
		t.Fatalf("expected 3 services but got %d", len(cluster.Services))
	}

	api := cluster.Services[1]
	if api.Image != "gcr.io/myproject/api:prod" {
		t.Errorf("expected api image 'gcr.io/myproject/api:prod' but got %q", api.Image)
	}
	if api.Environment["APP_ENV"] != "prod" {
		t.Errorf("expected APP_ENV='prod' but got %q", api.Environment["APP_ENV"])
	}

	web := cluster.Services[2]
	if web.Image != "gcr.io/myproject/web:prod" {
		t.Errorf("expected web image 'gcr.io/myproject/web:prod' but got %q", web.Image)
	}
	if web.Environment["API_URL"] != "http://api:3000" {
		t.Errorf("expected API_URL='http://api:3000' but got %q", web.Environment["API_URL"])
	}
}

func TestParseClusterConfigWithVariables(t *testing.T) {
	conf, err := ParseManifest(`
variable "svc_image" {
  default = "redis:7"
}

cluster "my-cluster" {
  service "cache" {
    image = var.svc_image
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	svc := conf.Clusters[0].Services[0]
	if svc.Image != "redis:7" {
		t.Errorf("expected image 'redis:7' but got %q", svc.Image)
	}
}

func TestLeaderConfigWithVariables(t *testing.T) {
	conf, err := ParseLeaderConfig(`
variable "leader_host" {
  default = "10.0.0.1"
}

variable "leader_port" {
  default = 6330
}

name = "leader-1"
host = var.leader_host
port = var.leader_port
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if conf.Host != "10.0.0.1" {
		t.Errorf("expected host '10.0.0.1' but got %q", conf.Host)
	}
	if conf.Port != 6330 {
		t.Errorf("expected port 6330 but got %d", conf.Port)
	}
}

func TestManifestServiceResourceRefs(t *testing.T) {
	conf, err := ParseManifest(`
cluster "test" {
  service "db" {
    image = "postgres:16"
  }

  service "cache" {
    image      = "redis:7"
    depends_on = [service.db]
  }

  service "api" {
    image      = "node:18"
    depends_on = [service.db, service.cache]
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	cluster := conf.Clusters[0]
	cache := cluster.Services[1]
	if len(cache.DependsOn) != 1 || cache.DependsOn[0] != "db" {
		t.Errorf("expected cache depends_on [db] but got %v", cache.DependsOn)
	}

	api := cluster.Services[2]
	if len(api.DependsOn) != 2 {
		t.Fatalf("expected api depends_on length 2 but got %d", len(api.DependsOn))
	}
	if api.DependsOn[0] != "db" || api.DependsOn[1] != "cache" {
		t.Errorf("expected api depends_on [db, cache] but got %v", api.DependsOn)
	}
}

func TestManifestServiceResourceRefsWithHyphens(t *testing.T) {
	conf, err := ParseManifest(`
cluster "test" {
  service "consul-server-1" {
    image = "consul:1.15"
  }

  service "consul-server-2" {
    image      = "consul:1.15"
    depends_on = [service.consul-server-1]
  }

  service "app-worker" {
    image      = "myapp:latest"
    depends_on = [service.consul-server-1, service.consul-server-2]
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	cluster := conf.Clusters[0]
	consul2 := cluster.Services[1]
	if len(consul2.DependsOn) != 1 || consul2.DependsOn[0] != "consul-server-1" {
		t.Errorf("expected consul-server-2 depends_on [consul-server-1] but got %v", consul2.DependsOn)
	}

	worker := cluster.Services[2]
	if len(worker.DependsOn) != 2 {
		t.Fatalf("expected app-worker depends_on length 2 but got %d", len(worker.DependsOn))
	}
	if worker.DependsOn[0] != "consul-server-1" || worker.DependsOn[1] != "consul-server-2" {
		t.Errorf("expected app-worker depends_on [consul-server-1, consul-server-2] but got %v", worker.DependsOn)
	}
}

func TestManifestResourceRefsWithVarsAndLocals(t *testing.T) {
	conf, err := ParseManifest(`
variable "registry" {
  default = "gcr.io/myproject"
}

locals {
  api_image = "${var.registry}/api:latest"
}

cluster "production" {
  service "db" {
    image = "postgres:16"
  }

  service "api" {
    image      = local.api_image
    depends_on = [service.db]
    environment = {
      "DB_HOST" = service.db
    }
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	api := conf.Clusters[0].Services[1]
	if api.Image != "gcr.io/myproject/api:latest" {
		t.Errorf("expected image 'gcr.io/myproject/api:latest' but got %q", api.Image)
	}
	if len(api.DependsOn) != 1 || api.DependsOn[0] != "db" {
		t.Errorf("expected depends_on [db] but got %v", api.DependsOn)
	}
	if api.Environment["DB_HOST"] != "db" {
		t.Errorf("expected DB_HOST='db' but got %q", api.Environment["DB_HOST"])
	}
}

func TestManifestResourceRefInInterpolation(t *testing.T) {
	conf, err := ParseManifest(`
cluster "test" {
  service "db" {
    image = "postgres:16"
  }

  service "api" {
    image   = "node:18"
    command = "node server.js --db-host ${service.db}"
    depends_on = [service.db]
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	api := conf.Clusters[0].Services[1]
	if api.Command != "node server.js --db-host db" {
		t.Errorf("expected command with interpolated service ref but got %q", api.Command)
	}
}

func TestManifestDynamicServiceBlock(t *testing.T) {
	conf, err := ParseManifest(`
variable "consul_servers" {
  default = ["consul-1", "consul-2", "consul-3"]
}

cluster "test" {
  dynamic "service" {
    for_each = var.consul_servers
    labels   = [each.value]
    content {
      image = "consul:1.15"
    }
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	services := conf.Clusters[0].Services
	if len(services) != 3 {
		t.Fatalf("expected 3 services but got %d", len(services))
	}
	for i, svc := range services {
		expected := fmt.Sprintf("consul-%d", i+1)
		if svc.Name != expected {
			t.Errorf("service[%d]: expected name %q but got %q", i, expected, svc.Name)
		}
		if svc.Image != "consul:1.15" {
			t.Errorf("service[%d]: expected image 'consul:1.15' but got %q", i, svc.Image)
		}
	}
}

func TestManifestDynamicWithRange(t *testing.T) {
	conf, err := ParseManifest(`
variable "leader_count" {
  default = 3
}

variable "image" {
  default = "helios-leader:latest"
}

cluster "test" {
  dynamic "service" {
    for_each = range(var.leader_count)
    labels   = [format("leader-%d", each.value + 1)]
    content {
      image    = var.image
      hostname = format("leader-%d.cluster.local", each.value + 1)
    }
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	services := conf.Clusters[0].Services
	if len(services) != 3 {
		t.Fatalf("expected 3 services but got %d", len(services))
	}
	for i, svc := range services {
		expectedName := fmt.Sprintf("leader-%d", i+1)
		expectedHost := fmt.Sprintf("leader-%d.cluster.local", i+1)
		if svc.Name != expectedName {
			t.Errorf("service[%d]: expected name %q but got %q", i, expectedName, svc.Name)
		}
		if svc.Image != "helios-leader:latest" {
			t.Errorf("service[%d]: expected image 'helios-leader:latest' but got %q", i, svc.Image)
		}
		if svc.Hostname != expectedHost {
			t.Errorf("service[%d]: expected hostname %q but got %q", i, expectedHost, svc.Hostname)
		}
	}
}

func TestManifestDynamicMixedWithStaticAndDeps(t *testing.T) {
	conf, err := ParseManifest(`
cluster "test" {
  service "db" {
    image = "postgres:16"
  }

  dynamic "service" {
    for_each = ["worker-1", "worker-2", "worker-3"]
    labels   = [each.value]
    content {
      image      = "worker:latest"
      depends_on = ["db"]
    }
  }

  service "api" {
    image      = "api:latest"
    depends_on = ["db"]
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	services := conf.Clusters[0].Services
	if len(services) != 5 {
		t.Fatalf("expected 5 services but got %d", len(services))
	}
	if services[0].Name != "db" {
		t.Errorf("expected first service 'db' but got %q", services[0].Name)
	}
	if services[1].Name != "worker-1" || services[2].Name != "worker-2" || services[3].Name != "worker-3" {
		t.Errorf("expected workers 1-3 but got %q, %q, %q", services[1].Name, services[2].Name, services[3].Name)
	}
	if services[4].Name != "api" {
		t.Errorf("expected last service 'api' but got %q", services[4].Name)
	}
	for _, svc := range services[1:4] {
		if len(svc.DependsOn) != 1 || svc.DependsOn[0] != "db" {
			t.Errorf("worker %s: expected depends_on [db] but got %v", svc.Name, svc.DependsOn)
		}
	}
}

func TestManifestDynamicWithResourceRefs(t *testing.T) {
	conf, err := ParseManifest(`
cluster "test" {
  dynamic "service" {
    for_each = ["consul-1", "consul-2", "consul-3"]
    labels   = [each.value]
    content {
      image = "consul:1.15"
    }
  }

  service "app" {
    image      = "myapp:latest"
    depends_on = [service.consul-1, service.consul-2, service.consul-3]
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	services := conf.Clusters[0].Services
	if len(services) != 4 {
		t.Fatalf("expected 4 services but got %d", len(services))
	}

	app := services[3]
	if len(app.DependsOn) != 3 {
		t.Fatalf("expected 3 dependencies but got %d", len(app.DependsOn))
	}
	for i, dep := range app.DependsOn {
		expected := fmt.Sprintf("consul-%d", i+1)
		if dep != expected {
			t.Errorf("depends_on[%d]: expected %q but got %q", i, expected, dep)
		}
	}
}

func TestManifestDynamicDependsOnSharedVariable(t *testing.T) {
	conf, err := ParseManifest(`
variable "consul_servers" {
  default = ["consul-1", "consul-2", "consul-3"]
}

cluster "test" {
  dynamic "service" {
    for_each = var.consul_servers
    labels   = [each.value]
    content {
      image = "consul:1.15"
    }
  }

  service "consul-client" {
    image      = "consul:1.15"
    depends_on = var.consul_servers
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	services := conf.Clusters[0].Services
	if len(services) != 4 {
		t.Fatalf("expected 4 services but got %d", len(services))
	}

	client := services[3]
	if client.Name != "consul-client" {
		t.Errorf("expected last service 'consul-client' but got %q", client.Name)
	}
	if len(client.DependsOn) != 3 {
		t.Fatalf("expected 3 dependencies but got %d", len(client.DependsOn))
	}
	for i, dep := range client.DependsOn {
		expected := fmt.Sprintf("consul-%d", i+1)
		if dep != expected {
			t.Errorf("depends_on[%d]: expected %q but got %q", i, expected, dep)
		}
	}
}

func TestManifestDynamicDependsOnWithLocalsForExpr(t *testing.T) {
	conf, err := ParseManifest(`
variable "consul_count" {
  default = 3
}

locals {
  consul_names = [for i in range(var.consul_count) : format("consul-%d", i + 1)]
}

cluster "test" {
  dynamic "service" {
    for_each = local.consul_names
    labels   = [each.value]
    content {
      image = "consul:1.15"
    }
  }

  service "consul-client" {
    image      = "consul:1.15"
    depends_on = local.consul_names
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	services := conf.Clusters[0].Services
	if len(services) != 4 {
		t.Fatalf("expected 4 services but got %d", len(services))
	}

	client := services[3]
	if client.Name != "consul-client" {
		t.Errorf("expected last service 'consul-client' but got %q", client.Name)
	}
	if len(client.DependsOn) != 3 {
		t.Fatalf("expected 3 dependencies but got %d", len(client.DependsOn))
	}
	for i, dep := range client.DependsOn {
		expected := fmt.Sprintf("consul-%d", i+1)
		if dep != expected {
			t.Errorf("depends_on[%d]: expected %q but got %q", i, expected, dep)
		}
	}
}

func TestManifestDynamicWithMapForEach(t *testing.T) {
	conf, err := ParseManifest(`
variable "services" {
  default = {
    web   = "nginx:latest"
    api   = "node:18"
    cache = "redis:7"
  }
}

cluster "test" {
  dynamic "service" {
    for_each = var.services
    labels   = [each.key]
    content {
      image = each.value
    }
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	services := conf.Clusters[0].Services
	if len(services) != 3 {
		t.Fatalf("expected 3 services but got %d", len(services))
	}
	imageMap := map[string]string{}
	for _, svc := range services {
		imageMap[svc.Name] = svc.Image
	}
	if imageMap["web"] != "nginx:latest" {
		t.Errorf("expected web image 'nginx:latest' but got %q", imageMap["web"])
	}
	if imageMap["api"] != "node:18" {
		t.Errorf("expected api image 'node:18' but got %q", imageMap["api"])
	}
	if imageMap["cache"] != "redis:7" {
		t.Errorf("expected cache image 'redis:7' but got %q", imageMap["cache"])
	}
}

func TestManifestConcatDependsOn(t *testing.T) {
	conf, err := ParseManifest(`
variable "consul_count" {
  default = 3
}

locals {
  consul_names = [for i in range(var.consul_count) : format("consul-%d", i + 1)]
}

cluster "test" {
  dynamic "service" {
    for_each = local.consul_names
    labels   = [each.value]
    content {
      image = "consul:1.15"
    }
  }

  service "leader" {
    image      = "leader:latest"
    depends_on = local.consul_names
  }

  service "worker" {
    image      = "worker:latest"
    depends_on = concat(local.consul_names, ["leader"])
  }
}
`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	services := conf.Clusters[0].Services
	if len(services) != 5 {
		t.Fatalf("expected 5 services but got %d", len(services))
	}

	leader := services[3]
	if len(leader.DependsOn) != 3 {
		t.Fatalf("expected leader to have 3 deps but got %d", len(leader.DependsOn))
	}

	worker := services[4]
	if worker.Name != "worker" {
		t.Errorf("expected worker but got %q", worker.Name)
	}
	if len(worker.DependsOn) != 4 {
		t.Fatalf("expected worker to have 4 deps but got %d: %v", len(worker.DependsOn), worker.DependsOn)
	}
	expectedDeps := []string{"consul-1", "consul-2", "consul-3", "leader"}
	for i, dep := range worker.DependsOn {
		if dep != expectedDeps[i] {
			t.Errorf("worker depends_on[%d]: expected %q but got %q", i, expectedDeps[i], dep)
		}
	}
}

func TestBuildArgsAndTarget(t *testing.T) {
	conf, err := ParseManifest(`cluster "test" {
		service "app" {
			build {
				context = "."
				dockerfile = "Dockerfile"
				target = "production"
				args = {
					"NODE_ENV" = "production"
					"VERSION" = "1.2.3"
				}
			}
		}
	}`)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	build := conf.Clusters[0].Services[0].Build
	if build.Target != "production" {
		t.Errorf("expected target 'production' but got '%s'", build.Target)
	}
	if len(build.Args) != 2 || build.Args["NODE_ENV"] != "production" || build.Args["VERSION"] != "1.2.3" {
		t.Errorf("expected build args {NODE_ENV:production, VERSION:1.2.3} but got %v", build.Args)
	}
}
