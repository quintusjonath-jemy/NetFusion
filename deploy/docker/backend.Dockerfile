# ==============================================================================
# NetFusion Backend Dockerfile
# Multi-stage build with Linux networking capabilities (iproute2 & CA certs)
# ==============================================================================

FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Cache Go modules
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./

# Build optimized binary
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /netfusion-server ./cmd/server/main.go

# Runtime stage
FROM alpine:3.20

# Install iproute2 for Linux MPTCP kernel utilities (ip mptcp) and ca-certificates
RUN apk add --no-cache ca-certificates iproute2 tzdata

WORKDIR /app

COPY --from=builder /netfusion-server /app/netfusion-server

# Default environment variables
ENV HOST=0.0.0.0 \
    PORT=8080 \
    SIMULATION_MODE=false \
    DB_DRIVER=postgres \
    LOG_LEVEL=INFO

EXPOSE 8080

ENTRYPOINT ["/app/netfusion-server"]
