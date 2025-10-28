## Build
FROM golang:1.25 AS build

ARG VERSION='dev'

WORKDIR /app

COPY . /app

RUN apt-get update \
    && apt-get install -y build-essential \
    && go mod download \
    && go tool sqlc generate \
    && CGO_ENABLED=1 go build -ldflags="-s -w -X github.com/Piszmog/pathwise/internal/version.Value=${VERSION}" -o jobs ./cmd/jobs

## Deploy
FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=build /app/jobs /jobs

EXPOSE 8081

CMD ["/jobs"]
