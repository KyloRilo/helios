default:
	go mod tidy
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/*.proto

run:
	go run .

run_docker:
	go run . --service docker

run_cloud:
	go run . --service cloud