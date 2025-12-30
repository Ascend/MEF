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

# 该文件在A500 A2恢复出厂设置|恢复最小系统的时候，主要作用是用于安装MEFEdge。

CurrentPath=$(dirname "$(readlink -f "$0")")
CurrentDir="$(basename "${CurrentPath}")"
CurrentScriptName=$(basename "$0")
InstallScriptPath=$(dirname "$(dirname "$(dirname "$CurrentPath")")")
InstallScript="install.sh"
InstallPath=$(dirname "$InstallScriptPath")
LogRootDir="/var/alog/MEFEdge_log/"
LogDir="${LogRootDir}/edge_installer/"
ResetLog="${LogDir}"/edge_installer_run.log

declare -i Err_OK=0
declare -i Err_InstallEdgeFailed=8

function check_log_dir()
{
    local dir_path=$1
    local dir_mode=$2

    if [[ -L "${dir_path}" ]]; then
        unlink "${dir_path}"
    elif [[ -f "${dir_path}" ]]; then
        unlink "${dir_path}"
    fi

    if [[ ! -d "${dir_path}" ]]; then
      mkdir -p "${dir_path}"
    fi

    chmod "${dir_mode}" "${dir_path}"
}

function check_log_file()
{
    local log_path=$1
    if [[ -L "${log_path}" ]]; then
        unlink "${log_path}"
    elif [[ -d "${log_path}" ]]; then
        rm -r "${log_path}"
    fi

    if [[ ! -f "${log_path}" ]]; then
        touch "${log_path}"
    fi

    chmod 640 "${log_path}"
}

function logger()
{
    level=$1
    info=$2
    line=$3

    check_log_file "${ResetLog}"
    printf "%-11s%s %-8s%s/%s:%d    %s\n" "[${level}]" "$(date "+%Y/%m/%d %T.%6N")" "1" "${CurrentDir}" \
    "${CurrentScriptName}" "${line}" "${info}" >> "${ResetLog}"
    echo "${info}"
}

function create_log_dir()
{
    check_log_dir "${LogRootDir}" 755
    check_log_dir "${LogDir}" 750
}

function  cleanUnpackPkg() {
    rm -rf "${InstallScriptPath}"
}

function unregister_service()
{
    rm -f /etc/systemd/system/multi-user.target.wants/reset_mefedge.service
    rm -f /lib/systemd/system/reset_mefedge.service
}

function install_mef_edge()
{
    if ! "${InstallScriptPath}"/"${InstallScript}" -install_dir "${InstallPath}";then
      return "${Err_InstallEdgeFailed}"
    fi

    return "${Err_OK}"
}

function main()
{
    create_log_dir
    logger INFO "start reset install" "${LINENO}"

    install_mef_edge
    ret=$?
    cleanUnpackPkg
    unregister_service
    if [ "${ret}" -ne 0 ];then
      logger ERROR "reset install failed:${ret}" "${LINENO}"
      return "${ret}"
    fi

    logger INFO "reset install success" "${LINENO}"
    return "${Err_OK}"
}

main "$@"
RESULT=$?
exit "${RESULT}"
