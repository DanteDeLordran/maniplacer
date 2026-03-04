# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with version injection
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/dantedelordran/maniplacer/internal/utils.Version=${VERSION}" \
    -o dist/maniplacer \
    cmd/main.go

# Final stage - use distroless for security
FROM gcr.io/distroless/static:nonroot

WORKDIR /workspace

# Copy binary from builder
COPY --from=builder /app/dist/maniplacer /usr/local/bin/maniplacer

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Run as non-root user (distroless default)
USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/maniplacer"]
