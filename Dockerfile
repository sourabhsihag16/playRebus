# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy backend go mod files
COPY backend/go.mod backend/go.sum ./backend/

# Download dependencies
WORKDIR /app/backend
RUN go mod download

# Copy all backend source code
WORKDIR /app
COPY backend/ ./backend/

# Build the application
WORKDIR /app/backend
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/backend/server .

# Create storage directory for images (if using local storage)
RUN mkdir -p /app/storage/images

# Expose port (default 8080, Railway will override with PORT env var)
EXPOSE 8080

# Run the server
CMD ["./server"]

