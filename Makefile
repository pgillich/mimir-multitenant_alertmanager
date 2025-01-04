SHELL=/bin/bash

# Includes Go go1.23.3
GOLANGCI_LINT_VERSION ?= v1.62.2
DOCKER_BUILDER_IMAGE ?= golangci-lint:${GOLANGCI_LINT_VERSION}
DOCKER_URL_PATH ?= golangci

DOCKER_SHELLCHECK_VERSION ?= v0.10.0
DOCKER_SHELLCHECK_IMAGE ?= shellcheck-alpine:${DOCKER_SHELLCHECK_VERSION}
DOCKER_SHELLCHECK_PATH ?= koalaman

DOCKER_MDLINT_VERSION ?= v0.15.0
DOCKER_MDLINT_IMAGE ?= markdownlint-cli2:${DOCKER_MDLINT_VERSION}
DOCKER_MDLINT_PATH ?= davidanson

DOCKER_OGEN_VERSION ?= v1.8.1
DOCKER_OGEN_IMAGE ?= ogen:${DOCKER_OGEN_VERSION}
DOCKER_OGEN_PATH ?= ghcr.io/ogen-go

OAPI_CODEGEN_VERSION ?= v2.4.1

API_PACKAGE_NAME ?= multitenant-alerts
APP_NAME ?= ${API_PACKAGE_NAME}

SRC_DIR ?= /build
BUILD_SCRIPTS_DIR ?= ${SRC_DIR}/build/scripts
BUILD_TARGET_DIR ?= ${SRC_DIR}/build/bin
TEST_COVERAGE_DIR ?= ${SRC_DIR}/build/coverage
GO_BUILD_FLAGS ?=
GO_TEST_FLAGS ?=
GO_TEST_EXCLUDES ?= /api
GO_LINT_CONFIG ?= .golangci.yaml
SHELLCHECK_SOURCEPATH ?= ${BUILD_SCRIPTS_DIR}
GO_PATH_SRC ?= build/_go
GO_PATH ?= ${SRC_DIR}/${GO_PATH_SRC}
GO_CACHE_SRC ?= build/_go-cache
GO_CACHE ?= ${SRC_DIR}/${GO_CACHE_SRC}
GO_TMPDIR_SRC ?= build/_go-tmpdir
GO_TMPDIR ?= ${SRC_DIR}/${GO_TMPDIR_SRC}

BUILD_VERSION ?= $(shell git describe --tags --always --dirty || echo "v0.0.0")
BUILD_TIME = $(shell date +%FT%T%z)

DOCKER_APP_IMAGE ?= ${APP_NAME}:${BUILD_VERSION}
DOCKER_APP_PATH ?= pgillich

export DOCKER_BUILDKIT=1

DEBUG_SCRIPTS ?=

DOCKER_RUN_FLAGS ?= --user $$(id -u):$$(id -g) \
	--network host \
	--mount=type=bind,source=$(shell readlink -e ${GO_PATH_SRC}),target=${GO_PATH} \
	--mount=type=bind,source=$(shell readlink -e ${GO_CACHE_SRC}),target=${GO_CACHE} \
	--mount=type=bind,source=$(shell readlink -e ${GO_TMPDIR_SRC}),target=${GO_TMPDIR} \
	-v /etc/group:/etc/group:ro \
	-v /etc/passwd:/etc/passwd:ro \
	-v /etc/shadow:/etc/shadow:ro \
	-v ${HOME}/.cache:${HOME}/.cache \
	-v $(shell pwd):${SRC_DIR} \
	-e HOME=${HOME} \
	-e SRC_DIR=${SRC_DIR} \
	-e BUILD_SCRIPTS_DIR=${BUILD_SCRIPTS_DIR} \
	-e BUILD_TARGET_DIR=${BUILD_TARGET_DIR} \
	-e BUILD_VERSION=${BUILD_VERSION} \
	-e BUILD_TIME=${BUILD_TIME} \
	-e CGO_ENABLED=0 \
	-e GOPATH=${GO_PATH} \
	-e GOCACHE=${GO_CACHE} \
	-e GOTMPDIR=${GO_TMPDIR} \
	-e GO_BUILD_FLAGS=${GO_BUILD_FLAGS} \
	-e GO_TEST_FLAGS=${GO_TEST_FLAGS} \
	-e GO_TEST_EXCLUDES=${GO_TEST_EXCLUDES} \
	-e TEST_COVERAGE_DIR=${TEST_COVERAGE_DIR} \
	-e GO_LINT_CONFIG=${GO_LINT_CONFIG} \
	-e SHELLCHECK_SOURCEPATH=${SHELLCHECK_SOURCEPATH} \
	-e DEBUG_SCRIPTS=${DEBUG_SCRIPTS} \
	-e APP_NAME=${APP_NAME}

