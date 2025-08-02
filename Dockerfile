# Multi-stage build for Go SMTP server
FROM golang:1.21-alpine AS go-builder

WORKDIR /app

# Copy Go modules first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o nullmail ./cmd/nullmail/main.go

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=go-builder /app/nullmail .

# Expose SMTP port
EXPOSE 2525

# Run the SMTP server
CMD ["./nullmail"]