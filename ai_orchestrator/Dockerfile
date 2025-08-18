FROM golang:1.24-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY ai_orchestrator/go.mod ai_orchestrator/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY ai_orchestrator/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy .env file from the ai_orchestrator directory
COPY ai_orchestrator/.env .env

# Run the binary
CMD ["./main"]

