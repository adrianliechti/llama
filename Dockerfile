# syntax=docker/dockerfile:1

FROM golang:1-alpine AS build

WORKDIR /src

COPY go.* ./
RUN go mod download

COPY . .

WORKDIR /src/cmd/server
RUN CGO_ENABLED=0 go build -o server

WORKDIR /src/cmd/client
RUN CGO_ENABLED=0 go build -o client

WORKDIR /src/cmd/ingest
RUN CGO_ENABLED=0 go build -o ingest

FROM alpine

RUN apk add --no-cache tini ca-certificates mailcap

WORKDIR /
COPY --from=build /src/cmd/server/server .
COPY --from=build /src/cmd/client/client .
COPY --from=build /src/cmd/ingest/ingest .

EXPOSE 8080

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/server"]