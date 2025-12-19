#!/bin/bash
# Perform  build inference
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
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

OUTPUT_INSTALLER_NAME="MEF-center-installer"
OUTPUT_CONTROLLER_NAME="MEF-center-controller"
OUTPUT_UPGRADE_NAME="MEF-center-upgrade"
INSTALL_SH_NAME="install.sh"

arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"

function clean() {
  rm -rf "${TOP_DIR}/output"
  mkdir -p "${TOP_DIR}/output"
  cd "${TOP_DIR}" && go mod tidy
}

function build_installer() {
  cd "${TOP_DIR}/tools/install"
  export GONOSUMDB="*"
  export CGO_ENABLED=1
  export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go build -x -mod=mod -buildmode=pie -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now \
          -X main.BuildName=${OUTPUT_INSTALLER_NAME} \
          -X main.BuildVersion=${build_version}_linux-${arch}" \
          -o ${OUTPUT_INSTALLER_NAME} \
          -trimpath
  ls ${OUTPUT_INSTALLER_NAME}
  if [ $? -ne 0 ]; then
    echo "fail to find ${OUTPUT_INSTALLER_NAME}"
    exit 1
  fi
}

function build_controller() {
  cd "${TOP_DIR}/tools/control"
  go build -x -mod=mod -buildmode=pie -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now \
          -X main.BuildName=${OUTPUT_CONTROLLER_NAME} \
          -X main.BuildVersion=${build_version}_linux-${arch}" \
          -o ${OUTPUT_CONTROLLER_NAME} \
          -trimpath
  ls ${OUTPUT_CONTROLLER_NAME}
  if [ $? -ne 0 ]; then
    echo "fail to find ${OUTPUT_CONTROLLER_NAME}"
    exit 1
  fi
}

function build_upgrade() {
  cd "${TOP_DIR}/tools/upgrade"
  go build -mod=mod -buildmode=pie -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now \
          -X main.BuildName=${OUTPUT_UPGRADE_NAME} \
          -X main.BuildVersion=${build_version}_linux-${arch}" \
          -o ${OUTPUT_UPGRADE_NAME} \
          -trimpath
  ls ${OUTPUT_UPGRADE_NAME}
  if [ $? -ne 0 ]; then
    echo "fail to find ${OUTPUT_UPGRADE_NAME}"
    exit 1
  fi
}

function mv_file() {
    mkdir -p "${TOP_DIR}/output/bin"
    mv "${TOP_DIR}/tools/install/${OUTPUT_INSTALLER_NAME}" "${TOP_DIR}/output/bin/"
    mv "${TOP_DIR}/scripts/install.sh" "${TOP_DIR}/output/"

    commit_id=$(git --git-dir "$(realpath "${TOP_DIR}"/../..)"/.git rev-parse HEAD)
    sed -i "s/{commit_id}/${commit_id}/g" "${TOP_DIR}/build/version.xml"

    sed -i "s/{version}/${build_version}/g" "${TOP_DIR}/build/version.xml"
    sed -i "s/{arch}/${arch}/g" "${TOP_DIR}/build/version.xml"

    mv "${TOP_DIR}/build/version.xml" "${TOP_DIR}/output/"
    mv "${TOP_DIR}/config" "${TOP_DIR}/output/"
    mv "${TOP_DIR}/scripts" "${TOP_DIR}/output/"
    mkdir -p "${TOP_DIR}/output/lib/kmc-lib"
    mkdir -p "${TOP_DIR}/output/lib/lib"

    chmod 700 "${TOP_DIR}/output/lib"
    chmod 500 "${TOP_DIR}/output/lib/"*
    chmod 500 "${TOP_DIR}/output/${INSTALL_SH_NAME}"
    chmod 700 "${TOP_DIR}/output/bin"
    chmod 500 "${TOP_DIR}/output/bin/${OUTPUT_INSTALLER_NAME}"
    chmod 400 "${TOP_DIR}/output/version.xml"
    chmod 700 "${TOP_DIR}/output/scripts"
    chmod 500 "${TOP_DIR}/output/scripts/"*
    chmod 700 "${TOP_DIR}/output/config"
    chmod 600 "${TOP_DIR}/output/config/"*

    # control
    chmod 500 "${TOP_DIR}/tools/control/${OUTPUT_CONTROLLER_NAME}"
    mv "${TOP_DIR}/tools/control/${OUTPUT_CONTROLLER_NAME}" "${TOP_DIR}/output/bin/"

    # upgrade
    chmod 500 "${TOP_DIR}/tools/upgrade/${OUTPUT_UPGRADE_NAME}"
    mv "${TOP_DIR}/tools/upgrade/${OUTPUT_UPGRADE_NAME}" "${TOP_DIR}/output/bin/"

}
function main() {
  clean
  build_installer
  build_controller
  build_upgrade
  mv_file
}
main