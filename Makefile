# Build configuration
APP_NAME := maniplacer
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR := dist
CMD_PATH := ./cmd
RELEASE_FILE := $(BUILD_DIR)/$(APP_NAME)-binaries-$(VERSION).tar.gz
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

# Supported architectures
OS_ARCHS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Color codes
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m

# LDFLAGS for version injection
LDFLAGS := -X github.com/dantedelordran/maniplacer/internal/utils.Version=$(VERSION)

.PHONY: all clean build build-all install release test lint

all: clean build

clean:
	@echo -e "${BLUE}ℹ️  Cleaning build artifacts...${NC}"
	@rm -rf $(BUILD_DIR)
	@mkdir -p $(BUILD_DIR)

build: clean
	@echo -e "${BLUE}ℹ️  Building for current architecture (version: $(VERSION))...${NC}"
	@CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo -e "${GREEN}✅ Build complete: $(BUILD_DIR)/$(APP_NAME)${NC}"

build-all: clean
	@echo -e "${BLUE}ℹ️  Building for all architectures (version: $(VERSION))...${NC}"
	@set -e; for os_arch in $(OS_ARCHS); do \
		os=$${os_arch%/*}; \
		arch=$${os_arch#*/}; \
		output="$(BUILD_DIR)/$(APP_NAME)-$${os}-$${arch}"; \
		if [ "$${os}" = "windows" ]; then output="$${output}.exe"; fi; \
		echo -e "${BLUE}ℹ️  Building $${os}/$${arch}...${NC}"; \
		CGO_ENABLED=0 GOOS=$${os} GOARCH=$${arch} go build -trimpath -ldflags "$(LDFLAGS)" -o "$${output}" $(CMD_PATH); \
		echo -e "${GREEN}✅ Built: $${output}${NC}"; \
	done
	@if command -v sha256sum >/dev/null 2>&1; then \
		cd $(BUILD_DIR) && sha256sum $(APP_NAME)-linux-amd64 $(APP_NAME)-linux-arm64 $(APP_NAME)-darwin-amd64 $(APP_NAME)-darwin-arm64 $(APP_NAME)-windows-amd64.exe > checksums.txt; \
	else \
		cd $(BUILD_DIR) && shasum -a 256 $(APP_NAME)-linux-amd64 $(APP_NAME)-linux-arm64 $(APP_NAME)-darwin-amd64 $(APP_NAME)-darwin-arm64 $(APP_NAME)-windows-amd64.exe > checksums.txt; \
	fi
	@echo -e "${GREEN}✅ All builds complete. Artifacts in $(BUILD_DIR)/${NC}"

install: build
	@mkdir -p "$(DESTDIR)$(BINDIR)"
	@install -m 0755 "$(BUILD_DIR)/$(APP_NAME)" "$(DESTDIR)$(BINDIR)/$(APP_NAME)"
	@echo -e "${GREEN}✅ Installed: $(DESTDIR)$(BINDIR)/$(APP_NAME)${NC}"

release: build-all
	@echo -e "${BLUE}ℹ️  Preparing release artifacts...${NC}"
	@cd $(BUILD_DIR) && \
		tar -czvf $(notdir $(RELEASE_FILE)) \
		$(APP_NAME)-linux-amd64 \
		$(APP_NAME)-linux-arm64 \
		$(APP_NAME)-darwin-amd64 \
		$(APP_NAME)-darwin-arm64 \
		$(APP_NAME)-windows-amd64.exe \
		checksums.txt
	@echo -e "${GREEN}✅ Release archive created: $(RELEASE_FILE)${NC}"

test:
	@echo -e "${BLUE}ℹ️  Running tests...${NC}"
	@go test -v -race -cover ./...
	@echo -e "${GREEN}✅ Tests complete${NC}"

lint:
	@echo -e "${BLUE}ℹ️  Running linters...${NC}"
	@go vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo -e "${YELLOW}⚠️  staticcheck not installed, skipping...${NC}"; \
	fi
	@echo -e "${GREEN}✅ Linting complete${NC}"
