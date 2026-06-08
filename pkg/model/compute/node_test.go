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
