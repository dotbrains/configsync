.PHONY: build test clean install help

# Build the binary
build:
	go build -o configsync ./cmd/configsync

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o bin/configsync-darwin-amd64 ./cmd/configsync
	GOOS=darwin GOARCH=arm64 go build -o bin/configsync-darwin-arm64 ./cmd/configsync

# Install to /usr/local/bin
install: build
	sudo cp configsync /usr/local/bin/

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -f configsync
	rm -rf bin/
	rm -f coverage.out

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Mod tidy
tidy:
	go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  build         Build the binary"
	@echo "  build-all     Build for multiple platforms"
	@echo "  install       Install to /usr/local/bin"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage"
	@echo "  clean         Clean build artifacts"
	@echo "  fmt           Format code"
	@echo "  lint          Run linter"
	@echo "  tidy          Run go mod tidy"
	@echo "  help          Show this help"