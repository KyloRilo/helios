package compute

import (
	"reflect"
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

	if n.GetName() != "test-node" {
		t.Errorf("Expected name 'test-node' but got '%s'", n.GetName())
	}

	if n.GetImage() != "alpine:latest" {
		t.Errorf("Expected image 'alpine:latest' but got '%s'", n.GetImage())
	}

	if n.GetCmd() != "echo Hello World" {
		t.Errorf("Expected cmd 'echo Hello World' but got '%s'", n.GetCmd())
	}

	expectedPorts := Ports{"8080": "80"}
	if !reflect.DeepEqual(n.GetPorts(), expectedPorts) {
		t.Errorf("Expected ports %v but got %v", expectedPorts, n.GetPorts())
	}

	expectedEnv := map[string]string{"ENV_VAR": "value"}
	if !reflect.DeepEqual(n.GetEnv(), expectedEnv) {
		t.Errorf("Expected env %v but got %v", expectedEnv, n.GetEnv())
	}

	expectedVolumes := map[string]string{"/host/path": "/container/path"}
	if !reflect.DeepEqual(n.GetVolumes(), expectedVolumes) {
		t.Errorf("Expected volumes %v but got %v", expectedVolumes, n.GetVolumes())
	}

	expectedTags := []string{"tag1", "tag2"}
	if !reflect.DeepEqual(n.tags, expectedTags) {
		t.Errorf("Expected tags %v but got %v", expectedTags, n.tags)
	}

	if n.GetStatus() != Ready {
		t.Errorf("Expected status 'READY' but got '%s'", n.GetStatus())
	}
}

func TestNodeSetters(t *testing.T) {
	n := NewNode()
	n.SetName("test-node")
	n.SetImage("alpine:latest")
	n.SetCmd("echo Hello World")
	n.SetPorts(Ports{"8080": "80"})
	n.SetEnv(map[string]string{"ENV_VAR": "value"})
	n.SetVolumes(map[string]string{"/host/path": "/container/path"})
	n.SetContext(&Context{Path: "/context", File: "Dockerfile"})

	if n.GetName() != "test-node" {
		t.Errorf("Expected name 'test-node' but got '%s'", n.GetName())
	}

	if n.GetImage() != "alpine:latest" {
		t.Errorf("Expected image 'alpine:latest' but got '%s'", n.GetImage())
	}

	if n.GetCmd() != "echo Hello World" {
		t.Errorf("Expected cmd 'echo Hello World' but got '%s'", n.GetCmd())
	}

	expectedPorts := Ports{"8080": "80"}
	if !reflect.DeepEqual(n.GetPorts(), expectedPorts) {
		t.Errorf("Expected ports %v but got %v", expectedPorts, n.GetPorts())
	}

	expectedEnv := map[string]string{"ENV_VAR": "value"}
	if !reflect.DeepEqual(n.GetEnv(), expectedEnv) {
		t.Errorf("Expected env %v but got %v", expectedEnv, n.GetEnv())
	}

	expectedVolumes := map[string]string{"/host/path": "/container/path"}
	if !reflect.DeepEqual(n.GetVolumes(), expectedVolumes) {
		t.Errorf("Expected volumes %v but got %v", expectedVolumes, n.GetVolumes())
	}

	expectedContext := &Context{Path: "/context", File: "Dockerfile"}
	if !reflect.DeepEqual(n.GetContext(), expectedContext) {
		t.Errorf("Expected context %v but got %v", expectedContext, n.GetContext())
	}
}
