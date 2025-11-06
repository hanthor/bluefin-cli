# justfile for bluefin-cli development

# Default recipe - show available commands
default:
    @just --list

# Build, test, and run in container (main development workflow)
test: build-container
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Building Go package and running tests in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        bash -c 'go build -o bluefin-cli && ./test-container.sh'

# Build the development container image (if not exists or force rebuild)
build-container:
    #!/usr/bin/env bash
    if ! podman image exists bluefin-cli-dev; then
        echo "Building development container image..."
        podman build -t bluefin-cli-dev -f Containerfile.dev .
    else
        echo "Development container image already exists (use 'just rebuild-container' to force rebuild)"
    fi

# Force rebuild the development container
rebuild-container:
    @echo "Rebuilding development container image..."
    podman build -t bluefin-cli-dev -f Containerfile.dev .

# Run unit tests in container
unit-test: build-container
    @echo "Running unit tests in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        go test ./... -v

# Build the binary in container
build: build-container
    @echo "Building binary in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        go build -o bluefin-cli

# Run integration tests in container
integration-test: build-container
    @echo "Running integration tests in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        bash -c 'go build -o bluefin-cli && ./test-integration.sh'

# Run container tests (comprehensive shell modification tests)
container-test: build-container
    @echo "Running comprehensive container tests..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        bash -c 'go build -o bluefin-cli && cp bluefin-cli /usr/local/bin/ && ./test-container.sh'

# Open an interactive shell in the development container
shell: build-container
    @echo "Opening interactive shell in development container..."
    podman run --rm -it \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        bash

# Clean up built artifacts
clean:
    @echo "Cleaning up built artifacts..."
    rm -f bluefin-cli
    go clean

# Clean up container images
clean-containers:
    @echo "Removing development container image..."
    -podman rmi bluefin-cli-dev
    -podman rmi bluefin-cli-test

# Full clean (artifacts + containers)
clean-all: clean clean-containers

# Run linter in container
lint: build-container
    @echo "Running linter in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        bash -c 'command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping..."'

# Format code
fmt:
    @echo "Formatting Go code..."
    go fmt ./...

# Show Go module info
mod-info: build-container
    @echo "Go module information:"
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        go version
