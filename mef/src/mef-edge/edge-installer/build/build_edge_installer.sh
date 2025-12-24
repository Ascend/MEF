#!/bin/bash
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

# global variable definition
readonly CUR_DIR=$(dirname "$(readlink -f "$0")")
readonly EDGE_INSTALLER_DIR=$(readlink -f "$CUR_DIR"/../)
readonly MEF_EDGE_DIR=$(readlink -f "$EDGE_INSTALLER_DIR"/../)
readonly COMPONENT_DIR=$(readlink -f "$MEF_EDGE_DIR"/../)
readonly GIT_DIR=$(readlink -f "$COMPONENT_DIR"/../../)

# build configuration variables
build_version="7.3.0"
version_file="${CUR_DIR}"/../../../../build/service_config.ini
product=""
arch=$(arch)

# related configuration path
DEVICE_PLUGIN_DIR="${COMPONENT_DIR}/device-plugin"
SRC_INSTALL_SHELL_PATH="${EDGE_INSTALLER_DIR}/script/install.sh"
SRC_SERVICE_DIR="${EDGE_INSTALLER_DIR}/config/service"
SRC_KMC_CONFIG_PATH="${EDGE_INSTALLER_DIR}/config/kmc-config.json"
SRC_POD_CONFIG_PATH="${EDGE_INSTALLER_DIR}/config/pod-config.json"
SRC_CONTAINER_CONFIG_PATH="${EDGE_INSTALLER_DIR}/config/container-config.json"
SOFTWARE_PATH="${EDGE_INSTALLER_DIR}/config/software.xml"

# source path
SRC_DEVICE_SHELL_DIR="${EDGE_INSTALLER_DIR}/script/device-plugin"
SRC_CORE_SHELL_DIR="${EDGE_INSTALLER_DIR}/script/edge-core"
SRC_INSTALLER_SHELL_DIR="${EDGE_INSTALLER_DIR}/script/edge-installer"
SRC_CTL_BIN_PATH="${EDGE_INSTALLER_DIR}/tool/edgectl/edgectl"
SRC_INSTALL_BIN_PATH="${EDGE_INSTALLER_DIR}/tool/install/install"
SRC_MAIN_BIN_PATH="${EDGE_INSTALLER_DIR}/cmd/edge-main/edge-main"
SRC_OM_BIN_PATH="${EDGE_INSTALLER_DIR}/cmd/edge-om/edge-om"
SRC_UPGRADE_BIN_PATH="${EDGE_INSTALLER_DIR}/tool/upgrade/upgrade"
SRC_DEVICE_BIN_PATH="${DEVICE_PLUGIN_DIR}/output/device-plugin"
SRC_CORE_BIN_PATH="${MEF_EDGE_DIR}/output/edgecore"
SRC_RUN_SH="${EDGE_INSTALLER_DIR}/script/run.sh"
SRC_PAUSE_PATH="${EDGE_INSTALLER_DIR}/build/pause/pause.tar.gz"

# destination path
DST_INSTALLER_DIR="${EDGE_INSTALLER_DIR}/output/software/edge_installer"
DST_SFW_DIR="${EDGE_INSTALLER_DIR}/output/software"
DST_MAIN_DIR="${EDGE_INSTALLER_DIR}/output/software/edge_main"
DST_OM_DIR="${EDGE_INSTALLER_DIR}/output/software/edge_om"
DST_CORE_DIR="${EDGE_INSTALLER_DIR}/output/software/edge_core"
DST_DEVICE_DIR="${EDGE_INSTALLER_DIR}/output/software/device_plugin"
DST_CONFIG_DIR="${EDGE_INSTALLER_DIR}/output/config"
DST_INSTALLER_BIN_DIR="${DST_INSTALLER_DIR}/bin"
DST_INSTALLER_SCRIPT_DIR="${DST_INSTALLER_DIR}/script"
DST_MAIN_BIN_DIR="${DST_MAIN_DIR}/bin"
DST_OM_BIN_DIR="${DST_OM_DIR}/bin"
DST_CORE_BIN_DIR="${DST_CORE_DIR}/bin"
DST_CORE_SCRIPT_DIR="${DST_CORE_DIR}/script"
DST_DEVICE_SCRIPT_DIR="${DST_DEVICE_DIR}/script"
DST_DEVICE_BIN_DIR="${DST_DEVICE_DIR}/bin"
DST_CONFIG_INSTALLER_DIR="${DST_CONFIG_DIR}/edge_installer"
DST_CONFIG_CORE_DIR="${DST_CONFIG_DIR}/edge_core"
DST_CONFIG_MAIN_DIR="${DST_CONFIG_DIR}/edge_main"
DST_CONFIG_OM_DIR="${DST_CONFIG_DIR}/edge_om"
DST_CORE_CONFIG_DIR="${DST_CONFIG_DIR}/edge_core/"
DST_VERSION_DIR="${EDGE_INSTALLER_DIR}/output/"
DST_SERVICE_DIR="${DST_INSTALLER_DIR}/service"
DST_LIB_DIR="${DST_SFW_DIR}/lib"

