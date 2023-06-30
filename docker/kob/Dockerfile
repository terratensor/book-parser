# syntax = docker/dockerfile:1-experimental
FROM golang:latest AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o ./kob-library-parser ./vpsssr/cmd

# 2

FROM scratch

LABEL org.opencontainers.image.source = https://github.com/terratendor/book-parser

WORKDIR /app

COPY --from=builder /app/kob-library-parser /app/kob-library-parser
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Europe/Moscow

ENV PARSER_STORE=gorm

CMD ["./kob-library-parser"]

