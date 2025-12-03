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
ATLAS_EDGE_BASE_DIR=${TOP_DIR}
INSTALL_BUILD_SCRIPT=$(realpath "${ATLAS_EDGE_BASE_DIR}/mef-center-install/build/build.sh")

# project directory name
EDGE_MANAGER_DIR_NAME="edge-manager"
CERT_MANAGER_DIR_NAME="cert-manager"
C_SCRIPT=${CUR_DIR}/build_c_package.sh

VER_FILE="${TOP_DIR}"/service_config.ini
if [ -f "$VER_FILE" ]; then
  cp "${VER_FILE}" "${TOP_DIR}/${EDGE_MANAGER_DIR_NAME}/"
  cp "${VER_FILE}" "${TOP_DIR}/${CERT_MANAGER_DIR_NAME}/"
fi

build_version="v2.0.4"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '6p' "$VER_FILE" 2>&1)
  #cut the chars after ':'
  build_version=${line#*:}
fi

arch_type=$(arch 2>&1)

function clean() {
    if [ -d "${TOP_DIR}/output" ]; then
        rm -rf "${TOP_DIR}/output"
    fi
    mkdir -p "${TOP_DIR}/output"
}

function build_and_zip_component() {
  component_name=$1
  cd "${ATLAS_EDGE_BASE_DIR}/${component_name}/build/"
  dos2unix build.sh
  chmod u+x ./build.sh
  # execute build script
  ./build.sh

  cd ../
  if [ "${component_name}" == "${TASK_MANAGER_DIR_NAME}" ]
  then
    folder="Ascend-mindxdl-${component_name}-inner_${build_version}_linux-${arch_type}"
  else
    folder="Ascend-mindxdl-${component_name}_${build_version}_linux-${arch_type}"
  fi

  if [ -d "${folder}" ]; then
    rm -rf "${folder}"
  fi
  cd "${ATLAS_EDGE_BASE_DIR}/${component_name}/"
  cp -rf output/ "${folder}"
  if [ -d "${ATLAS_EDGE_BASE_DIR}/output/${folder}" ];then
    rm -rf "${ATLAS_EDGE_BASE_DIR}/output/${folder}"
  fi
  mv "${folder}" "${ATLAS_EDGE_BASE_DIR}/output/"
}

function build_install_bin() {
    bash ${INSTALL_BUILD_SCRIPT}
    cp -r ${ATLAS_EDGE_BASE_DIR}/mef-center-install/output/* ${ATLAS_EDGE_BASE_DIR}/output/
}

function build_c_files() {
    dos2unix ${C_SCRIPT}
    bash ${C_SCRIPT}
}

function main() {
  clean
  build_c_files
  build_and_zip_component ${EDGE_MANAGER_DIR_NAME}
  build_and_zip_component ${CERT_MANAGER_DIR_NAME}
  build_install_bin
}

main
