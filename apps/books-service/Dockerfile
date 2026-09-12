# syntax=docker/dockerfile:1.19.0

FROM golang:1.25.4-bookworm AS build

WORKDIR /app

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOFLAGS=-buildvcs=false

RUN apt-get update && apt-get install -y ca-certificates tzdata git && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/server ./cmd/server

FROM alpine:latest AS prod

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

RUN mkdir -p /.cache && chmod 777 /.cache

COPY --from=build /app/bin/server /app/server

EXPOSE 8080
USER 1000

ENTRYPOINT ["/app/server"]
