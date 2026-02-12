

default: 
	go mod tidy
	go build -v ./...
	go test ./test/unit/... && go test ./test/config/...

gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/api/proto/*.proto

launchpad: export HELIOS_CONFIG_FILE = ./bin/helios/local.cluster.hcl
launchpad: default
	go run ./cmd/launchpad/launchpad.go