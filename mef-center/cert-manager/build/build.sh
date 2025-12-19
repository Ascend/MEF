#!/bin/bash
# Perform  build inference
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MEF is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

export GO111MODULE="on"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="7.3.0"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  #cut the chars after ':'
  build_version=${line#*=}
fi

OUTPUT_NAME="cert-manager"
DOCKER_FILE_NAME="Dockerfile"
arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"

function clean() {
  rm -rf "${TOP_DIR}/output"
  mkdir -p "${TOP_DIR}/output"
  cd "${TOP_DIR}" && go mod tidy
}

function build() {
  cd "${TOP_DIR}/cmd"
  export GO111MODULE="on"
  export GONOSUMDB="*"
  export CGO_ENABLED=1
  export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go build -mod=mod -buildmode=pie -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now \
          -X main.BuildName=${OUTPUT_NAME} \
          -X main.BuildVersion=${build_version}_linux-${arch}" \
          -o ${OUTPUT_NAME} \
          -trimpath
  ls ${OUTPUT_NAME}
  if [ $? -ne 0 ]; then
    echo "fail to find ${OUTPUT_NAME}"
    exit 1
  fi
}

function mv_file() {
  mv "${TOP_DIR}/cmd/${OUTPUT_NAME}" "${TOP_DIR}/output"
  cp "${TOP_DIR}/build/${OUTPUT_NAME}".yaml "${TOP_DIR}/output/${OUTPUT_NAME}".yaml
  cp "${TOP_DIR}/build/${DOCKER_FILE_NAME}" "${TOP_DIR}/output"
  mkdir -p "${TOP_DIR}/output/config"
  chmod 700 "${TOP_DIR}/output/config"
  cp -r "${TOP_DIR}/config/"* "${TOP_DIR}/output/config"
  chmod 400 "${TOP_DIR}/output/"*
  chmod 700 "${TOP_DIR}/output/config"
  chmod 400 "${TOP_DIR}/output/config/"*
  chmod 500 "${TOP_DIR}/output/${OUTPUT_NAME}"
  chmod 600 "${TOP_DIR}/output/${OUTPUT_NAME}.yaml"
}

function main() {
  clean
  build
  mv_file
}

main