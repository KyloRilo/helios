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

	path := filepath.Join(cwd, "helios.hcl")
	conf, err := model.ReadClusterConfigFile(path)
	if err != nil {
		t.Errorf("TestConfigRead() => %s", err)
	}
	fmt.Println("Helios Config: ", conf)
}

func TestClusterConfig(t *testing.T) {
	tests := []string{
		`cluster "test" {
			host = "test"
			port = "6300"
			service "test" {
				image = ""
				command = "echo hello"
				volumes = ["/data"]
				environment = ["ENV=test"]
				ports = ["8080:80"]

			}
		}`,
		`cluster "test" {
			host = "test"
			port = "6300"
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
			conf, err := model.ParseClusterConfig(test)
			if err != nil {
				t.Error(err)
			}
			t.Logf("Parsed Config: %+v", conf)
		})
	}
}
