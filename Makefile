.PHONY: all build clean deps test

# Project name
BINARY_NAME = setup-docker
BINARY_PATH = bin/$(BINARY_NAME)

# Go build parameters
GOOS = linux
GOARCH = amd64
CGO_ENABLED = 0

all: build

# Download dependencies
deps:
	go mod download
	go mod verify

# Build static binary
build: deps
	@mkdir -p bin
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags="-s -w" \
		-o $(BINARY_PATH) \
		./cmd/...
	@echo "Build complete: $(BINARY_PATH)"
	@echo "File size: $$(du -h $(BINARY_PATH) | cut -f1)"

# Compress binary (requires upx)
compress: build
	upx --best --lzma $(BINARY_PATH)
	@echo "Compressed size: $$(du -h $(BINARY_PATH) | cut -f1)"

# Run tests
test:
	go test -v ./internal/...

# Clean build artifacts
clean:
	rm -rf bin
	rm -f *.tar.gz

# Build release package
release: compress
	tar -czvf $(BINARY_NAME)-$(GOOS)-$(GOARCH).tar.gz -C bin $(BINARY_NAME) README.md
	@echo "Release package created: $(BINARY_NAME)-$(GOOS)-$(GOARCH).tar.gz"

# Test container - builds and runs in SUSE container
test-container: build
	docker build -t sles-docker-setup-test .
	docker run --rm -it --privileged sles-docker-setup-test
