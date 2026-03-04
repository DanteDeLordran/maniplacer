# Build configuration
APP_NAME := maniplacer
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR := dist
CMD_PATH := cmd/main.go
RELEASE_FILE := $(BUILD_DIR)/$(APP_NAME)-binaries-$(VERSION).tar.gz

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

.PHONY: all clean build build-all release test lint

all: clean build

clean:
	@echo -e "${BLUE}ℹ️  Cleaning build artifacts...${NC}"
	@rm -rf $(BUILD_DIR)
	@mkdir -p $(BUILD_DIR)

build: clean
	@echo -e "${BLUE}ℹ️  Building for current architecture (version: $(VERSION))...${NC}"
	@go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo -e "${GREEN}✅ Build complete: $(BUILD_DIR)/$(APP_NAME)${NC}"

build-all: clean
	@echo -e "${BLUE}ℹ️  Building for all architectures (version: $(VERSION))...${NC}"
	@for os_arch in $(OS_ARCHS); do \
		os=$${os_arch%/*}; \
		arch=$${os_arch#*/}; \
		output="$(BUILD_DIR)/$(APP_NAME)-$${os}-$${arch}"; \
		if [ "$${os}" = "windows" ]; then output="$${output}.exe"; fi; \
		echo -e "${BLUE}ℹ️  Building $${os}/$${arch}...${NC}"; \
		GOOS=$${os} GOARCH=$${arch} go build -ldflags "$(LDFLAGS)" -o $${output} $(CMD_PATH); \
		if [ $$? -eq 0 ]; then \
			echo -e "${GREEN}✅ Built: $${output}${NC}"; \
		else \
			echo -e "${RED}❌ Failed to build $${os}/$${arch}${NC}"; \
		fi; \
	done
	@echo -e "${GREEN}✅ All builds complete. Artifacts in $(BUILD_DIR)/${NC}"

release: build-all
	@echo -e "${BLUE}ℹ️  Preparing release artifacts...${NC}"
	@cd $(BUILD_DIR) && \
		tar -czvf $(notdir $(RELEASE_FILE)) \
		$(APP_NAME)-linux-amd64 \
		$(APP_NAME)-linux-arm64 \
		$(APP_NAME)-darwin-amd64 \
		$(APP_NAME)-darwin-arm64 \
		$(APP_NAME)-windows-amd64.exe
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
