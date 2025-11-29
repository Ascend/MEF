#!/bin/bash
# Perform  build inference
# Copyright (c) 2021. Huawei Technologies Co., Ltd. All rights reserved.
set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

export GO111MODULE="on"

OUTPUT_INSTALLER_NAME="MEF-center-installer"
OUTPUT_CONTROLLER_NAME="MEF-center-controller"
OUTPUT_UPGRADE_NAME="MEF-center-upgrade"
INSTALL_SH_NAME="install.sh"

arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"

function clean() {
  rm -rf "${TOP_DIR}/output"
  mkdir -p "${TOP_DIR}/output"
}

function build_installer() {
  cd "${TOP_DIR}/tools/install"
  export CGO_ENABLED=1
  export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go mod tidy
  go build -mod=mod -buildmode=pie -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now \
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
  export CGO_ENABLED=1
  export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go build -mod=mod -buildmode=pie -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now \
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
  export CGO_ENABLED=1
  export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
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
    mkdir -p "${TOP_DIR}/output/mef_center_tools"
    mkdir -p "${TOP_DIR}/output/mef_center_tools/bin"
    # 编译测试时暂时使用cp
    cp -r "${TOP_DIR}/scripts" "${TOP_DIR}/output/mef_center_tools"
    # 将install.sh放到根目录中
    mv "${TOP_DIR}/output/mef_center_tools/scripts/install.sh" "${TOP_DIR}/output/"

    mv "${TOP_DIR}/tools/install/${OUTPUT_INSTALLER_NAME}" "${TOP_DIR}/output/mef_center_tools/bin"
    mv "${TOP_DIR}/tools/control/${OUTPUT_CONTROLLER_NAME}" "${TOP_DIR}/output/mef_center_tools/bin"
    mv "${TOP_DIR}/tools/upgrade/${OUTPUT_UPGRADE_NAME}" "${TOP_DIR}/output/mef_center_tools/bin"
    chmod 500 "${TOP_DIR}/output/mef_center_tools/${INSTALL_SH_NAME}"
    chmod 500 "${TOP_DIR}/output/mef_center_tools/bin/${OUTPUT_INSTALLER_NAME}"
    chmod 500 "${TOP_DIR}/output/mef_center_tools/bin/${OUTPUT_CONTROLLER_NAME}"
    chmod 500 "${TOP_DIR}/output/mef_center_tools/bin/${OUTPUT_UPGRADE_NAME}"
}
function main() {
  clean
  build_installer
  build_controller
  build_upgrade
  mv_file
}
main