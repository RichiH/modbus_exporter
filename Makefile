# Copyright 2023 Richard Hartmann
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Ensure that 'all' is the default target otherwise it will be the first target from Makefile.common.
all::

# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64 armv7 arm64 ppc64le s390x
DOCKER_REPO  ?= richih

include Makefile.common

DOCKER_IMAGE_NAME ?= modbus-exporter

EMBEDMD_BIN:=$(FIRST_GOPATH)/bin/embedmd

build: common-build README.md

README.md: help.txt $(EMBEDMD_BIN)
	$(EMBEDMD_BIN) -w README.md
	rm help.txt

help.txt: common-build
	./modbus_exporter --help 2> help.txt || true

$(EMBEDMD_BIN):
	$(GO) install github.com/campoy/embedmd@latest
