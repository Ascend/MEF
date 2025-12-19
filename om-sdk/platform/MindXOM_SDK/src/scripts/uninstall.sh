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
# Description: OM uninstall script

OM_USER="MindXOM"
USER_ID=""
OM_WORK_DIR="/usr/local/mindx/MindXOM"
OM_UPGRADE_DIR="/usr/local/mindx/MindXOMUpgrade"
PLOG_LOGROTATE="${OM_WORK_DIR}/config/plog"
SCRIPTS_PY_DIR="${OM_WORK_DIR}/scripts/python"
UNINSTALL_ENTRY="${SCRIPTS_PY_DIR}/uninstall/uninstall_task.py"
MANAGER_LOG_DIR="/var/plog/manager"
MANAGER_LOG_FILE="manager_operate.log"
WEB_EDGE_LOG_DIR="/var/plog/web_edge"
IBMA_EDGE_LOG_DIR="/var/plog/ibma_edge"
REDFISH_LOG_DIR="/var/plog/redfish"
UPGRADE_LOG_FILE="/var/plog/upgrade.log"
NTP_SERVICE_LOG_FILE="/var/plog/ntp_service.log"
OM_SYNC_LOG_DIR="/home/log/om_sync"

source "/home/data/config/mindx_om_env.conf"
source "${OM_WORK_DIR}"/scripts/log_print.sh

init_log_path "/var/plog" "upgrade.log"

# 功能描述: 执行卸载OM前检查，包括OM用户名和属组存在性检查
# 全局变量: OM_USER USER_ID
# 返回值: 0代表成功，其他代表失败
function pre_check()
{
    local group_id
    local min_id=1224
    local max_id=65535
    local pwd_path="/etc/passwd"

    # 检查MindXOM用户是否存在
    if ! USER_ID=$(id -u "${OM_USER}" 2> /dev/null); then
        logger_error "The ${OM_USER} user does not exist. The uninstallation is not allowed."
        return 1
    fi

    # 检查MindXOM用户属组是否存在
    if ! group_id=$(id -g "${OM_USER}" 2> /dev/null); then
        logger_error "The ${OM_USER} user group does not exist. The uninstallation is not allowed."
        return 1
    fi

    # 检查MindXOM用户的id是否符合范围
    if [[ ${USER_ID} -lt ${min_id} || ${USER_ID} -gt ${max_id} ]]; then
        logger_error "The ${OM_USER} user ID is not in the range ${min_id} to ${max_id}. The uninstallation is not allowed."
        return 1
    fi

    # 检查MindXOM用户的是否为nologin权限
    if [[ $(grep -w "^${OM_USER}" ${pwd_path} | awk -F ":" '{print $NF}') != "$(which nologin)" ]]; then
        logger_error "The ${OM_USER} user is not nologin permission. The uninstallation is not allowed."
        return 1
    fi

    # 检查MindXOM用户组的id是否符合范围
    if [[ ${group_id} -lt ${min_id} || ${group_id} -gt ${max_id} ]]; then
        logger_error "The ${OM_USER} group ID is not in the range ${min_id} to ${max_id}. The uninstallation is not allowed."
        return 1
    fi

    return 0
}

# 功能描述: 卸载OM之后，删除OM用户及其属组
# 全局变量: OM_USER USER_ID
function del_user()
{
    # 杀掉MindXOM用户残留进程
    pgrep -U "${USER_ID}" | while IFS=" " read -r line
    do
        kill -9 "${line}" > /dev/null 2>&1
    done

    # 删除MindXOM用户及其家目录
    userdel -r "${OM_USER}" > /dev/null 2>&1
    # 删除MindXOM用户组
    groupdel "${OM_USER}" > /dev/null 2>&1
}

# 功能描述: 卸载OM之后转测其相关日志
# 全局变量: PLOG_LOGROTATE
function log_dump()
{
    # 强制执行plog日志转储
    logrotate -f "${PLOG_LOGROTATE}" > /dev/null 2>&1
}

# 功能描述: 卸载OM之后清理其工作目录和日志目录
# 全局变量: OM_WORK_DIR、MANAGER_LOG_DIR、WEB_EDGE_LOG_DIR、IBMA_EDGE_LOG_DIR、REDFISH_LOG_DIR、UPGRADE_LOG_FILE、NTP_SERVICE_LOG_FILE
function clean_up()
{
    local clear_dirs=(
        "${WEB_EDGE_LOG_DIR}"
        "${IBMA_EDGE_LOG_DIR}"
        "${REDFISH_LOG_DIR}"
        "${MANAGER_LOG_DIR}"
        "${OM_WORK_DIR}"
        "${OM_UPGRADE_DIR}"
        "${OM_SYNC_LOG_DIR}"
    )

    for dir in "${clear_dirs[@]}"; do
        rm -rf "${dir}"
    done

    local clear_files=(
        "${UPGRADE_LOG_FILE}"
        "${NTP_SERVICE_LOG_FILE}"
    )

    for file in "${clear_files[@]}"; do
        rm -f "${file}"
    done
}

# 功能描述: 执行卸载OM
# 返回值: 0代表成功，其他代表失败
function uninstall()
{
    if ! pre_check; then
        return 1
    fi

    (
      export PYTHONPATH="${OM_WORK_DIR}"/software/ibma:"${OM_WORK_DIR}"/software/ibma/opensource/python:"${OM_WORK_DIR}"/scripts/python:${PYTHONPATH}
      python3 "${UNINSTALL_ENTRY}"
    )
    return $?
}

function main()
{
    local ret=0
    record_op_log "${MANAGER_LOG_DIR}" "${MANAGER_LOG_FILE}" "Uninstall MindXOM executing."
    echo -e "Uninstall MindXOM start."

    if uninstall; then
    	record_op_log "${MANAGER_LOG_DIR}" "${MANAGER_LOG_FILE}" "Uninstall MindXOM success."
        echo -e "Uninstall MindXOM success."
        log_dump
        clean_up
        del_user
        ret=0
    else
        record_op_log "${MANAGER_LOG_DIR}" "${MANAGER_LOG_FILE}" "Uninstall MindXOM failed."
        echo -e "Uninstall MindXOM failed, please check </var/plog/upgrade.log> to see specific reason."
        ret=1
    fi

    return ${ret}
}

main "$@"
exit $?
