## Build
FROM golang:1.24 AS build

ARG VERSION='dev'

RUN apt-get update

WORKDIR /app

COPY ./ /app

RUN go mod download \
    && go tool templ generate -path ./components \
    && go tool go-tw -i ./styles/input.css -o ./dist/assets/css/output@${VERSION}.css --minify \
    && go tool sqlc generate \
    && go build -ldflags="-s -w -X version.Value=${VERSION}" -o pathwise

## Deploy
FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=build /app/pathwise /pathwise

EXPOSE 8080

CMD ["/pathwise"]
