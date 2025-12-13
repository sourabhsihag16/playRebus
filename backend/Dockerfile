# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# Debug: Show what we have
RUN echo "=== Current directory ===" && \
    pwd && \
    echo "=== Directory listing ===" && \
    ls -la && \
    echo "=== Looking for main.go ===" && \
    find . -name "main.go" -type f && \
    echo "=== Root directory ===" && \
    ls -la . 2>/dev/null || echo "current dir listing failed"

# Build - try the most likely path first
RUN if [ -f "main.go" ]; then \
        echo "✓ Found main.go - building..." && \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .; \
    else \
        echo "✗ main.go not found. Available files:" && \
        find . -type f -name "*.go" | head -10 && \
        exit 1; \
    fi

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server .

# Create storage directory for images (if using local storage)
RUN mkdir -p /app/storage/images

# Expose port (default 8080, Railway will override with PORT env var)
EXPOSE 8080

# Run the server
CMD ["./server"]
