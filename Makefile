# Ensure that 'all' is the default target otherwise it will be the first target from Makefile.common.
all::

# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64 arm64
DOCKER_REPO  ?= richih

include Makefile.common

DOCKER_IMAGE_NAME ?= modbus-exporter

EMBEDMD_BIN:=$(FIRST_GOPATH)/bin/embedmd

build: common-build README.md

README.md: help.txt $(EMBEDMD_BIN)
	$(EMBEDMD_BIN) -w README.md
	rm help.txt

help.txt: modbus_exporter
	./modbus_exporter --help 2> help.txt || true

$(EMBEDMD_BIN):
	@go install github.com/campoy/embedmd@latest
