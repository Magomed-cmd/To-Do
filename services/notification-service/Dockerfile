# ---------- Build stage ----------
FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY services/notification-service/go.mod services/notification-service/go.sum ./
RUN go mod download

COPY services/notification-service ./notification-service

WORKDIR /app/notification-service

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/main.go

# ---------- Runtime stage ----------
FROM alpine:latest

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
