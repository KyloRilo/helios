FROM golang:1.24.1

WORKDIR /src
COPY ./ ./

RUN echo "=== Build context contents ===" && \
    find . && \
    echo "=============================="

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o worker ./cmd/worker/worker.go

ENTRYPOINT [ "/worker" ]