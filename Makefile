default:
	go mod tidy
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/service/api/proto/*.proto

run_edge:
	go run ./cmd/edge/*

run_leader:
	go run ./cmd/leader/*