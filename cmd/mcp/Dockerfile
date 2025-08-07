## Build
FROM golang:1.24 AS build

ARG VERSION='dev'

WORKDIR /app

COPY . /app

RUN apt-get update \
    && go mod download \
    && go tool sqlc generate \
    && go build -ldflags="-s -w -X github.com/Piszmog/pathwise/internal/version.Value=${VERSION}" -o mcp ./cmd/mcp

## Deploy
FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=build /app/mcp /mcp

EXPOSE 8080

CMD ["/mcp"]
