.PHONY: all build build-all clean test fmt run install dev install-hooks desktop-install desktop-dev desktop-build desktop-build-all

# Binary name
BINARY_NAME=aurora
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags for embedding version
# -s: strip symbol table, -w: strip DWARF debug info (reduces binary size ~30-40%)
LDFLAGS="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
all: build

# Build for current platform
build: install
	go build -trimpath -ldflags=$(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd

# Build all common platforms
build-all: install
	mkdir -p dist/linux-amd64 dist/linux-arm64 dist/darwin-amd64 dist/darwin-arm64 dist/windows-amd64
	GOOS=linux   GOARCH=amd64    go build -trimpath -ldflags=$(LDFLAGS) -o dist/linux-amd64/$(BINARY_NAME)      ./cmd
	GOOS=linux   GOARCH=arm64    go build -trimpath -ldflags=$(LDFLAGS) -o dist/linux-arm64/$(BINARY_NAME)      ./cmd
	GOOS=darwin  GOARCH=amd64    go build -trimpath -ldflags=$(LDFLAGS) -o dist/darwin-amd64/$(BINARY_NAME)     ./cmd
	GOOS=darwin  GOARCH=arm64    go build -trimpath -ldflags=$(LDFLAGS) -o dist/darwin-arm64/$(BINARY_NAME)     ./cmd
	GOOS=windows GOARCH=amd64    go build -trimpath -ldflags=$(LDFLAGS) -o dist/windows-amd64/$(BINARY_NAME).exe ./cmd
	@echo "✓ Binaries built successfully in dist/"
	@echo "  Creating compressed archives..."
	@cd dist && tar -czf $(BINARY_NAME)-linux-amd64.tar.gz   -C linux-amd64   $(BINARY_NAME)     && echo "  - $(BINARY_NAME)-linux-amd64.tar.gz"
	@cd dist && tar -czf $(BINARY_NAME)-linux-arm64.tar.gz   -C linux-arm64   $(BINARY_NAME)     && echo "  - $(BINARY_NAME)-linux-arm64.tar.gz"
	@cd dist && tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz  -C darwin-amd64  $(BINARY_NAME)     && echo "  - $(BINARY_NAME)-darwin-amd64.tar.gz"
	@cd dist && tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz  -C darwin-arm64  $(BINARY_NAME)     && echo "  - $(BINARY_NAME)-darwin-arm64.tar.gz"
	@cd dist && zip -q $(BINARY_NAME)-windows-amd64.zip      -j windows-amd64/$(BINARY_NAME).exe && echo "  - $(BINARY_NAME)-windows-amd64.zip"
	@rm -rf dist/linux-amd64 dist/linux-arm64 dist/darwin-amd64 dist/darwin-arm64 dist/windows-amd64
	@echo "✓ Compressed archives created"

clean:
	rm -rf bin/ dist/ dist-desktop/

test: install
	go test ./cmd/... ./pkg/... ./internal/... -v

fmt:
	go fmt ./...

install:
	go mod tidy

# For local development
run: build
	./bin/$(BINARY_NAME) config

# Run without building (development mode)
dev: install
	go run ./cmd config

# Install git hooks
install-hooks:
	@echo "Installing git hooks..."
	@mkdir -p .git/hooks
	@cp hooks/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✓ Pre-commit hook installed successfully"
	@echo "  The hook will run 'make fmt' before each commit"

# Desktop app targets (Wails)
# Note: webkit2_41 tag required for Ubuntu 24.04+ (webkit2gtk-4.1)
DESKTOP_LDFLAGS=-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

desktop-install:
	cd desktop/frontend && bun install

desktop-dev: desktop-install
	cd desktop && wails dev -tags webkit2_41

desktop-build: desktop-install
	cd desktop && wails build -tags webkit2_41 -ldflags "$(DESKTOP_LDFLAGS)"

# Desktop name for releases
DESKTOP_NAME=aurora-desktop

# Build desktop app for all platforms (cross-compilation from Linux)
# Note: Only linux/amd64 and windows/amd64 reliably work from Linux
#       arm64 and darwin require native builds or cross-compiler toolchains
# Uses webkit2_41 tag for Ubuntu 24.04+ compatibility
desktop-build-all: desktop-install
	@echo "Building desktop app for all platforms..."
	@mkdir -p dist-desktop
	cd desktop && wails build -tags webkit2_41 -platform linux/amd64 -o $(DESKTOP_NAME) && \
		mv build/bin/$(DESKTOP_NAME) ../dist-desktop/$(DESKTOP_NAME)-linux-amd64
	-cd desktop && wails build -tags webkit2_41 -platform linux/arm64 -o $(DESKTOP_NAME) && \
		mv build/bin/$(DESKTOP_NAME) ../dist-desktop/$(DESKTOP_NAME)-linux-arm64
	cd desktop && wails build -tags webkit2_41 -platform windows/amd64 -o $(DESKTOP_NAME).exe && \
		mv build/bin/$(DESKTOP_NAME).exe ../dist-desktop/$(DESKTOP_NAME)-windows-amd64.exe
	-cd desktop && wails build -tags webkit2_41 -platform darwin/amd64 -o $(DESKTOP_NAME) && \
		mv build/bin/$(DESKTOP_NAME) ../dist-desktop/$(DESKTOP_NAME)-darwin-amd64
	-cd desktop && wails build -tags webkit2_41 -platform darwin/arm64 -o $(DESKTOP_NAME) && \
		mv build/bin/$(DESKTOP_NAME) ../dist-desktop/$(DESKTOP_NAME)-darwin-arm64
	@echo "✓ Desktop binaries built in dist-desktop/"
	@echo "  Creating compressed archives..."
	@cd dist-desktop && for f in $(DESKTOP_NAME)-linux-* $(DESKTOP_NAME)-darwin-*; do \
		[ -f "$$f" ] && tar -czf "$$f.tar.gz" "$$f" && rm "$$f" && echo "  - $$f.tar.gz"; \
	done; true
	@cd dist-desktop && [ -f $(DESKTOP_NAME)-windows-amd64.exe ] && \
		zip -q $(DESKTOP_NAME)-windows-amd64.zip $(DESKTOP_NAME)-windows-amd64.exe && \
		rm $(DESKTOP_NAME)-windows-amd64.exe && echo "  - $(DESKTOP_NAME)-windows-amd64.zip" || true
	@echo "✓ Compressed archives created"