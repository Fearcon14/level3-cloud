# Build for linux/amd64 (e.g. STACKIT SKE). On ARM Macs, omit --platform for local arm64.
# Build stage
FROM --platform=linux/amd64 golang:1.25-alpine AS builder

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary (no CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

# Runtime stage (same platform as builder)
FROM --platform=linux/amd64 alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Binary from builder
COPY --from=builder /build/api .

# Template required at runtime (default REDIS_FAILOVER_TEMPLATE path)
COPY --from=builder /build/internal/k8s/templates ./internal/k8s/templates

EXPOSE 8080

# Config via env: KUBECONFIG, PAAS_NAMESPACE, API_LISTEN_ADDR, REDIS_FAILOVER_TEMPLATE, PAAS_DEFAULT_STORAGE_CLASS
ENV API_LISTEN_ADDR=:8080

ENTRYPOINT ["./api"]
