.PHONY: build test clean install docker run-example

# Build the planguard binary
build:
	@echo "Building planguard..."
	@go build -o bin/planguard ./cmd/planguard

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean

# Install planguard locally
install:
	@echo "Installing planguard..."
	@go install ./cmd/planguard

# Build Docker image
docker:
	@echo "Building Docker image..."
	@docker build -t planguard:latest .

# Run on example files (note: uses explicit -rules-dir since examples use repo rules)
run-example: build
	@echo "Running planguard on example files..."
	@./bin/planguard \
		-config examples/.planguard/config.hcl \
		-directory examples/terraform \
		-rules-dir rules

# Run with JSON output
run-example-json: build
	@./bin/planguard \
		-config examples/.planguard/config.hcl \
		-directory examples/terraform \
		-rules-dir rules \
		-format json

# Run with SARIF output
run-example-sarif: build
	@./bin/planguard \
		-config examples/.planguard/config.hcl \
		-directory examples/terraform \
		-rules-dir rules \
		-format sarif

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code (optional - requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed - skipping"; \
		echo "Install: https://golangci-lint.run/usage/install/"; \
	fi

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@GOOS=linux GOARCH=amd64 go build -o bin/planguard-linux-amd64 ./cmd/planguard
	@GOOS=linux GOARCH=arm64 go build -o bin/planguard-linux-arm64 ./cmd/planguard
	@GOOS=darwin GOARCH=amd64 go build -o bin/planguard-darwin-amd64 ./cmd/planguard
	@GOOS=darwin GOARCH=arm64 go build -o bin/planguard-darwin-arm64 ./cmd/planguard
	@GOOS=windows GOARCH=amd64 go build -o bin/planguard-windows-amd64.exe ./cmd/planguard

# Quick verification that everything works
verify: build test
	@echo ""
	@echo "Running example scan to verify functionality..."
	@./bin/planguard -version
	@echo ""
	@echo "Scanning examples (expecting violations and exceptions)..."
	@./bin/planguard \
		-config examples/.planguard/config.hcl \
		-directory examples/terraform \
		-rules-dir rules \
		-format json > /tmp/planguard-test.json || true
	@if [ -s /tmp/planguard-test.json ]; then \
		VIOLATIONS=$$(cat /tmp/planguard-test.json | grep -c '"RuleID"' || echo "0"); \
		echo "✅ Verification successful! Found $$VIOLATIONS violations in example files."; \
		echo "   (Some violations are excepted - run 'make run-example' to see details)"; \
	else \
		echo "❌ Verification failed - no output generated"; \
		exit 1; \
	fi
	@rm -f /tmp/planguard-test.json

# CI-friendly scan with human-readable output for GitHub Actions logs
ci-scan: build
	@echo "🔍 Running Planguard scan..."
	@echo ""
	@./bin/planguard \
		-config examples/.planguard/config.hcl \
		-directory examples/terraform \
		-rules-dir rules \
		-format text

# Scan your own Terraform directory (uses default ~/.planguard/rules/)
# Usage: make scan DIR=/path/to/terraform
scan: build
	@if [ -z "$(DIR)" ]; then \
		echo "Error: Please specify DIR=/path/to/terraform"; \
		echo "Example: make scan DIR=./my-terraform"; \
		echo ""; \
		echo "Note: Uses rules from ~/.planguard/rules/ by default"; \
		echo "      Override with RULES=/path/to/rules"; \
		exit 1; \
	fi
	@echo "Scanning $(DIR)..."
	@if [ -n "$(RULES)" ]; then \
		./bin/planguard -directory $(DIR) -rules-dir $(RULES); \
	else \
		./bin/planguard -directory $(DIR); \
	fi

# Scan with custom config (uses default ~/.planguard/rules/)
# Usage: make scan-config CONFIG=/path/to/config.hcl DIR=/path/to/terraform
scan-config: build
	@if [ -z "$(DIR)" ] || [ -z "$(CONFIG)" ]; then \
		echo "Error: Please specify both CONFIG and DIR"; \
		echo "Example: make scan-config CONFIG=.planguard/config.hcl DIR=./terraform"; \
		exit 1; \
	fi
	@echo "Scanning $(DIR) with config $(CONFIG)..."
	@if [ -n "$(RULES)" ]; then \
		./bin/planguard -config $(CONFIG) -directory $(DIR) -rules-dir $(RULES); \
	else \
		./bin/planguard -config $(CONFIG) -directory $(DIR); \
	fi

# Show version
version: build
	@./bin/planguard -version

# Show help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build & Test:"
	@echo "  build          - Build the planguard binary"
	@echo "  test           - Run tests"
	@echo "  verify         - Build, test, and verify everything works"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download dependencies"
	@echo "  build-all      - Build for all platforms"
	@echo ""
	@echo "Run Planguard:"
	@echo "  run-example         - Run on example files (text output)"
	@echo "  run-example-json    - Run on example files (JSON output)"
	@echo "  run-example-sarif   - Run on example files (SARIF output)"
	@echo "  scan DIR=<path>     - Scan your own Terraform directory (uses ~/.planguard/rules/)"
	@echo "  scan-config CONFIG=<config> DIR=<path> - Scan with custom config"
	@echo "  version             - Show planguard version"
	@echo ""
	@echo "CI/CD:"
	@echo "  ci-scan        - Run scan with text output (for GitHub Actions logs)"
	@echo ""
	@echo "Development:"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  install        - Install planguard locally"
	@echo "  docker         - Build Docker image"
	@echo ""
	@echo "Default Locations:"
	@echo "  Config:  ./.planguard/config.hcl or ~/.planguard/config.hcl"
	@echo "  Rules:   ~/.planguard/rules/"
	@echo ""
	@echo "Setup:"
	@echo "  mkdir -p ~/.planguard/rules"
	@echo "  cp -r rules/* ~/.planguard/rules/"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make verify"
	@echo "  make ci-scan"
	@echo "  make scan DIR=./my-terraform"
	@echo "  make scan DIR=./terraform RULES=./custom-rules"
	@echo "  make scan-config CONFIG=.planguard/config.hcl DIR=./terraform"
