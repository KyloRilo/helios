FROM golang:1.22.2

WORKDIR ../

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./proto/*.go ./pkg/*.go ./models/*.go ./models/errors/*.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /core-service

CMD ["/core-service"]