# directory permission
DIR_MOD="700"

function usage() {
    echo "Usage: $0 -p <product>"
    echo "Example: $0 -p MEF_Edge or $0 -p MEF_Edge_SDK"
    exit 1
}

function get_build_version() {
    if [ ! -f "$version_file" ]; then
        return
    fi

    local line
    line=$(sed -n '1p' "$version_file" 2>&1)
    if [[ "${line}" != *"="* ]]; then
        echo "Invalid version file format, use default version: ${build_version}" >&2
        return
    fi
    build_version=${line#*=}
}

function print_build_info() {
    echo "Build Version is ${build_version}"
    echo "Build Architecture is ${arch}"
    echo "Build Product is ${product}"
}

function parse_and_validate_args() {
    local opt
    while getopts ":p:" opt; do
        case $opt in
            p) product="$OPTARG" ;;
            \?) echo "Invalid option: -$OPTARG" >&2; usage ;;
            :) echo "Option -$OPTARG requires a parameter" >&2; usage ;;
        esac
    done

    if [ -z "$product" ]; then
        echo "Error: product must be specified" >&2
        usage
    fi
}

function setup_product_specific_vars() {
    case "$product" in
        MEF_Edge)
            SRC_SERVICE_SPECIFIC_DIR="${EDGE_INSTALLER_DIR}/config/service-a500"
            SRC_CORE_CONFIG_PATH="${EDGE_INSTALLER_DIR}/config/edgecore.json"
            SRC_VERSION_PATH="${EDGE_INSTALLER_DIR}/config/version.xml"
            SRC_CAP_CONFIG_PATH="${EDGE_INSTALLER_DIR}/config/capability.json"
            SRC_INSTALLER_A500_SHELL_DIR="${EDGE_INSTALLER_DIR}/script/edge-installer-a500"
            PKG_NAME="Ascend-mefedge_${build_version}_linux-${arch}.tar.gz"
            BUILD_TAG="MEFEdge_A500"
            ;;
        MEF_Edge_SDK)
            SRC_SERVICE_SPECIFIC_DIR="${EDGE_INSTALLER_DIR}/config/service-sdk"
            SRC_CORE_CONFIG_PATH="${EDGE_INSTALLER_DIR}/config/edgecore_sdk.json"
            SRC_VERSION_PATH="${EDGE_INSTALLER_DIR}/config/version_sdk.xml"
            SRC_INSTALLER_CONFIG_PATH="${EDGE_INSTALLER_DIR}/config/serial-number.json"
            SRC_NET_TYPE_PATH="${EDGE_INSTALLER_DIR}/config/net-type.json"
            PKG_NAME="Ascend-mefedgesdk_${build_version}_linux-${arch}.tar.gz"
            BUILD_TAG="MEFEdge_SDK"
            ;;
        *) echo "Error: Invalid product specified: $product" >&2; usage ;;
    esac
}

function build_binaries() {
    local script="${EDGE_INSTALLER_DIR}/build/build_binaries.sh"
    if ! bash "$script" -t "$BUILD_TAG"; then
        echo "Failed to execute script $script" >&2
        exit 1
    fi
}

function build_additional_binaries() {
    rm -rf "${DEVICE_PLUGIN_DIR}"/doc
    bash "${EDGE_INSTALLER_DIR}"/build/pause/build_pause_image.sh
    cp -rf "${version_file}" "${DEVICE_PLUGIN_DIR}/"
    cd "${DEVICE_PLUGIN_DIR}"/build/ && bash build_edge.sh nokmc
    bash "${EDGE_INSTALLER_DIR}"/build/build_core.sh
}

function write_version_info() {
    local commit_id
    commit_id=$(git --git-dir "${GIT_DIR}"/.git rev-parse HEAD)
    sed -i "s/{commit_id}/${commit_id}/g" "${SRC_VERSION_PATH}"
    sed -i "s/{version}/${build_version}/g" "${SRC_VERSION_PATH}"
    sed -i "s/{version}/${build_version}/g" "${SOFTWARE_PATH}"
    sed -i "s/{arch}/${arch}/g" "${SRC_VERSION_PATH}"
    sed -i "s/{arch}/${arch}/g" "${SOFTWARE_PATH}"
}

function create_directories() {
    local dirs=(
        "$DST_INSTALLER_BIN_DIR" "$DST_MAIN_BIN_DIR" "$DST_OM_BIN_DIR" "$DST_CORE_BIN_DIR"
        "$DST_DEVICE_BIN_DIR" "$DST_CONFIG_INSTALLER_DIR" "$DST_CONFIG_CORE_DIR"
        "$DST_CONFIG_MAIN_DIR" "$DST_CONFIG_OM_DIR" "$DST_CORE_SCRIPT_DIR"
        "$DST_DEVICE_SCRIPT_DIR" "$DST_SERVICE_DIR" "$DST_LIB_DIR"
        "$DST_INSTALLER_SCRIPT_DIR"
    )

    for dir in "${dirs[@]}"; do
        mkdir -p "$dir"
    done
}

