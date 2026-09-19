package eval

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

func TestBuildEvalContextVariables(t *testing.T) {
	src := []byte(`
variable "env" {
  default = "production"
}

variable "region" {
  default = "us-east-1"
}

name = "${var.env}-${var.region}"
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Name string `hcl:"name"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if result.Name != "production-us-east-1" {
		t.Errorf("expected 'production-us-east-1' but got %q", result.Name)
	}
}

func TestBuildEvalContextLocals(t *testing.T) {
	src := []byte(`
variable "project" {
  default = "helios"
}

variable "env" {
  default = "dev"
}

locals {
  prefix = "${var.project}-${var.env}"
  tag    = "${local.prefix}-v1"
}

name = local.tag
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Name string `hcl:"name"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if result.Name != "helios-dev-v1" {
		t.Errorf("expected 'helios-dev-v1' but got %q", result.Name)
	}
}

func TestBuildEvalContextFunctions(t *testing.T) {
	src := []byte(`
variable "name" {
  default = "helios"
}

upper_name = upper(var.name)
formatted  = format("Hello, %s!", var.name)
joined     = join("-", ["a", "b", "c"])
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		UpperName string `hcl:"upper_name"`
		Formatted string `hcl:"formatted"`
		Joined    string `hcl:"joined"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	if result.UpperName != "HELIOS" {
		t.Errorf("expected 'HELIOS' but got %q", result.UpperName)
	}
	if result.Formatted != "Hello, helios!" {
		t.Errorf("expected 'Hello, helios!' but got %q", result.Formatted)
	}
	if result.Joined != "a-b-c" {
		t.Errorf("expected 'a-b-c' but got %q", result.Joined)
	}
}

func TestBuildEvalContextLookup(t *testing.T) {
	src := []byte(`
variable "images" {
  default = {
    web = "nginx:latest"
    api = "node:18"
  }
}

web_image = lookup(var.images, "web", "default:latest")
missing   = lookup(var.images, "db", "postgres:16")
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		WebImage string `hcl:"web_image"`
		Missing  string `hcl:"missing"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	if result.WebImage != "nginx:latest" {
		t.Errorf("expected 'nginx:latest' but got %q", result.WebImage)
	}
	if result.Missing != "postgres:16" {
		t.Errorf("expected 'postgres:16' but got %q", result.Missing)
	}
}

func TestVariableWithNoDefault(t *testing.T) {
	src := []byte(`
variable "required_var" {}

name = var.required_var
`)
	_, _, err := BuildEvalContext(src, "test.hcl")
	if err == nil {
		t.Fatal("expected error for variable with no default")
	}
}

func TestParseError(t *testing.T) {
	src := []byte(`{{{invalid`)
	_, _, err := BuildEvalContext(src, "test.hcl")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestNoVariablesOrLocals(t *testing.T) {
	src := []byte(`
name = "plain-value"
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Name string `hcl:"name"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if result.Name != "plain-value" {
		t.Errorf("expected 'plain-value' but got %q", result.Name)
	}
}

func TestLocalReferencesAnotherLocal(t *testing.T) {
	src := []byte(`
variable "base" {
  default = "app"
}

locals {
  name    = "${var.base}-service"
  fullname = "${local.name}-prod"
}

result = local.fullname
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Result string `hcl:"result"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if result.Result != "app-service-prod" {
		t.Errorf("expected 'app-service-prod' but got %q", result.Result)
	}
}

func TestFunctionChaining(t *testing.T) {
	src := []byte(`
variable "name" {
  default = "  Helios  "
}

result = upper(trimspace(var.name))
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Result string `hcl:"result"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if result.Result != "HELIOS" {
		t.Errorf("expected 'HELIOS' but got %q", result.Result)
	}
}

func TestBuildFileEvalContext(t *testing.T) {
	src := []byte(`
variable "env" {
  default = "staging"
}

locals {
  prefix = "helios-${var.env}"
}

name = local.prefix
`)
	ctx, err := BuildFileEvalContext("test.hcl", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if ctx.Variables["var"].IsNull() {
		t.Error("expected var namespace to be populated")
	}
	if ctx.Variables["local"].IsNull() {
		t.Error("expected local namespace to be populated")
	}
}

func TestFunctions(t *testing.T) {
	fns := Functions()
	expected := []string{
		"format", "join", "upper", "lower", "replace",
		"split", "trimspace", "substr", "strlen", "concat",
		"coalesce", "lookup",
	}
	for _, name := range expected {
		if _, ok := fns[name]; !ok {
			t.Errorf("expected function %q to be registered", name)
		}
	}
}

func TestVariableDescription(t *testing.T) {
	src := []byte(`
variable "env" {
  description = "The deployment environment"
  default     = "dev"
}

name = var.env
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Name string `hcl:"name"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if result.Name != "dev" {
		t.Errorf("expected 'dev' but got %q", result.Name)
	}
}

func TestNumericVariable(t *testing.T) {
	src := []byte(`
variable "replicas" {
  default = 3
}

count = var.replicas
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Count int `hcl:"count"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if result.Count != 3 {
		t.Errorf("expected 3 but got %d", result.Count)
	}
}

func TestBooleanVariable(t *testing.T) {
	src := []byte(`
variable "enabled" {
  default = true
}

active = var.enabled
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Active bool `hcl:"active"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if !result.Active {
		t.Error("expected true but got false")
	}
}

func TestListVariable(t *testing.T) {
	src := []byte(`
variable "networks" {
  default = ["frontend", "backend"]
}

nets = var.networks
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Nets []string `hcl:"nets"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if len(result.Nets) != 2 || result.Nets[0] != "frontend" || result.Nets[1] != "backend" {
		t.Errorf("expected [frontend, backend] but got %v", result.Nets)
	}
}

func TestServiceResourceRef(t *testing.T) {
	type testService struct {
		Name      string   `hcl:"name,label"`
		Image     string   `hcl:"image"`
		DependsOn []string `hcl:"depends_on,optional"`
	}
	type testCluster struct {
		Name     string        `hcl:"name,label"`
		Services []testService `hcl:"service,block"`
	}
	type testManifest struct {
		Clusters []testCluster `hcl:"cluster,block"`
	}

	src := []byte(`
cluster "test" {
  service "db" {
    image = "postgres:16"
  }

  service "api" {
    image      = "node:18"
    depends_on = [service.db]
  }

  service "web" {
    image      = "nginx:latest"
    depends_on = [service.db, service.api]
  }
}
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result testManifest
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	api := result.Clusters[0].Services[1]
	if len(api.DependsOn) != 1 || api.DependsOn[0] != "db" {
		t.Errorf("expected api depends_on [db] but got %v", api.DependsOn)
	}

	web := result.Clusters[0].Services[2]
	if len(web.DependsOn) != 2 || web.DependsOn[0] != "db" || web.DependsOn[1] != "api" {
		t.Errorf("expected web depends_on [db, api] but got %v", web.DependsOn)
	}
}

func TestServiceResourceRefWithHyphen(t *testing.T) {
	type testService struct {
		Name      string   `hcl:"name,label"`
		Image     string   `hcl:"image"`
		DependsOn []string `hcl:"depends_on,optional"`
	}
	type testCluster struct {
		Name     string        `hcl:"name,label"`
		Services []testService `hcl:"service,block"`
	}
	type testManifest struct {
		Clusters []testCluster `hcl:"cluster,block"`
	}

	src := []byte(`
cluster "test" {
  service "consul-1" {
    image = "consul:1.15"
  }

  service "consul-2" {
    image      = "consul:1.15"
    depends_on = [service.consul-1]
  }

  service "app-server" {
    image      = "myapp:latest"
    depends_on = [service.consul-1, service.consul-2]
  }
}
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result testManifest
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	consul2 := result.Clusters[0].Services[1]
	if len(consul2.DependsOn) != 1 || consul2.DependsOn[0] != "consul-1" {
		t.Errorf("expected consul-2 depends_on [consul-1] but got %v", consul2.DependsOn)
	}

	app := result.Clusters[0].Services[2]
	if len(app.DependsOn) != 2 || app.DependsOn[0] != "consul-1" || app.DependsOn[1] != "consul-2" {
		t.Errorf("expected app depends_on [consul-1, consul-2] but got %v", app.DependsOn)
	}
}

func TestResourceRefInInterpolation(t *testing.T) {
	type testService struct {
		Name    string `hcl:"name,label"`
		Image   string `hcl:"image"`
		Command string `hcl:"command,optional"`
	}
	type testCluster struct {
		Name     string        `hcl:"name,label"`
		Services []testService `hcl:"service,block"`
	}
	type testManifest struct {
		Clusters []testCluster `hcl:"cluster,block"`
	}

	src := []byte(`
cluster "test" {
  service "db" {
    image = "postgres:16"
  }

  service "api" {
    image   = "node:18"
    command = "connect --host ${service.db}"
  }
}
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result testManifest
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	api := result.Clusters[0].Services[1]
	if api.Command != "connect --host db" {
		t.Errorf("expected command 'connect --host db' but got %q", api.Command)
	}
}

func TestResourceRefWithVariablesAndLocals(t *testing.T) {
	type testService struct {
		Name      string   `hcl:"name,label"`
		Image     string   `hcl:"image"`
		DependsOn []string `hcl:"depends_on,optional"`
	}
	type testCluster struct {
		Name     string        `hcl:"name,label"`
		Services []testService `hcl:"service,block"`
	}
	type testManifest struct {
		Clusters []testCluster `hcl:"cluster,block"`
	}

	src := []byte(`
variable "tag" {
  default = "latest"
}

locals {
  api_image = "myapp:${var.tag}"
}

cluster "test" {
  service "db" {
    image = "postgres:16"
  }

  service "api" {
    image      = local.api_image
    depends_on = [service.db]
  }
}
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result testManifest
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	api := result.Clusters[0].Services[1]
	if api.Image != "myapp:latest" {
		t.Errorf("expected image 'myapp:latest' but got %q", api.Image)
	}
	if len(api.DependsOn) != 1 || api.DependsOn[0] != "db" {
		t.Errorf("expected depends_on [db] but got %v", api.DependsOn)
	}
}

func TestClusterResourceRef(t *testing.T) {
	src := []byte(`
cluster "prod" {
  service "web" {
    image = "nginx:latest"
  }
}

name = cluster.prod
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Name     string `hcl:"name"`
		Remain   hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if result.Name != "prod" {
		t.Errorf("expected 'prod' but got %q", result.Name)
	}
}

func TestDynamicBlockWithList(t *testing.T) {
	type testService struct {
		Name  string `hcl:"name,label"`
		Image string `hcl:"image"`
	}
	type testCluster struct {
		Name     string        `hcl:"name,label"`
		Services []testService `hcl:"service,block"`
	}
	type testManifest struct {
		Clusters []testCluster `hcl:"cluster,block"`
	}

	src := []byte(`
cluster "test" {
  dynamic "service" {
    for_each = ["consul-1", "consul-2", "consul-3"]
    labels   = [each.value]
    content {
      image = "consul:1.15"
    }
  }
}
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result testManifest
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	services := result.Clusters[0].Services
	if len(services) != 3 {
		t.Fatalf("expected 3 services but got %d", len(services))
	}
	expected := []string{"consul-1", "consul-2", "consul-3"}
	for i, name := range expected {
		if services[i].Name != name {
			t.Errorf("service[%d]: expected name %q but got %q", i, name, services[i].Name)
		}
		if services[i].Image != "consul:1.15" {
			t.Errorf("service[%d]: expected image 'consul:1.15' but got %q", i, services[i].Image)
		}
	}
}

func TestDynamicBlockWithRange(t *testing.T) {
	type testService struct {
		Name  string `hcl:"name,label"`
		Image string `hcl:"image"`
	}
	type testCluster struct {
		Name     string        `hcl:"name,label"`
		Services []testService `hcl:"service,block"`
	}
	type testManifest struct {
		Clusters []testCluster `hcl:"cluster,block"`
	}

	src := []byte(`
variable "count" {
  default = 3
}

cluster "test" {
  dynamic "service" {
    for_each = range(var.count)
    labels   = [format("worker-%d", each.value)]
    content {
      image = "worker:latest"
    }
  }
}
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result testManifest
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	services := result.Clusters[0].Services
	if len(services) != 3 {
		t.Fatalf("expected 3 services but got %d", len(services))
	}
	for i, svc := range services {
		expected := fmt.Sprintf("worker-%d", i)
		if svc.Name != expected {
			t.Errorf("service[%d]: expected name %q but got %q", i, expected, svc.Name)
		}
	}
}

func TestDynamicBlockWithEachInContent(t *testing.T) {
	type testService struct {
		Name     string `hcl:"name,label"`
		Image    string `hcl:"image"`
		Hostname string `hcl:"hostname,optional"`
	}
	type testCluster struct {
		Name     string        `hcl:"name,label"`
		Services []testService `hcl:"service,block"`
	}
	type testManifest struct {
		Clusters []testCluster `hcl:"cluster,block"`
	}

	src := []byte(`
variable "image_tag" {
  default = "1.15"
}

cluster "test" {
  dynamic "service" {
    for_each = ["consul-1", "consul-2", "consul-3"]
    labels   = [each.value]
    content {
      image    = "consul:${var.image_tag}"
      hostname = "${each.value}.cluster.local"
    }
  }
}
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result testManifest
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	services := result.Clusters[0].Services
	if len(services) != 3 {
		t.Fatalf("expected 3 services but got %d", len(services))
	}
	for i, svc := range services {
		expectedName := fmt.Sprintf("consul-%d", i+1)
		expectedHost := fmt.Sprintf("consul-%d.cluster.local", i+1)
		if svc.Name != expectedName {
			t.Errorf("service[%d]: expected name %q but got %q", i, expectedName, svc.Name)
		}
		if svc.Image != "consul:1.15" {
			t.Errorf("service[%d]: expected image 'consul:1.15' but got %q", i, svc.Image)
		}
		if svc.Hostname != expectedHost {
			t.Errorf("service[%d]: expected hostname %q but got %q", i, expectedHost, svc.Hostname)
		}
	}
}

func TestDynamicBlockWithMap(t *testing.T) {
	type testService struct {
		Name  string `hcl:"name,label"`
		Image string `hcl:"image"`
	}
	type testCluster struct {
		Name     string        `hcl:"name,label"`
		Services []testService `hcl:"service,block"`
	}
	type testManifest struct {
		Clusters []testCluster `hcl:"cluster,block"`
	}

	src := []byte(`
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
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result testManifest
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	services := result.Clusters[0].Services
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

func TestDynamicBlockMixedWithStatic(t *testing.T) {
	type testService struct {
		Name      string   `hcl:"name,label"`
		Image     string   `hcl:"image"`
		DependsOn []string `hcl:"depends_on,optional"`
	}
	type testCluster struct {
		Name     string        `hcl:"name,label"`
		Services []testService `hcl:"service,block"`
	}
	type testManifest struct {
		Clusters []testCluster `hcl:"cluster,block"`
	}

	src := []byte(`
cluster "test" {
  service "db" {
    image = "postgres:16"
  }

  dynamic "service" {
    for_each = ["worker-1", "worker-2"]
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
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result testManifest
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	services := result.Clusters[0].Services
	if len(services) != 4 {
		t.Fatalf("expected 4 services but got %d", len(services))
	}
	if services[0].Name != "db" {
		t.Errorf("expected first service 'db' but got %q", services[0].Name)
	}
	if services[1].Name != "worker-1" {
		t.Errorf("expected second service 'worker-1' but got %q", services[1].Name)
	}
	if services[2].Name != "worker-2" {
		t.Errorf("expected third service 'worker-2' but got %q", services[2].Name)
	}
	if services[3].Name != "api" {
		t.Errorf("expected fourth service 'api' but got %q", services[3].Name)
	}
}

func TestDynamicBlockMissingForEach(t *testing.T) {
	src := []byte(`
cluster "test" {
  dynamic "service" {
    content {
      image = "nginx:latest"
    }
  }
}
`)
	_, _, err := BuildEvalContext(src, "test.hcl")
	if err == nil {
		t.Fatal("expected error for dynamic block missing for_each")
	}
}

func TestDynamicBlockMissingContent(t *testing.T) {
	src := []byte(`
cluster "test" {
  dynamic "service" {
    for_each = ["a", "b"]
  }
}
`)
	_, _, err := BuildEvalContext(src, "test.hcl")
	if err == nil {
		t.Fatal("expected error for dynamic block missing content")
	}
}

func TestDynamicBlockResourceRefsFromExpanded(t *testing.T) {
	type testService struct {
		Name      string   `hcl:"name,label"`
		Image     string   `hcl:"image"`
		DependsOn []string `hcl:"depends_on,optional"`
	}
	type testCluster struct {
		Name     string        `hcl:"name,label"`
		Services []testService `hcl:"service,block"`
	}
	type testManifest struct {
		Clusters []testCluster `hcl:"cluster,block"`
	}

	src := []byte(`
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
    depends_on = [service.consul-1, service.consul-2]
  }
}
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result testManifest
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}

	services := result.Clusters[0].Services
	if len(services) != 4 {
		t.Fatalf("expected 4 services but got %d", len(services))
	}

	app := services[3]
	if len(app.DependsOn) != 2 || app.DependsOn[0] != "consul-1" || app.DependsOn[1] != "consul-2" {
		t.Errorf("expected app depends_on [consul-1, consul-2] but got %v", app.DependsOn)
	}
}

func TestRangeFunction(t *testing.T) {
	src := []byte(`
items = range(5)
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Items []int `hcl:"items"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if len(result.Items) != 5 {
		t.Fatalf("expected 5 items but got %d", len(result.Items))
	}
	for i, v := range result.Items {
		if v != i {
			t.Errorf("items[%d]: expected %d but got %d", i, i, v)
		}
	}
}

func TestRangeFunctionZero(t *testing.T) {
	src := []byte(`
items = range(0)
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Items []int `hcl:"items"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if len(result.Items) != 0 {
		t.Errorf("expected 0 items but got %d", len(result.Items))
	}
}

func TestMapVariable(t *testing.T) {
	src := []byte(`
variable "labels" {
  default = {
    app  = "helios"
    tier = "backend"
  }
}

tags = var.labels
`)
	ctx, body, err := BuildEvalContext(src, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Tags map[string]string `hcl:"tags"`
	}
	diags := gohcl.DecodeBody(body, ctx, &result)
	if diags.HasErrors() {
		t.Fatalf("decode error: %s", diags.Error())
	}
	if result.Tags["app"] != "helios" || result.Tags["tier"] != "backend" {
		t.Errorf("unexpected tags: %v", result.Tags)
	}
}
