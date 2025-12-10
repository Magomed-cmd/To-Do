# syntax=docker/dockerfile:1.4
# ---------- Build stage ----------
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy dependency files first for caching
COPY services/notification-service/go.mod services/notification-service/go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY services/notification-service ./notification-service

WORKDIR /app/notification-service

# Build arguments for cross-compilation
ARG TARGETOS TARGETARCH

# Build binary with cache mount for faster rebuilds
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags='-w -s' \
    -o main ./cmd/main.go

# ---------- Runtime stage ----------
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

COPY --from=builder /app/notification-service/main ./main

RUN chown -R appuser:appgroup /app
USER appuser

ENV SERVICE_PORT=8084
EXPOSE 8084

CMD ["./main"]
