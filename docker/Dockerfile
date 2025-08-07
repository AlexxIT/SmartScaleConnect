# syntax=docker/dockerfile:labs

# 1. Build binary
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

WORKDIR /build

RUN apk add git

# Cache dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -ldflags "-s -w" -trimpath -o scaleconnect


# 2. Final image
FROM alpine AS base

RUN apk add --no-cache tini jq

COPY docker/run.sh /
RUN chmod a+x /run.sh

COPY --from=build /build/scaleconnect /usr/local/bin/

ENTRYPOINT ["/sbin/tini", "--"]
VOLUME /config
WORKDIR /config

CMD ["/run.sh"]
