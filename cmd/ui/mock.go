package main

import (
	"math/rand"
	"time"
)

type ServiceStatus string

const (
	StatusRunning  ServiceStatus = "Running"
	StatusStopped  ServiceStatus = "Stopped"
	StatusCreating ServiceStatus = "Creating"
	StatusError    ServiceStatus = "Error"
)

type ServiceInfo struct {
	Name      string
	Image     string
	Status    ServiceStatus
	CPU       float64
	MemoryMB  float64
	NetworkMB float64
	Uptime    time.Duration
	Ports     string
	Replicas  int
}

type ClusterInfo struct {
	Name      string
	Services  []ServiceInfo
	CreatedAt time.Time
	IsHelios  bool
}

func heliosCluster() ClusterInfo {
	return ClusterInfo{
		Name:      "helios-local",
		IsHelios:  true,
		CreatedAt: time.Now().Add(-2 * time.Hour),
		Services: []ServiceInfo{
			{Name: "consul", Image: "hashicorp/consul:latest", Status: StatusRunning, CPU: 2.3, MemoryMB: 128, NetworkMB: 4.2, Uptime: 2 * time.Hour, Ports: "8500, 53", Replicas: 3},
			{Name: "consul-client", Image: "hashicorp/consul:latest", Status: StatusRunning, CPU: 2.3, MemoryMB: 128, NetworkMB: 4.2, Uptime: 2 * time.Hour, Ports: "8500, 53", Replicas: 3},
			{Name: "localstack", Image: "localstack/localstack:latest", Status: StatusRunning, CPU: 5.1, MemoryMB: 512, NetworkMB: 1.8, Uptime: 2 * time.Hour, Ports: "4566", Replicas: 1},
			{Name: "leader", Image: "helios/leader:latest", Status: StatusRunning, CPU: 8.4, MemoryMB: 256, NetworkMB: 12.5, Uptime: 2 * time.Hour, Ports: "6331", Replicas: 3},
			{Name: "worker", Image: "helios/worker:latest", Status: StatusRunning, CPU: 3.7, MemoryMB: 192, NetworkMB: 8.1, Uptime: 2 * time.Hour, Ports: "-", Replicas: 1},
			{Name: "api", Image: "helios/api:latest", Status: StatusRunning, CPU: 4.2, MemoryMB: 180, NetworkMB: 22.3, Uptime: 2 * time.Hour, Ports: "8080", Replicas: 1},
		},
	}
}

func sampleUserClusters() []ClusterInfo {
	return []ClusterInfo{
		{
			Name:      "webapp-prod",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			Services: []ServiceInfo{
				{Name: "nginx", Image: "nginx:1.25", Status: StatusRunning, CPU: 1.2, MemoryMB: 64, NetworkMB: 45.2, Uptime: 48 * time.Hour, Ports: "80, 443", Replicas: 3},
				{Name: "app-server", Image: "myapp/server:v2.1", Status: StatusRunning, CPU: 22.5, MemoryMB: 1024, NetworkMB: 18.7, Uptime: 48 * time.Hour, Ports: "3000", Replicas: 4},
				{Name: "postgres", Image: "postgres:16", Status: StatusRunning, CPU: 15.3, MemoryMB: 2048, NetworkMB: 8.4, Uptime: 48 * time.Hour, Ports: "5432", Replicas: 1},
				{Name: "redis", Image: "redis:7-alpine", Status: StatusRunning, CPU: 3.8, MemoryMB: 256, NetworkMB: 12.1, Uptime: 48 * time.Hour, Ports: "6379", Replicas: 2},
				{Name: "worker", Image: "myapp/worker:v2.1", Status: StatusRunning, CPU: 18.1, MemoryMB: 512, NetworkMB: 5.3, Uptime: 48 * time.Hour, Ports: "-", Replicas: 6},
			},
		},
		{
			Name:      "ml-pipeline",
			CreatedAt: time.Now().Add(-12 * time.Hour),
			Services: []ServiceInfo{
				{Name: "jupyter", Image: "jupyter/scipy:latest", Status: StatusRunning, CPU: 45.2, MemoryMB: 4096, NetworkMB: 2.1, Uptime: 12 * time.Hour, Ports: "8888", Replicas: 1},
				{Name: "mlflow", Image: "mlflow/mlflow:2.8", Status: StatusRunning, CPU: 8.9, MemoryMB: 512, NetworkMB: 3.4, Uptime: 12 * time.Hour, Ports: "5000", Replicas: 1},
				{Name: "minio", Image: "minio/minio:latest", Status: StatusRunning, CPU: 6.1, MemoryMB: 1024, NetworkMB: 28.5, Uptime: 12 * time.Hour, Ports: "9000, 9001", Replicas: 1},
				{Name: "trainer", Image: "ml/trainer:v1.4", Status: StatusStopped, CPU: 0, MemoryMB: 0, NetworkMB: 0, Uptime: 0, Ports: "-", Replicas: 2},
			},
		},
	}
}

func randomizeMetrics(svc *ServiceInfo) {
	if svc.Status != StatusRunning {
		return
	}
	svc.CPU += (rand.Float64() - 0.5) * 2
	if svc.CPU < 0.1 {
		svc.CPU = 0.1
	}
	svc.MemoryMB += (rand.Float64() - 0.5) * 20
	if svc.MemoryMB < 10 {
		svc.MemoryMB = 10
	}
	svc.NetworkMB += (rand.Float64() - 0.5) * 3
	if svc.NetworkMB < 0.1 {
		svc.NetworkMB = 0.1
	}
}
