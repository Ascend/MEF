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
ATLAS_EDGE_BASE_DIR=${TOP_DIR}
INSTALL_BUILD_SCRIPT=$(realpath "${ATLAS_EDGE_BASE_DIR}/mef-center-install/build/build.sh")
OPENSOURCE_BUILD_SCRIPT=$(realpath "${ATLAS_EDGE_BASE_DIR}/build/prepare_dependency.sh")

# project directory name
EDGE_MANAGER_DIR_NAME="edge-manager"
CERT_MANAGER_DIR_NAME="cert-manager"
ALARM_MANAGER_DIR_NAME="alarm-manager"
NGINX_MANAGER_DIR_NAME="nginx-manager"
MEF_CENTER_INSTALL_DIR="mef-center-install"
C_SCRIPT=${CUR_DIR}/build_c_package.sh

VER_FILE="${TOP_DIR}"/service_config.ini
if [ -f "$VER_FILE" ]; then
  cp "${VER_FILE}" "${TOP_DIR}/${EDGE_MANAGER_DIR_NAME}/"
  cp "${VER_FILE}" "${TOP_DIR}/${CERT_MANAGER_DIR_NAME}/"
  cp "${VER_FILE}" "${TOP_DIR}/${ALARM_MANAGER_DIR_NAME}/"
  cp "${VER_FILE}" "${TOP_DIR}/${NGINX_MANAGER_DIR_NAME}/"
  cp "${VER_FILE}" "${TOP_DIR}/${MEF_CENTER_INSTALL_DIR}/"
fi

build_version="7.3.0"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  #cut the chars after ':'
  build_version=${line#*=}
fi

arch_type=$(arch 2>&1)

function clean() {
    if [ -d "${TOP_DIR}/output" ]; then
        rm -rf "${TOP_DIR}/output"
    fi
    mkdir -p "${TOP_DIR}/output"
}

function mv_file() {
    mkdir -p "${TOP_DIR}/savedir/edge-manager"
    mkdir -p "${TOP_DIR}/savedir/cert-manager"
    mkdir -p "${TOP_DIR}/savedir/installer"
    mkdir -p "${TOP_DIR}/savedir/nginx-manager"
    mkdir -p "${TOP_DIR}/savedir/alarm-manager"

    chmod 700 "${TOP_DIR}/savedir"

    fakeroot cp -rf "${TOP_DIR}/${EDGE_MANAGER_DIR_NAME}/output/"* "${TOP_DIR}/savedir/edge-manager/"
    fakeroot cp -rf "${TOP_DIR}/${MEF_CENTER_INSTALL_DIR}/output/"* "${TOP_DIR}/savedir/installer/"
    fakeroot cp -rf "${TOP_DIR}/${NGINX_MANAGER_DIR_NAME}/output/"* "${TOP_DIR}/savedir/nginx-manager/"
    fakeroot cp -rf "${TOP_DIR}/${CERT_MANAGER_DIR_NAME}/output/"* "${TOP_DIR}/savedir/cert-manager/"
    fakeroot cp -rf "${TOP_DIR}/${ALARM_MANAGER_DIR_NAME}/output/"* "${TOP_DIR}/savedir/alarm-manager/"

    chmod 400 "${TOP_DIR}/savedir/installer/version.xml"
    chmod 600 "${TOP_DIR}/savedir/installer/config/"*

    bash "${TOP_DIR}/build/chmod_prepare.sh" "${TOP_DIR}/savedir"

    cd "${TOP_DIR}/savedir"
    fakeroot tar -zcf "Ascend-mindxedge-mefcenter_${build_version}_linux-${arch_type}.tar.gz" ./*

    mkdir -p "${TOP_DIR}/output"

    mv "${TOP_DIR}/savedir/Ascend-mindxedge-mefcenter_${build_version}_linux-${arch_type}.tar.gz" "${TOP_DIR}/output/"

    fakeroot rm -rf "${TOP_DIR}/output/lib" "${TOP_DIR}/output/include"

    chmod 400 "${TOP_DIR}/output/"*

    cd "${TOP_DIR}/output"
    zip -r "Ascend-mindxedge-mefcenter_${build_version}_linux-${arch_type}.zip" .

    rm -rf "${TOP_DIR}/savedir"
}

function build_and_zip_component() {
  component_name=$1
  cd "${ATLAS_EDGE_BASE_DIR}/${component_name}/build/"
  dos2unix build.sh
  chmod u+x ./build.sh
  # execute build script
  ./build.sh

}

function build_install_bin() {
    bash ${INSTALL_BUILD_SCRIPT}
}

function build_c_files() {
    dos2unix ${C_SCRIPT}
    bash ${C_SCRIPT}
}


function prepare_dependency() {
    bash ${OPENSOURCE_BUILD_SCRIPT}
}

function main() {
  clean
  prepare_dependency
  build_c_files
  build_and_zip_component ${EDGE_MANAGER_DIR_NAME}
  build_and_zip_component ${CERT_MANAGER_DIR_NAME}
  build_and_zip_component ${ALARM_MANAGER_DIR_NAME}
  build_and_zip_component ${NGINX_MANAGER_DIR_NAME}
  build_install_bin
  mv_file


}

main
