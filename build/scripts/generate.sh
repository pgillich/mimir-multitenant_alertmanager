#!/bin/bash

[ -n "${DEBUG_SCRIPTS}" ] && set -x

set -euo pipefail

cd "${SRC_DIR}"
mkdir -p "${BUILD_TARGET_DIR}"

go generate \
    -x \
    "${GO_BUILD_FLAGS}" \
    ./...
