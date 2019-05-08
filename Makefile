GO           ?= go
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOLANGCI_LINT := $(FIRST_GOPATH)/bin/golangci-lint
GOLANGCI_LINT_OPTS ?=--enable-all --new-from-rev=HEAD~
GOLANGCI_LINT_VERSION ?= v1.16.0
pkgs          = ./...


$(GOLANGCI_LINT):
	mkdir -p $(FIRST_GOPATH)/bin
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(FIRST_GOPATH)/bin $(GOLANGCI_LINT_VERSION)

lint: $(GOLANGCI_LINT)
	GO111MODULE=on $(GO) list -e -compiled -test=true -export=false -deps=true -find=false -tags= -- ./... > /dev/null
	GO111MODULE=on $(GOLANGCI_LINT) run $(GOLANGCI_LINT_OPTS) $(pkgs)
