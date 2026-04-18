package unit

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/KyloRilo/helios/pkg/model"
)

func TestReadFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(cwd, "../../bin/helios/local.cluster.hcl")
	_, err = model.ReadManifestFile(path)
	if err != nil {
		t.Errorf("TestConfigRead() => %s", err)
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
			conf, err := model.ParseManifest(test)
			if err != nil {
				t.Error(err)
			}
			t.Logf("Parsed Config: %+v", conf)
		})
	}
}
