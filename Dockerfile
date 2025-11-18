# Multi-stage build for Go server
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY server/go.mod server/go.sum ./
RUN go mod download

# Copy source
COPY server/ ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /mmo-server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary
COPY --from=builder /mmo-server .

# Expose port
EXPOSE 8080

# Run
CMD ["./mmo-server"]
