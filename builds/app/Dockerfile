FROM golang:1.22-bookworm AS build

WORKDIR /app

COPY . .

RUN set -ex \
# ---- project builds
 && go mod download \
 && go build -o app cmd/app/main.go

FROM debian:bookworm

WORKDIR /app

EXPOSE 8080

COPY --from=build /app/app /usr/bin/app

CMD ["app"]
