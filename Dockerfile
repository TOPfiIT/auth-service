FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download -x && go mod verify

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /main \
    ./cmd

FROM alpine:latest AS runtime

LABEL org.opencontainers.image.title="Auth Service" \
    org.opencontainers.image.description="Authentication service for AI Interviewer" \
    org.opencontainers.image.vendor="TOPfiIT"

COPY --from=builder /main /main

EXPOSE 8080

ENTRYPOINT ["/main"]