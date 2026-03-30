BINARY_NAME=ol
VERSION=1.01

BUILD_DIR=./build

.PHONY: all clean install build-all release

all: build

build:
	go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

install: build
	sudo cp $(BINARY_NAME) /usr/local/bin/

# Cross compilation for all platforms
build-all: clean
	mkdir -p $(BUILD_DIR)
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=linux GOARCH=386 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-386 .
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe .
	@echo "Build complete. Binaries in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Create release archives
release: build-all
	mkdir -p $(BUILD_DIR)/release
	# Create tar.gz for Unix-like systems
	tar -czf $(BUILD_DIR)/release/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-amd64
	tar -czf $(BUILD_DIR)/release/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-arm64
	tar -czf $(BUILD_DIR)/release/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-amd64
	tar -czf $(BUILD_DIR)/release/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-arm64
	tar -czf $(BUILD_DIR)/release/$(BINARY_NAME)-$(VERSION)-linux-386.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-386
	# Create zip for Windows
	cd $(BUILD_DIR) && zip release/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	cd $(BUILD_DIR) && zip release/$(BINARY_NAME)-$(VERSION)-windows-arm64.zip $(BINARY_NAME)-windows-arm64.exe
	@echo "Release archives created in $(BUILD_DIR)/release/"
	@ls -la $(BUILD_DIR)/release/

# Development
dev:
	go run . --help

test:
	go test -v ./...
