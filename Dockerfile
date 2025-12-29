# Multi-stage build for smaller final image
# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install grpcui using Go
RUN GO111MODULE=on go install github.com/fullstorydev/grpcui/cmd/grpcui@latest

# Copy the entire project
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/oms-api/main.go

# Stage 2: Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and netcat for healthcheck
RUN apk --no-cache add ca-certificates netcat-openbsd

# Create a non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy grpcui binary from builder
COPY --from=builder /go/bin/grpcui /usr/local/bin/grpcui

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports: 8089 for gRPC, 8080 for grpcui
EXPOSE 8089 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD nc -z localhost 8089 || exit 1

# Command to run the application
CMD ["./main"]
