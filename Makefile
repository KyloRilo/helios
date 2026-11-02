

default: 
	go mod tidy
	go build -v ./...
	go test -v -cover ./pkg/...

gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/api/proto/*.proto

integration: # default
	go test ./test/integration/... -timeout 300s

cli:
	go build -o bin/helios/helios ./cmd/cli/

ui: export HELIOS_CONFIG_FILE = ./bin/helios/local.cluster.hcl
ui: default
	go run ./cmd/ui/

