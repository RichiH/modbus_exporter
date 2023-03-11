GO           ?= go
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOLANG_CI_BIN := $(FIRST_GOPATH)/bin/golangci-lint
GOLANGCI_LINT_OPTS ?=--enable-all --new-from-rev=HEAD~
GOLANGCI_LINT_VERSION ?= v1.51.2
pkgs          = ./...
EMBEDMD_BIN:=$(FIRST_GOPATH)/bin/embedmd


.PHONY: all
all: build test lint

.PHONY: build
build: modbus_exporter README.md

modbus_exporter: $(shell find -name "*.go" | grep -v vendor)
	$(GO) build

README.md: help.txt $(EMBEDMD_BIN)
	$(EMBEDMD_BIN) -w README.md
	rm help.txt

help.txt: modbus_exporter
	./modbus_exporter --help 2> help.txt || true

.PHONY: lint
lint: $(GOLANG_CI_BIN)
	$(GO) list -e -compiled -test=true -export=false -deps=true -find=false -tags= -- ./... > /dev/null
	$(GOLANG_CI_BIN) run $(GOLANGCI_LINT_OPTS) $(pkgs)

.PHONY: test
test:
	go test ./...

# Binaries

$(EMBEDMD_BIN):
	@go install github.com/campoy/embedmd

$(GOLANG_CI_BIN):
	mkdir -p $(FIRST_GOPATH)/bin
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(FIRST_GOPATH)/bin $(GOLANGCI_LINT_VERSION)
