# Containerfile for bluefin-cli integration testing
FROM docker.io/library/golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git bash

# Copy source
COPY . .

# Build
RUN go build -o bluefin-cli

# Test stage
FROM docker.io/library/alpine:latest

# Install runtime dependencies for testing
RUN apk add --no-cache bash zsh fish curl git

# Create test user
RUN adduser -D testuser

# Copy built binary
COPY --from=builder /app/bluefin-cli /usr/local/bin/bluefin-cli

# Set up test environment
USER testuser
WORKDIR /home/testuser

# Run tests as the test user
CMD ["sh", "-c", "bluefin-cli --version && bluefin-cli status && bluefin-cli bling bash on && bluefin-cli motd show"]
