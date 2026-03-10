.PHONY: dev build build:mac build:win build:linux clean install

# Install dependencies
install:
	cd frontend && yarn install

# Development
dev:
	wails dev

# Build all platforms
build:
	wails build

# Build for macOS
build:mac:
	wails build -platform=darwin/arm64
	wails build -platform=darwin/amd64

# Build for Windows
build:win:
	wails build -platform=windows/amd64

# Build for Linux
build:linux:
	wails build -platform=linux/arm64
	wails build -platform=linux/amd64

# Clean build artifacts
clean:
	rm -rf build/*
