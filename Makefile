default:
	go mod tidy
	go build -v ./...

test:
	go test ./...

gen_proto: default
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/api/proto/*.proto

cluster_up:
	docker-compose up --build
