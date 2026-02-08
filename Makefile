APP_NAME := pingai
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR := build/bin

.PHONY: all dev build build-mac build-mac-arm64 build-mac-amd64 build-mac-universal build-windows build-linux clean install-deps

all: build

dev:
	wails dev

build:
	wails build -clean

# --- macOS ---

build-mac-arm64:
	GOOS=darwin GOARCH=arm64 wails build -clean -platform darwin/arm64
	@echo "Built: $(BUILD_DIR)/$(APP_NAME).app (ARM64)"

build-mac-amd64:
	GOOS=darwin GOARCH=amd64 wails build -clean -platform darwin/amd64
	@echo "Built: $(BUILD_DIR)/$(APP_NAME).app (AMD64)"

build-mac-universal:
	wails build -clean -platform darwin/universal
	@echo "Built: $(BUILD_DIR)/$(APP_NAME).app (Universal)"

build-mac: build-mac-universal

# --- Windows ---

build-windows:
	wails build -clean -platform windows/amd64
	@echo "Built: $(BUILD_DIR)/$(APP_NAME).exe (Windows AMD64)"

build-windows-arm64:
	wails build -clean -platform windows/arm64
	@echo "Built: $(BUILD_DIR)/$(APP_NAME).exe (Windows ARM64)"

# --- Linux ---

build-linux:
	wails build -clean -platform linux/amd64
	@echo "Built: $(BUILD_DIR)/$(APP_NAME) (Linux AMD64)"

# --- 全平台构建 ---

build-all: clean
	@echo "=== Building macOS Universal ==="
	wails build -clean -platform darwin/universal
	@mkdir -p dist
	@cp -r "$(BUILD_DIR)/$(APP_NAME).app" "dist/$(APP_NAME)-mac-universal.app"
	@echo ""
	@echo "=== Building Windows AMD64 ==="
	wails build -clean -platform windows/amd64
	@cp "$(BUILD_DIR)/$(APP_NAME).exe" "dist/$(APP_NAME)-windows-amd64.exe"
	@echo ""
	@echo "=== Building Linux AMD64 ==="
	wails build -clean -platform linux/amd64
	@cp "$(BUILD_DIR)/$(APP_NAME)" "dist/$(APP_NAME)-linux-amd64"
	@echo ""
	@echo "=== Build Complete ==="
	@ls -lh dist/

# --- 工具 ---

install-deps:
	cd frontend && npm install

clean:
	rm -rf $(BUILD_DIR)/*
	rm -rf dist/
	@echo "Cleaned."

generate:
	wails generate module
