# Build stage
FROM golang:1.25.7-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /build/shoppilot \
    .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 shoppilot && \
    adduser -D -u 1000 -G shoppilot shoppilot

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/shoppilot /app/shoppilot

# Copy migrations
COPY --from=builder /build/migrations /app/migrations

# Change ownership
RUN chown -R shoppilot:shoppilot /app

# Switch to non-root user
USER shoppilot

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

# Run the application
CMD ["/app/shoppilot"]
