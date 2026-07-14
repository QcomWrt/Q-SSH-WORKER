# Metadata Dinamis
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE  ?= $(shell date -u +'%Y-%m-%d_%H:%M:%S')

# Package Target untuk LDFlags Injection
PKG_VERSION = github.com/QcomWrt/Q-SSH-WORKER/version

# Flags Optimizer Go
LDFLAGS = -s -w \
          -X '$(PKG_VERSION).Version=$(VERSION)' \
          -X '$(PKG_VERSION).Commit=$(COMMIT_HASH)' \
          -X '$(PKG_VERSION).BuildDate=$(BUILD_DATE)'

# Direktori Output
DIST_DIR = release

.PHONY: all clean build-amd64 build-arm64 build-armv7 build-android release

all: clean build-amd64 build-arm64 build-armv7 build-android

clean:
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)

build-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/Q-SSH-WORKER-linux-amd64 main.go

build-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/Q-SSH-WORKER-linux-arm64 main.go

build-armv7:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/Q-SSH-WORKER-linux-armv7 main.go

build-android:
	CGO_ENABLED=0 GOOS=android GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/Q-SSH-WORKER-android-arm64 main.go