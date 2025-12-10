# syntax=docker/dockerfile:1.4
# ---------- Build stage ----------
FROM --platform=$BUILDPLATFORM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy dependency files first for caching
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the entire source tree
COPY . .

WORKDIR /app/services/analytics-service

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

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

COPY --from=builder /app/services/analytics-service/main ./main

RUN chown -R appuser:appgroup /app
USER appuser

ARG SERVICE_PORT=8083
ENV SERVICE_PORT=${SERVICE_PORT}
EXPOSE ${SERVICE_PORT}

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${SERVICE_PORT}/health || exit 1

CMD ["./main"]
