# syntax = docker/dockerfile:1-experimental
FROM golang:latest AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o ./book-parser-common ./common/cmd

# 2

FROM scratch

WORKDIR /app

COPY --from=builder /app/book-parser-common /app/book-parser-common
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Europe/Moscow

ENV PARSER_STORE=gorm

CMD ["./book-parser-common"]

