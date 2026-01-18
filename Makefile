# lu Makefile
# GitHub: https://github.com/ipanardian/lu-hut
# Author: Ipan Ardian
# Version: v1.0.0

.PHONY: build build-linux build-mac build-all install install-linux install-mac clean test help

# Default target
all: build

# Build for current platform
build:
	go build -ldflags="-s -w" -o bin/lu ./cmd/lu

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/lu-linux-amd64 ./cmd/lu

# Build for macOS
build-mac:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/lu-darwin-amd64 ./cmd/lu
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/lu-darwin-arm64 ./cmd/lu

# Build for all platforms
build-all: build-linux build-mac

# Install to ~/bin (current platform)
install: build
	@echo "Installing lu to ~/bin..."
	@mkdir -p ~/bin
	@cp bin/lu ~/bin/
	@echo "Installation complete! Make sure ~/bin is in your PATH."

# Install Linux binary
install-linux: build-linux
	@echo "Installing lu for Linux..."
	@mkdir -p ~/bin
	@cp lu-linux-amd64 ~/bin/lu
	@echo "Installation complete!"

# Install macOS binary
install-mac: build-mac
	@echo "Installing lu for macOS..."
	@if [ "$(shell uname -m)" = "arm64" ]; then \
		cp lu-darwin-arm64 ~/bin/lu; \
	else \
		cp lu-darwin-amd64 ~/bin/lu; \
	fi
	@mkdir -p ~/bin
	@echo "Installation complete!"

# Install to /usr/local/bin (requires sudo)
install-system: build
	@echo "Installing lu to /usr/local/bin..."
	@sudo cp bin/lu /usr/local/bin/
	@echo "Installation complete! lu is now available system-wide."

# Run lu from source (pass args with ARGS, e.g. make run ARGS="-tg")
ARGS ?=
run:
	@echo "Running lu from source with args: $(ARGS)"
	@go run ./cmd/lu $(ARGS)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f bin/lu bin/lu-*
	@go clean -cache
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the lu binary"
	@echo "  install        - Install lu to ~/bin"
	@echo "  install-system - Install lu to /usr/local/bin (requires sudo)"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run tests"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  help           - Show this help message"
