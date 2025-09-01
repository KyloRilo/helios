test:
	go test ./...

default: test
	go mod tidy
	go build -v ./...

cluster-up: default
	docker-compose up --build

gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/api/proto/*.proto