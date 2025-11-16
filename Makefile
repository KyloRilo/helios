

default: 
	go mod tidy
	go build -v ./...
	go test ./test/...

gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/api/proto/*.proto

# cluster-up: default
# 	docker-compose up --build

launchpad: export HELIOS_CONFIG_FILE = ./build/helios/local.cluster.hcl
launchpad: default
	go run ./cmd/launchpad/launchpad.go