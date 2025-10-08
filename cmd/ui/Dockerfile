## Build
FROM golang:1.25 AS build

ARG VERSION='dev'

WORKDIR /app

COPY . /app

RUN apt-get update \
    && apt-get install -y build-essential \
    && go mod download \
    && go tool templ generate -path ./internal/ui/components \
    && go tool go-tw -i ./internal/ui/styles/input.css -o ./internal/ui/dist/assets/css/output@${VERSION}.css --minify \
    && go tool sqlc generate \
    && CGO_ENABLED=1 go build -ldflags="-s -w -X github.com/Piszmog/pathwise/internal/version.Value=${VERSION}" -o ui ./cmd/ui

## Deploy
FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=build /app/ui /ui

EXPOSE 8080

CMD ["/ui"]
