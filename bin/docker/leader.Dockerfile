FROM golang:1.24.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ pkg/
COPY cmd/leader/ cmd/leader/
COPY test/config /helios/config


RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o leader ./cmd/leader/leader.go

ENV DOCKER_HOST=unix:///var/run/docker.sock
VOLUME [ "/var/run/docker.sock", "/var/run/docker.sock" ]
EXPOSE 6330
ENTRYPOINT [ "./leader" ]