function copy_common_files() {
    cp "${SRC_INSTALL_SHELL_PATH}" "${EDGE_INSTALLER_DIR}/output"
    cp -r "${SRC_SERVICE_DIR}"/* "${DST_SERVICE_DIR}"
    cp -r "${SRC_SERVICE_SPECIFIC_DIR}"/* "${DST_SERVICE_DIR}"
    cp -r "${SRC_INSTALLER_SHELL_DIR}"/* "${DST_INSTALLER_SCRIPT_DIR}"
    cp -r "${SRC_CORE_SHELL_DIR}"/* "${DST_CORE_SCRIPT_DIR}"
    cp -r "${SRC_DEVICE_SHELL_DIR}"/* "${DST_DEVICE_SCRIPT_DIR}"
    cp "${SRC_CTL_BIN_PATH}" "${DST_INSTALLER_BIN_DIR}"
    cp "${SRC_INSTALL_BIN_PATH}" "${DST_INSTALLER_BIN_DIR}"
    cp "${EDGE_INSTALLER_DIR}"/tool/innerctl/innerctl "${DST_INSTALLER_BIN_DIR}"
    cp "${SRC_MAIN_BIN_PATH}" "${DST_MAIN_BIN_DIR}"
    cp "${SRC_OM_BIN_PATH}" "${DST_OM_BIN_DIR}"
    cp "${SRC_CORE_CONFIG_PATH}" "${DST_CORE_CONFIG_DIR}"
    cp "${SRC_KMC_CONFIG_PATH}" "${DST_CONFIG_MAIN_DIR}"
    cp "${SRC_KMC_CONFIG_PATH}" "${DST_CONFIG_OM_DIR}"
    cp "${SRC_KMC_CONFIG_PATH}" "${DST_CONFIG_CORE_DIR}"
    cp "${SRC_POD_CONFIG_PATH}" "${DST_CONFIG_OM_DIR}"
    cp "${SRC_CONTAINER_CONFIG_PATH}" "${DST_CONFIG_OM_DIR}"
    cp "${SRC_VERSION_PATH}" "${DST_VERSION_DIR}"
    cp "${SRC_UPGRADE_BIN_PATH}" "${DST_INSTALLER_BIN_DIR}"
    cp "${SRC_PAUSE_PATH}" "${DST_CORE_BIN_DIR}"
    cp -d "${MEF_EDGE_DIR}"/output/lib/* "${DST_LIB_DIR}"
    cp "${SRC_RUN_SH}" "${DST_SFW_DIR}"
    mv "${SRC_DEVICE_BIN_PATH}" "${DST_DEVICE_BIN_DIR}"
    mv "${SRC_CORE_BIN_PATH}" "${DST_CORE_BIN_DIR}"
}

function handle_product_specific_files() {
    case "$product" in
        MEF_Edge)
            cp "${SRC_CAP_CONFIG_PATH}" "${DST_CONFIG_MAIN_DIR}"
            cp -r "${SRC_INSTALLER_A500_SHELL_DIR}"/* "${DST_INSTALLER_SCRIPT_DIR}"
            ;;
        MEF_Edge_SDK)
            cp "${SRC_NET_TYPE_PATH}" "${DST_CONFIG_MAIN_DIR}"
            cp "${SRC_INSTALLER_CONFIG_PATH}" "${DST_CONFIG_INSTALLER_DIR}"
            mv "${DST_CORE_CONFIG_DIR}/edgecore_sdk.json" "${DST_CORE_CONFIG_DIR}/edgecore.json"
            mv "${DST_VERSION_DIR}/version_sdk.xml" "${DST_VERSION_DIR}/version.xml"
            ;;
    esac
}

function set_permissions() {
    chmod -R ${DIR_MOD} "${EDGE_INSTALLER_DIR}/output"
    bash "${EDGE_INSTALLER_DIR}"/build/chmod_prepare.sh "${EDGE_INSTALLER_DIR}/output"
    fakeroot chown -R root:root "${EDGE_INSTALLER_DIR}/output/"
}

function create_package() {
    cd "${EDGE_INSTALLER_DIR}/output/" && fakeroot tar -zcvf "${PKG_NAME}" *
}

function main() {
    echo "------------------------ start build edge_installer ------------------------"
    get_build_version
    parse_and_validate_args "$@"
    setup_product_specific_vars
    print_build_info

    echo "Starting build process for ${product}..."

    build_binaries
    build_additional_binaries
    write_version_info
    create_directories
    copy_common_files
    handle_product_specific_files
    set_permissions
    create_package

    echo "Build process completed successfully!"
    echo "Output package: ${EDGE_INSTALLER_DIR}/output/${PKG_NAME}"
    echo "------------------------ end build edge_installer ------------------------"
}

main "$@"