# ZenNetwork Dockerfile
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache \
    git \
    build-base \
    linux-headers \
    ca-certificates \
    pkgconfig

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o zennetworkd ./cmd/zennetworkd/

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create app user
RUN addgroup -g 1000 -S zennetwork && \
    adduser -u 1000 -S zennetwork -G zennetwork

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/zennetworkd .

# Copy config files
COPY --from=builder /app/config ./config

# Change ownership
RUN chown -R zennetwork:zennetwork /app

# Switch to non-root user
USER zennetwork

# Expose ports
EXPOSE 26656 26657 8545 8546 30303 30304

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./zennetworkd status || exit 1

# Default command
CMD ["./zennetworkd", "start"]
