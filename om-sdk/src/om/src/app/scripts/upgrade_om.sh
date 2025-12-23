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
# 功能描述: 安装OM

CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(dirname "${CUR_DIR}")
OM_FILES_LIST=("om-sdk.tar.gz" "A500-A2-om.tar.gz" "install.sh")

function clean_up_install_dir()
{
    for (( i= 0; i < "${#OM_FILES_LIST[@]}"; i++)); do
        if [[ -L "${TOP_DIR}/${OM_FILES_LIST[i]}" ]]; then
            continue
        fi
        rm -rf "${TOP_DIR}/${OM_FILES_LIST[i]}" > /dev/null 2>&1
    done
}

function main()
{
    # omsdk软件包
    local omsdk_package_path="$(find "${TOP_DIR}" -type f -name "om-sdk.tar.gz")"
    # om扩展包
    local om_extend_path="$(find "${TOP_DIR}" -type f -name "A500-A2-om.tar.gz")"

    # 整合A500-A2包和om-sdk包代码
    if ! bash "${CUR_DIR}/integrate_om_files.sh" "${TOP_DIR}" "${omsdk_package_path}" "${om_extend_path}" "${TOP_DIR}/version.xml"; then
        return 1
    fi

    # 为了兼容rc1版本，将cann的软件包解压到/run/upgrade目录下
    local cann_package=$(find "/run/upgrade" -maxdepth 1 -name "Ascend-cann-nnrt_*.tar.gz")
    if [[ -f "${cann_package}" ]]; then
        local tmp_cann_path="/run/upgrade/cann"
        rm -rf "${tmp_cann_path}"
        mkdir -p "${tmp_cann_path}"
        tar --no-same-owner -zxf "${cann_package}" -C "${tmp_cann_path}" > /dev/null 2>&1;
    fi

    # 执行升级
    export LD_LIBRARY_PATH="${TOP_DIR}"/lib:"${LD_LIBRARY_PATH}"
    export PYTHONPATH="${TOP_DIR}"/software/ibma:"${TOP_DIR}"/software/ibma/opensource/python:"${TOP_DIR}"/scripts/python
    local ret
    python3 -u "${TOP_DIR}"/scripts/python/upgrade_om.py; ret="$?"
    # 清理多余文件
    clean_up_install_dir
    return "${ret}"
}

main