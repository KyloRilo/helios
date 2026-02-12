FROM golang:1.24.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ pkg/
COPY cmd/api/ cmd/api/
COPY test/config /helios/config


RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o api ./cmd/api/api.go

EXPOSE 8080
ENTRYPOINT [ "./api" ]