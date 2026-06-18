# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

ENV GOPROXY=https://goproxy.cn,direct \
    GOMODCACHE=/go/pkg/mod \
    GOCACHE=/root/.cache/go-build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/netdisk-server ./cmd/server

FROM alpine:3.22

RUN --mount=type=cache,target=/var/cache/apk,sharing=locked \
    apk add ca-certificates ffmpeg tzdata \
    && addgroup -S netdisk \
    && adduser -S -G netdisk netdisk

WORKDIR /app

COPY --from=builder /out/netdisk-server /usr/local/bin/netdisk-server
COPY config.docker.yaml /app/config.yaml

RUN mkdir -p /data /app/logs \
    && chown -R netdisk:netdisk /data /app/logs

USER netdisk

EXPOSE 8080

ENTRYPOINT ["netdisk-server"]
CMD ["--config", "/app/config.yaml"]
