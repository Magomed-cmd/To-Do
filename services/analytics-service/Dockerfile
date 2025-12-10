# ---------- Build stage ----------
FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/services/analytics-service

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/main.go

# ---------- Runtime stage ----------
FROM alpine:latest

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