DOCKERFILE_APP_DIR ?= build

install-oapi-codegen:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}

build-init:
	mkdir -p ${GO_PATH_SRC} ${GO_CACHE_SRC} ${GO_TMPDIR_SRC}
.PHONY: build-init

generate: build-init
	mkdir -p ${GO_PATH_SRC} ${GO_CACHE_SRC} ${GO_TMPDIR_SRC}
	docker run ${DOCKER_RUN_FLAGS} \
		${DOCKER_URL_PATH}/${DOCKER_BUILDER_IMAGE} \
		bash -c ${BUILD_SCRIPTS_DIR}/generate.sh
.PHONY: generate

build: build-init
	docker run ${DOCKER_RUN_FLAGS} \
		${DOCKER_URL_PATH}/${DOCKER_BUILDER_IMAGE} \
		bash -c ${BUILD_SCRIPTS_DIR}/build.sh
.PHONY: build

build-local:
	go build \
		-v \
		-o "./build/bin/${APP_NAME}"
.PHONY: build-local

clean: build-init
	sudo rm -rf ${GO_PATH_SRC}/* ${GO_CACHE_SRC}/* ${GO_TMPDIR_SRC}/*
.PHONY: clean

tidy: build-init
	docker run ${DOCKER_RUN_FLAGS} \
		${DOCKER_URL_PATH}/${DOCKER_BUILDER_IMAGE} \
		bash -c ${BUILD_SCRIPTS_DIR}/tidy.sh
.PHONY: tidy

goenv: build-init
	docker run ${DOCKER_RUN_FLAGS} \
		${DOCKER_URL_PATH}/${DOCKER_BUILDER_IMAGE} \
		bash -c ${BUILD_SCRIPTS_DIR}/env.sh
.PHONY: goenv

test: build-init
	docker run ${DOCKER_RUN_FLAGS} \
		${DOCKER_URL_PATH}/${DOCKER_BUILDER_IMAGE} \
		bash -c ${BUILD_SCRIPTS_DIR}/test.sh
.PHONY: test

lint:
	docker run ${DOCKER_RUN_FLAGS} \
		${DOCKER_URL_PATH}/${DOCKER_BUILDER_IMAGE} \
		bash -c ${BUILD_SCRIPTS_DIR}/lint.sh
.PHONY: test

shellcheck:
	docker run ${DOCKER_RUN_FLAGS} \
		-e SCRIPTDIR=${BUILD_SCRIPTS_DIR} \
		${DOCKER_SHELLCHECK_PATH}/${DOCKER_SHELLCHECK_IMAGE} \
		${BUILD_SCRIPTS_DIR}/shellcheck.sh
.PHONY: shellcheck

mdlint:
	docker run ${DOCKER_RUN_FLAGS} \
		-w ${SRC_DIR} \
		${DOCKER_MDLINT_PATH}/${DOCKER_MDLINT_IMAGE} \
		"**/*.md" "#node_modules"
.PHONY: mdlint

check: lint test shellcheck mdlint
.PHONY: check

image:
	docker build \
	    --network host \
		--build-arg APP_NAME=${APP_NAME} \
		--tag ${DOCKER_APP_PATH}/${DOCKER_APP_IMAGE} \
		.
.PHONY: image

image-push:
	docker image push \
		${DOCKER_APP_PATH}/${DOCKER_APP_IMAGE}
.PHONY: image-push

image-kind:
	kind load docker-image ${DOCKER_APP_PATH}/${DOCKER_APP_IMAGE} --name demo
	sync && echo 3 | sudo tee /proc/sys/vm/drop_caches
.PHONY: image-kind

include openapi.mk
