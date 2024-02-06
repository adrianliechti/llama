# syntax=docker/dockerfile:1

FROM golang:1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOEXPERIMENT=loopvar go build -o server

CMD ["/app/server"]