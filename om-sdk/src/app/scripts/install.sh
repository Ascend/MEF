#!/bin/bash
# 版权所有 (c) 华为技术有限公司 2022-2023
# 功能描述: 安装OM

CUR_DIR=$(dirname "$(readlink -f "$0")")
SCRIPTS_PATH="${CUR_DIR}/scripts"
OM_UNPACK_PATH="${CUR_DIR}/om_install"
OM_FILES_LIST=("om-sdk.tar.gz" "A500-A2-om.tar.gz" "install.sh" "scripts" "om_install" "version.xml")

#*****************************************************************************
# Description: 清理安装目录
# Returns: NA
#*****************************************************************************
function clean_up_install_dir()
{
    # 清理安装目录
    for (( i= 0; i < "${#OM_FILES_LIST[@]}"; i++)); do
        if [[ -L "${CUR_DIR}/${OM_FILES_LIST[i]}" ]]; then
            continue
        fi

        if [[ "${CUR_DIR}" != "/" ]]; then
            rm -rf "${CUR_DIR}/${OM_FILES_LIST[i]}" > /dev/null 2>&1
        fi
    done
}

function main()
{
    # omsdk软件包
    local omsdk_package_path="$(find "${CUR_DIR}" -type f -name "om-sdk.tar.gz")"
    # om扩展包
    local om_extend_path="$(find "${CUR_DIR}" -type f -name "A500-A2-om.tar.gz")"

    # 将omsdk构建成om软件包
    if ! bash "${SCRIPTS_PATH}/integrate_om_files.sh" "${OM_UNPACK_PATH}" "${omsdk_package_path}" "${om_extend_path}" "${CUR_DIR}/version.xml"; then
        echo "build om package failed"
        return 1
    fi

    # 执行om安装
    local ret
    bash "${OM_UNPACK_PATH}/install.sh"; ret="$?"
    clean_up_install_dir
    return "${ret}"
}

main