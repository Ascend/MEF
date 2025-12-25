#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
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

init_output_dir() {
    rm -rf "${OUTPUT_DIR}"
    if ! mkdir -p "${OUTPUT_DIR}"; then
        echo "Create output dir ${OUTPUT_DIR} failed"
        exit 1
    fi
}

# download dependencies
download_dependencies() {
    local dependency_script="${TOP_DIR}/src/om/platform/MindXOM_SDK/build/download_dependencies.sh"

    if [[ ! -f "${dependency_script}" ]]; then
        echo "Script ${dependency_script} does not exist"
        return 1
    fi

    if ! bash "${dependency_script}"; then
        echo "Execute script ${dependency_script} failed"
        return 1
    fi
}

run_build() {
    local component_dir="$1"
    local build_dir="${component_dir}/build"

    pushd "${build_dir}" || return 1

    if ! dos2unix *.sh && chmod +x *.sh; then
        echo "Set permission for scripts in ${build_dir} failed"
        return 1
    fi

    if ! bash build.sh; then
        echo "execute build script failed"
        return 1
    fi

    popd || return 1

    # copy output packages
    cp "${component_dir}"/output/*.zip "${OUTPUT_DIR}"
}

main() {
    init_output_dir

    download_dependencies || exit 1

    echo "Start build component: OM SDK..."
    local omsdk_dir="${TOP_DIR}/src/om/platform/MindXOM_SDK"
    run_build "${omsdk_dir}" || exit 1

    echo "Start build component: OM..."
    local om_dir="${TOP_DIR}/src/om"
    run_build "${om_dir}" || exit 1

    echo "Build process completed successfully!"
    echo "Output files are located in: ${OUTPUT_DIR}"
}

main