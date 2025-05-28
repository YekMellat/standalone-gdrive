# Build stage
FROM golang:1.21-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X github.com/standalone-gdrive/version.Version=${VERSION}" -o gdrive ./cmd/gdrive
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o token ./cmd/token
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o oauth_test ./cmd/oauth_test

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh gdrive

# Set working directory
WORKDIR /home/gdrive

# Copy binaries from builder
COPY --from=builder /app/gdrive /usr/local/bin/
COPY --from=builder /app/token /usr/local/bin/
COPY --from=builder /app/oauth_test /usr/local/bin/

# Copy documentation
COPY --from=builder /app/README.md ./
COPY --from=builder /app/docs ./docs/
COPY --from=builder /app/examples ./examples/

# Create config directory
RUN mkdir -p .config/standalone-gdrive && chown -R gdrive:gdrive .config

# Switch to non-root user
USER gdrive

# Set default command
ENTRYPOINT ["gdrive"]
CMD ["--help"]

# Add labels
LABEL org.opencontainers.image.title="Standalone Google Drive Client"
LABEL org.opencontainers.image.description="A lightweight, standalone Google Drive client based on rclone"
LABEL org.opencontainers.image.source="https://github.com/YekMellat/standalone-gdrive"
LABEL org.opencontainers.image.licenses="MIT"
