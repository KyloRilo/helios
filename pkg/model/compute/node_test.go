package compute

import (
	"reflect"
	"sort"
	"testing"
)

func TestNode(t *testing.T) {
	n := NewNode(
		WithId("test-id"),
		WithName("test-node"),
		WithImage("alpine:latest"),
		WithCmd("echo Hello World"),
		WithPorts(Ports{"8080": "80"}),
		WithEnv(map[string]string{"ENV_VAR": "value"}),
		WithVolumes(map[string]string{"/host/path": "/container/path"}),
		WithTags([]string{"tag1", "tag2"}),
	)

	if n.Name != "test-node" {
		t.Errorf("Expected name 'test-node' but got '%s'", n.Name)
	}

	if n.Image != "alpine:latest" {
		t.Errorf("Expected image 'alpine:latest' but got '%s'", n.Image)
	}

	if n.Cmd != "echo Hello World" {
		t.Errorf("Expected cmd 'echo Hello World' but got '%s'", n.Cmd)
	}

	expectedPorts := Ports{"8080": "80"}
	if !reflect.DeepEqual(n.Ports, expectedPorts) {
		t.Errorf("Expected ports %v but got %v", expectedPorts, n.Ports)
	}

	expectedEnv := map[string]string{"ENV_VAR": "value"}
	if !reflect.DeepEqual(n.Env, expectedEnv) {
		t.Errorf("Expected env %v but got %v", expectedEnv, n.Env)
	}

	expectedVolumes := map[string]string{"/host/path": "/container/path"}
	if !reflect.DeepEqual(n.Volumes, expectedVolumes) {
		t.Errorf("Expected volumes %v but got %v", expectedVolumes, n.Volumes)
	}

	expectedTags := []string{"tag1", "tag2"}
	if !reflect.DeepEqual(n.Tags, expectedTags) {
		t.Errorf("Expected tags %v but got %v", expectedTags, n.Tags)
	}

	if n.Status != Ready {
		t.Errorf("Expected status 'READY' but got '%s'", n.Status)
	}
}

func TestNodeAllOptions(t *testing.T) {
	hc := &Healthcheck{
		Test:        "curl -f http://localhost/",
		Interval:    "30s",
		Timeout:     "10s",
		Retries:     3,
		StartPeriod: "5s",
	}
	res := &Resources{
		CPULimit:          "0.5",
		MemoryLimit:       "512m",
		CPUReservation:    "0.25",
		MemoryReservation: "256m",
	}
	ctx := &Context{Path: "./app", File: "Dockerfile", Args: map[string]string{"ENV": "prod"}, Target: "production"}

	n := NewNode(
		WithName("full-node"),
		WithContext(ctx),
		WithDependsOn([]string{"consul", "core"}),
		WithHostname("myhost"),
		WithRestart("always"),
		WithReplicas(3),
		WithNetworks([]string{"frontend", "backend"}),
		WithHealthcheck(hc),
		WithResources(res),
		WithEntrypoint("/entrypoint.sh"),
		WithWorkingDir("/app"),
		WithUser("node:node"),
		WithLabels(map[string]string{"tier": "web"}),
		WithEnvFile(".env"),
		WithExpose([]string{"3000", "9090"}),
		WithStatus(Created),
	)

	if n.Context != ctx {
		t.Errorf("expected context to be set")
	}
	if n.Context.Args["ENV"] != "prod" {
		t.Errorf("expected context args ENV=prod")
	}
	if n.Context.Target != "production" {
		t.Errorf("expected context target 'production'")
	}
	if n.Hostname != "myhost" {
		t.Errorf("expected hostname 'myhost' but got '%s'", n.Hostname)
	}
	if n.Restart != "always" {
		t.Errorf("expected restart 'always' but got '%s'", n.Restart)
	}
	if n.Replicas != 3 {
		t.Errorf("expected replicas 3 but got %d", n.Replicas)
	}
	if !reflect.DeepEqual(n.Networks, []string{"frontend", "backend"}) {
		t.Errorf("expected networks [frontend, backend] but got %v", n.Networks)
	}
	if n.Healthcheck != hc {
		t.Errorf("expected healthcheck to be set")
	}
	if n.Resources != res {
		t.Errorf("expected resources to be set")
	}
	if n.Entrypoint != "/entrypoint.sh" {
		t.Errorf("expected entrypoint '/entrypoint.sh' but got '%s'", n.Entrypoint)
	}
	if n.WorkingDir != "/app" {
		t.Errorf("expected working_dir '/app' but got '%s'", n.WorkingDir)
	}
	if n.User != "node:node" {
		t.Errorf("expected user 'node:node' but got '%s'", n.User)
	}
	if n.Labels["tier"] != "web" {
		t.Errorf("expected label tier=web but got %v", n.Labels)
	}
	if n.EnvFile != ".env" {
		t.Errorf("expected env_file '.env' but got '%s'", n.EnvFile)
	}
	if !reflect.DeepEqual(n.Expose, []string{"3000", "9090"}) {
		t.Errorf("expected expose [3000, 9090] but got %v", n.Expose)
	}
	if !reflect.DeepEqual(n.DependsOn, []string{"consul", "core"}) {
		t.Errorf("expected depends_on [consul, core] but got %v", n.DependsOn)
	}
	// WithStatus is applied before NewNode sets Ready, so final status is Ready
	if n.Status != Ready {
		t.Errorf("expected status Ready but got '%s'", n.Status)
	}
}

func TestPortsToStringArray(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		p := Ports{}
		result := p.ToStringArray()
		if len(result) != 0 {
			t.Errorf("expected empty array but got %v", result)
		}
	})

	t.Run("single", func(t *testing.T) {
		p := Ports{"8080": "80"}
		result := p.ToStringArray()
		if len(result) != 1 || result[0] != "8080:80/tcp" {
			t.Errorf("expected ['8080:80/tcp'] but got %v", result)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		p := Ports{"8080": "80", "443": "443"}
		result := p.ToStringArray()
		if len(result) != 2 {
			t.Fatalf("expected 2 entries but got %d", len(result))
		}
		sort.Strings(result)
		if result[0] != "443:443/tcp" || result[1] != "8080:80/tcp" {
			t.Errorf("unexpected port strings: %v", result)
		}
	})

	t.Run("nil", func(t *testing.T) {
		var p Ports
		result := p.ToStringArray()
		if len(result) != 0 {
			t.Errorf("expected empty array for nil ports but got %v", result)
		}
	})
}

func TestNewNodeDefaults(t *testing.T) {
	n := NewNode()
	if n.Status != Ready {
		t.Errorf("expected default status Ready but got '%s'", n.Status)
	}
	if n.Name != "" {
		t.Errorf("expected empty name but got '%s'", n.Name)
	}
}
