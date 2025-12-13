# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# List directory structure for debugging
RUN echo "=== Directory structure ===" && \
    ls -la && \
    echo "=== cmd directory ===" && \
    ls -la cmd/ 2>/dev/null || echo "cmd/ not found" && \
    echo "=== cmd/server directory ===" && \
    ls -la cmd/server/ 2>/dev/null || echo "cmd/server/ not found"

# Build the application - try multiple path formats
RUN if [ -f "cmd/server/main.go" ]; then \
        echo "Building from cmd/server/main.go" && \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go; \
    elif [ -f "./cmd/server/main.go" ]; then \
        echo "Building from ./cmd/server/main.go" && \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server/main.go; \
    elif [ -d "./cmd/server" ]; then \
        echo "Building from ./cmd/server" && \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server; \
    else \
        echo "ERROR: Cannot find cmd/server directory or main.go file" && \
        find . -name "main.go" -type f && \
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
