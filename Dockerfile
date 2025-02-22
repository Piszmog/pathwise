## Build
FROM golang:1.23-alpine AS build

ARG VERSION='dev'

RUN apk update && apk add --no-cache curl

RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64 \
	&& mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

RUN go install github.com/a-h/templ/cmd/templ@v0.3.819 \
    && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

WORKDIR /app

COPY ./ /app

RUN templ generate -path ./components \
	&& tailwindcss -i ./styles/input.css -o ./dist/assets/css/output@${VERSION}.css --minify \
	&& sqlc generate

RUN go build -ldflags="-s -w -X version.Value=${VERSION}" -o pathwise

## Deploy
FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=build /app/pathwise /pathwise

EXPOSE 8080

CMD ["/pathwise"]
