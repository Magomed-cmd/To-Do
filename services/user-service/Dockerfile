# ---------- Build stage ----------
FROM golang:alpine AS builder

WORKDIR /app

# Install git for go modules
RUN apk add --no-cache git

# Copy dependency files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source tree
COPY . .

# Set build context to user-service
WORKDIR /app/services/user-service

# Build binary (static, optimized)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
-ldflags='-w -s -extldflags "-static"' \
-a -installsuffix cgo \
-o main ./cmd/main.go


# ---------- Runtime stage ----------
FROM alpine:latest

# Install certs and timezone data
RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
adduser -u 1001 -S appuser -G appgroup

# Copy compiled binary from builder
COPY --from=builder /app/services/user-service/main ./main

# Ownership to non-root
RUN chown -R appuser:appgroup /app
USER appuser

# Allow dynamic port configuration
ARG SERVICE_PORT=8081
ENV SERVICE_PORT=${SERVICE_PORT}
EXPOSE ${SERVICE_PORT}

# Health check endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
CMD wget --no-verbose --tries=1 --spider http://localhost:${SERVICE_PORT}/health || exit 1

# Run binary
CMD ["./main"]