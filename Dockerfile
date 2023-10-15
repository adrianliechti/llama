# syntax=docker/dockerfile:1

FROM golang:1-alpine AS build

WORKDIR /src

COPY go.* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOEXPERIMENT=loopvar go build -o server ./cmd/server


FROM alpine

RUN apk add --no-cache tini ca-certificates

WORKDIR /
COPY --from=build /src/server server

EXPOSE 50051

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/server"]