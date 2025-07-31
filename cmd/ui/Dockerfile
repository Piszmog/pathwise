## Build
FROM golang:1.24 AS build

ARG VERSION='dev'

WORKDIR /app

COPY . /app

RUN apt-get update \
    && go mod download \
    && go tool templ generate -path ./ui/components \
    && go tool go-tw -i ./ui/styles/input.css -o ./ui/dist/assets/css/output@${VERSION}.css --minify \
    && go tool sqlc generate \
    && go build -ldflags="-s -w -X github.com/Piszmog/pathwise/ui/version.Value=${VERSION}" -o ui ./cmd/ui

## Deploy
FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=build /app/ui /ui

EXPOSE 8080

CMD ["/ui"]
