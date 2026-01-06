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

CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
OUTPUT_DIR="${TOP_DIR}/output"
COMPONENTS=()

check_arch() {
    local arch
    arch=$(arch)

    if [[ "${arch}" == "aarch64" ]]; then
        COMPONENTS=("mef-edge" "mef-center")
    elif [[ "${arch}" == "x86_64" ]]; then
        echo "MEF Edge does not support the x86_64 architecture, will not build corresponding software packages"
        COMPONENTS=("mef-center")
    else
        echo "The current architecture ${arch} is not supported, no software packages will be built"
        exit 1
    fi
}

init_output_dir() {
    rm -rf "${OUTPUT_DIR}"
    if ! mkdir -p "${OUTPUT_DIR}"; then
        echo "Create output dir ${OUTPUT_DIR} failed"
        exit 1
    fi
}

build_component() {
    local component=$1
    local component_dir="${TOP_DIR}/src/${component}"
    local build_dir="${component_dir}/build"

    pushd "${build_dir}" || return 1

    if ! dos2unix *.sh && chmod +x *.sh; then
        echo "Set permission for scripts in ${build_dir} failed"
        return 1
    fi

    # prepare dependency
    if ! bash prepare_dependency.sh; then
        echo "Execute prepare_dependency.sh failed"
        return 1
    fi

    if [[ "${component}" == "mef-edge" ]]; then
        bash build.sh -p "MEF_Edge_SDK" || return 1
    else
        bash build.sh || return 1
    fi

    popd || return 1

    # copy output packages
    if ! cp "${component_dir}/output"/*.zip "${OUTPUT_DIR}"; then
        echo "Copy output files failed"
        return 1
    fi
}

main() {
    check_arch
    init_output_dir

    for component in "${COMPONENTS[@]}"; do
        echo "Start build component: ${component}"
        if ! build_component "${component}"; then
            echo "Build component ${component} failed"
            exit 1
        fi
        echo "Build component ${component} completed"
    done

    echo "Build process completed successfully!"
    echo "Output files are located in: ${OUTPUT_DIR}"
}

main