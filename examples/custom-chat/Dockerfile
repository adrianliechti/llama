# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.* ./
RUN go mod download

COPY . .

WORKDIR examples/custom-chat
RUN CGO_ENABLED=0 GOOS=linux go build -o server

FROM alpine

RUN apk add --no-cache tini ca-certificates

WORKDIR /
COPY --from=build /src/examples/custom-chat/server .

EXPOSE 8080

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/server"]