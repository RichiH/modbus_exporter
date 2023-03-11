# This file originally copied from prometheus/systemd_exporter:
# https://github.com/prometheus-community/systemd_exporter/blob/9f476c669993db46702116f70ce88dce4d1fd475/Makefile

# Copyright 2022 The Prometheus Authors
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

# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64 armv7 arm64 ppc64le s390x
DOCKER_REPO  ?= RichiH

include Makefile.common

STATICCHECK_IGNORE =

DOCKER_IMAGE_NAME       ?= modbus-exporter
