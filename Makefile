.SILENT :

# App name
APPNAME=scim-ctl

# Main path
MAIN_PATH=./cmd/$(APPNAME)

# Go configuration
GOOS?=$(shell go env GOHOSTOS)
GOARCH?=$(shell go env GOHOSTARCH)

# Add exe extension if windows target
is_windows:=$(filter windows,$(GOOS))
EXT:=$(if $(is_windows),".exe","")

# Archive name
ARCHIVE=$(APPNAME)-$(GOOS)-$(GOARCH).tgz

# Executable name
EXECUTABLE=$(APPNAME)$(EXT)

# Extract version infos
PKG_VERSION:=github.com/ncarlier/$(APPNAME)/internal/version
VERSION:=`git describe --always --tags --dirty`
GIT_COMMIT:=`git rev-list -1 HEAD --abbrev-commit`
BUILT:=`date`
define LDFLAGS
-X '$(PKG_VERSION).Version=$(VERSION)' \
-X '$(PKG_VERSION).GitCommit=$(GIT_COMMIT)' \
-X '$(PKG_VERSION).Built=$(BUILT)' \
-s -w -buildid=
endef

all: build

## This help screen
help:
	printf "Available targets:\n\n"
	awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "%-15s %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
.PHONY: help

## Clean built files
clean:
	-rm -rf release
.PHONY: clean

## Build executable
build:
	-mkdir -p release
	@echo ">>> Building: $(EXECUTABLE) $(VERSION) for $(GOOS)-$(GOARCH) ..."
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -tags osusergo,netgo -ldflags "$(LDFLAGS)" -o release/$(EXECUTABLE) $(MAIN_PATH)
	@echo ">>> Build complete: $(EXECUTABLE) $(VERSION) for $(GOOS)-$(GOARCH) ✅"
.PHONY: build

release/$(EXECUTABLE): build

# Check code style
check-style:
	@echo ">>> Checking code style..."
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...
	@echo ">>> Code style check complete ✅"
.PHONY: check-style

# Check code criticity
check-criticity:
	@echo ">>> Checking code criticity..."
	go run github.com/go-critic/go-critic/cmd/gocritic@latest check -enableAll ./...
	@echo ">>> Code criticity check complete ✅"
.PHONY: check-criticity

# Check code security
check-security:
	@echo ">>> Checking code security..."
	go run github.com/securego/gosec/v2/cmd/gosec@latest -quiet ./...
	@echo ">>> Code security check complete ✅"
.PHONY: check-security

## Code quality checks
checks: check-style check-criticity
.PHONY: checks

## Run tests
test: 
	@echo ">>> Running tests..."
	go test ./...
	@echo ">>> Tests complete ✅"
.PHONY: test

## Install executable
install: release/$(EXECUTABLE)
	@echo ">>> Installing $(EXECUTABLE) to ${HOME}/.local/bin/$(EXECUTABLE) ..."
	cp release/$(EXECUTABLE) ${HOME}/.local/bin/$(EXECUTABLE)
	@echo ">>> Installation complete ✅"
.PHONY: install

# Generate changelog
CHANGELOG.md:
	standard-changelog --first-release

## Create archive
archive: release/$(EXECUTABLE) CHANGELOG.md
	echo ">>> Creating release/$(ARCHIVE) archive..."
	tar czf release/$(ARCHIVE) README.md LICENSE CHANGELOG.md -C release/ $(EXECUTABLE)
	rm release/$(EXECUTABLE)
	@echo ">>> Release/$(ARCHIVE) archive created ✅"
.PHONY: archive

## Create distribution binaries
distribution:
	@echo ">>> Creating distribution binaries..."
	GOARCH=amd64 make build archive
	GOARCH=arm64 make build archive
	GOARCH=arm make build archive
	GOOS=darwin make build archive
	GOOS=windows make build archive
	@echo ">>> Distribution binaries created ✅"
.PHONY: distribution