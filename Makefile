default:
	go mod tidy
	go build -v ./...

test:
	go test ./...

gen_proto: default
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/service/api/proto/*.proto

run_edge: default
	go run ./cmd/edge/*

run_leader: default
	go run ./cmd/leader/*