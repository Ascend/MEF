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
readonly MEF_EDGE_DIR=$(readlink -f "$CUR_DIR"/../)
readonly INSTALLER_NAME="edge-installer"
readonly INSTALLER_DIR="${MEF_EDGE_DIR}/${INSTALLER_NAME}"
readonly OUTPUT_DIR="${MEF_EDGE_DIR}/output"

# build configuration variables
build_version="7.3.0"
version_file="${CUR_DIR}"/../../../build/service_config.ini
product=""
arch=$(arch)

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

function usage() {
    echo "Usage: $0 -p <product>"
    echo "Example: $0 -p MEF_Edge or $0 -p MEF_Edge_SDK"
    exit 1
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

function print_build_info() {
    echo "Build Version is ${build_version}"
    echo "Build Architecture is ${arch}"
    echo "Build Product is ${product}"
}

function prepare_dependencies() {
    local script="${MEF_EDGE_DIR}/build/prepare_dependency.sh"
    if ! bash "$script"; then
        echo "Failed to execute script $script"
        exit 1
    fi
}

function run_pre_build_commands() {
    cd "$INSTALLER_DIR" || exit 1
    go mod tidy
}

function build_dependencies() {
    local script="${MEF_EDGE_DIR}/build/build_dependency.sh"
    if ! bash "$script"; then
        echo "Failed to execute script $script"
        exit 1
    fi
}

function build_edge_installer() {
    local script="${INSTALLER_DIR}/build/build_edge_installer.sh"
    cp -rf "${version_file}" "${INSTALLER_DIR}/build/"
    if ! bash "$script" -p "$product"; then
        echo "Failed to execute script $script"
        exit 1
    fi
}

function process_output_files() {
    mkdir -p "$OUTPUT_DIR"
    cp "$INSTALLER_DIR/output"/*.tar.gz "$OUTPUT_DIR/" 2>/dev/null || true
    rm -rf "${OUTPUT_DIR:?}/lib" "${OUTPUT_DIR:?}/include" 2>/dev/null || true
}

function prepare_product_specific_files() {
    if [ "$product" = "MEF_Edge" ]; then
        cp "$INSTALLER_DIR/config/software.xml" "$OUTPUT_DIR/" 2>/dev/null || true
    fi
}

function set_file_permissions() {
    chmod 400 "$OUTPUT_DIR"/*
}

function create_archive() {
    if ! command -v zip &> /dev/null; then
        echo "Warning: zip command not found, skipping archive creation"
        return
    fi

    local pkg_name
    if [ "$product" = "MEF_Edge" ]; then
        pkg_name="Ascend-mefedge_${build_version}_linux-${arch}.zip"
    else
        pkg_name="Ascend-mefedgesdk_${build_version}_linux-${arch}.zip"
    fi

    cd "$OUTPUT_DIR" || exit 1
    zip -r "$pkg_name" ./*
    cd - > /dev/null
}

function main() {
    # initialization
    get_build_version
    parse_and_validate_args "$@"
    print_build_info

    echo "Starting build process for ${product}..."

    # perform the build steps
    prepare_dependencies
    run_pre_build_commands
    build_dependencies
    build_edge_installer

    # process output files
    process_output_files
    prepare_product_specific_files
    set_file_permissions
    create_archive

    echo "Build process completed successfully!"
    echo "Output files are located in: $OUTPUT_DIR"
}

main "$@"