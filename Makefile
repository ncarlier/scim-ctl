# Makefile for SCIM CTL

.PHONY: build test clean install help run-setup run-interactive

# Default Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=scim-ctl
MAIN_PATH=./cmd/scim-ctl

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: $(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
	@echo "✅ Tests complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	@echo "✅ Clean complete"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies updated"

# Install binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BINARY_NAME) $(shell go env GOPATH)/bin/
	@echo "✅ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

# Run the setup wizard using the bash script
run-setup: build
	@echo "Running setup wizard..."
	./scim.sh setup

# Run interactive mode using the bash script
run-interactive: build
	@echo "Starting interactive mode..."
	./scim.sh interactive

# Build and show help
run-help: build
	@echo "SCIM CTL Help:"
	@echo "=============="
	./$(BINARY_NAME) --help
	@echo
	@echo "Bash Script Help:"
	@echo "=================="
	./scim.sh --help

# Lint the code (requires golangci-lint)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running linter..."; \
		golangci-lint run; \
		echo "✅ Lint complete"; \
	else \
		echo "⚠️  golangci-lint not installed. Install with:"; \
		echo "   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "✅ Format complete"

# Generate documentation
docs:
	@echo "Current project structure:"
	@tree -I 'vendor|.git' . || echo "Install 'tree' command for better output"
	@echo
	@echo "Available commands:"
	@make help

# Show available make targets
help:
	@echo "Available targets:"
	@echo "  build         - Build the scim-ctl binary"
	@echo "  test          - Run tests"
	@echo "  clean         - Remove build artifacts" 
	@echo "  deps          - Download and tidy dependencies"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  run-setup     - Run configuration setup wizard"
	@echo "  run-interactive - Start interactive mode"
	@echo "  run-help      - Show help for both CLI and script"
	@echo "  lint          - Run code linter (requires golangci-lint)"
	@echo "  fmt           - Format code"
	@echo "  docs          - Show project documentation"
	@echo "  help          - Show this help"

# Default target
all: deps fmt lint test build