## Build
FROM golang:1.24 AS build

ARG VERSION='dev'

WORKDIR /app

COPY . /app

RUN apt-get update \
    && apt-get install -y build-essential \
    && go mod download \
    && go tool sqlc generate \
    && CGO_ENABLED=1 go build -ldflags="-s -w -X github.com/Piszmog/pathwise/internal/version.Value=${VERSION}" -o mcp ./cmd/mcp

## Deploy
FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=build /app/mcp /mcp

EXPOSE 8080

CMD ["/mcp"]
