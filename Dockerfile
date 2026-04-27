# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -o server cmd/server/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates gcc musl-dev

# Copy binary from builder
COPY --from=builder /app/server .
COPY --from=builder /app/internal ./internal

# Expose port
EXPOSE 8080

# Run the server
CMD ["./server"]
