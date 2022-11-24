#!/bin/bash
# Perform  build inference
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
ATLAS_EDGE_BASE_DIR=${TOP_DIR}

# project directory name
EDGE_MANAGER_DIR_NAME="edge-manager"
SOFTWARE_MANAGER_DIR_NAME="software-manager"

VER_FILE="${TOP_DIR}"/service_config.ini
if [ -f "$VER_FILE" ]; then
  cp "${VER_FILE}" "${TOP_DIR}/${EDGE_MANAGER_DIR_NAME}/"
  cp "${VER_FILE}" "${TOP_DIR}/${SOFTWARE_MANAGER_DIR_NAME}/"
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

function main() {
  clean
  build_and_zip_component ${EDGE_MANAGER_DIR_NAME}
  build_and_zip_component ${SOFTWARE_MANAGER_DIR_NAME}
}

main
