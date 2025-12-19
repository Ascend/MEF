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

# build configuration variables
build_version="7.3.0"
version_file="${CUR_DIR}/service_config.ini"
arch=$(arch)
build_tag=""

# components
COMPONENTS=(
    "edge-main"
    "edge-om"
    "edgectl"
    "innerctl"
    "install"
    "upgrade"
)

function usage() {
    echo "Usage: $0 -t <build_tag>"
    echo "Example: $0 -t MEFEdge_A500 or $0 -t MEF_Edge_SDK"
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
}

function parse_and_validate_args() {
    local opt
    while getopts ":t:" opt; do
        case $opt in
            t) build_tag="$OPTARG" ;;
            \?) echo "Invalid option: -$OPTARG" >&2; usage ;;
            :) echo "Option -$OPTARG requires a parameter" >&2; usage ;;
        esac
    done

    if [ -z "$build_tag" ]; then
        echo "Error: build_tag must be specified" >&2
        usage
    fi
}

function setup_go_env() {
    export GO111MODULE=on
    export CGO_ENABLED=1
    export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    echo "Go environment variables configured"
}

function get_component_dir() {
    local component="$1"
    if [ "$component" = "edge-main" ] || [ "$component" = "edge-om" ]; then
        echo "$EDGE_INSTALLER_DIR/cmd/$component"
    else
        echo "$EDGE_INSTALLER_DIR/tool/$component"
    fi
}

function get_ldflags() {
    local component="$1"
    if [ "$component" = "edgectl" ]; then
        echo "-X \"main.BuildName=$component\" -X \"main.BuildVersion=${build_version}\" -s -linkmode=external -extldflags=-Wl,-z,now"
    else
        echo "-X \"main.BuildName=$component\" -X \"main.BuildVersion=${build_version}_linux-${arch}\" -s -linkmode=external -extldflags=-Wl,-z,now"
    fi
}

function build_component() {
    local component="$1"
    local component_dir
    local ldflags
    component_dir=$(get_component_dir "$component")
    ldflags=$(get_ldflags "$component")

    echo "Building $component in $component_dir..."

    cd "$component_dir" || {
        echo "Error: Failed to enter directory: $component_dir" >&2
        return 1
    }

    go build \
        -buildmode=pie \
        -mod=mod \
        -tags="$build_tag" \
        -ldflags="$ldflags" \
        -o "$component" \
        -trimpath

    if [ $? -eq 0 ]; then
        echo "Successfully built $component"
        return 0
    else
        echo "Error: Failed to build $component" >&2
        return 1
    fi
}

function build_all_components() {
    local failed_components=()

    for component in "${COMPONENTS[@]}"; do
        if ! build_component "$component"; then
            failed_components+=("$component")
        fi
    done

    if [ ${#failed_components[@]} -eq 0 ]; then
        echo "All components built successfully."
        return 0
    else
        echo "Error: Failed to build components: ${failed_components[*]}" >&2
        return 1
    fi
}

function main() {
    get_build_version
    print_build_info
    parse_and_validate_args "$@"
    setup_go_env

    echo "Build Tag is ${build_tag}"
    echo "Starting build process..."

    if ! build_all_components; then
        exit 1
    fi
}

main "$@"