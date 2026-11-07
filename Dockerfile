# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25.7-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
	go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
	go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/pgo-mcp ./src

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build --chown=65532:65532 /out/pgo-mcp /app/pgo-mcp

USER 65532:65532

EXPOSE 8001

ENTRYPOINT ["/app/pgo-mcp"]
