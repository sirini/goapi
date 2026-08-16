# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26.4
FROM golang:${GO_VERSION}-bookworm AS go-toolchain

FROM ubuntu:22.04 AS build

ENV DEBIAN_FRONTEND=noninteractive
ENV PATH=/usr/local/go/bin:${PATH}

RUN apt-get update \
    && apt-get install --yes --no-install-recommends \
        build-essential \
        ca-certificates \
        libvips-dev \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

COPY --from=go-toolchain /usr/local/go /usr/local/go

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/goapi-linux ./cmd

FROM ubuntu:22.04 AS verify

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install --yes --no-install-recommends binutils libvips42 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/goapi-linux /usr/local/bin/goapi-linux
RUN ldd /usr/local/bin/goapi-linux \
    && ! ldd /usr/local/bin/goapi-linux | grep --quiet "not found" \
    && ! readelf --version-info /usr/local/bin/goapi-linux | grep --extended-regexp "GLIBC_2\.(3[6-9]|[4-9][0-9])"

FROM ubuntu:24.04 AS verify-newer

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install --yes --no-install-recommends libvips42 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=verify /usr/local/bin/goapi-linux /usr/local/bin/goapi-linux
RUN ldd /usr/local/bin/goapi-linux \
    && ! ldd /usr/local/bin/goapi-linux | grep --quiet "not found"

FROM scratch AS artifact
COPY --from=verify-newer /usr/local/bin/goapi-linux /goapi-linux
