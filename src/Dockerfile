# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
WORKDIR /app

# Copy dependency files and download first (for layer caching)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -trimpath -o product-service .

# Runtime stage using distroless for security and minimal size
FROM gcr.io/distroless/base-debian12
WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/product-service ./product-service

# Expose port
EXPOSE 8080

# Health check for container orchestration (ECS, Kubernetes)
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/product-service", "-healthcheck"]

# Run as non-root (distroless defaults to nonroot user)
USER nonroot:nonroot

ENTRYPOINT ["/app/product-service"]
