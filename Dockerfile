# syntax=docker/dockerfile:1

# ---------- builder ----------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git build-base openssl-dev pkgconf

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/geoguide_api ./cmd/api

# ---------- runtime ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget openssl libc6-compat

WORKDIR /app

COPY --from=builder /out/geoguide_api ./geoguide_api
COPY config ./config

EXPOSE 8080

ENTRYPOINT ["./geoguide_api"]