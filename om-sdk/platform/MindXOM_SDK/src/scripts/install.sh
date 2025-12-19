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
OP_LOG_DIR="/var/plog/manager"
OP_LOG_FILE="manager_operate.log"
SYS_AREA_REST_CAPACITY="10"  # 安装OM所需的系统分区容量10M
OM_USER_NAME="MindXOM"
OM_USER_ID=1224
OM_SERVICE_LIST=("om-init.service" "ibma-edge-start.service" "platform-app.service" "start-nginx.service")
OM_FILES_LIST=("bin" "config" "lib" "scripts" "software" "tools" "version.xml" "install.sh" "uninstall.sh" "upgrade.sh")
OM_WORK_DIR="/usr/local/mindx/MindXOM"

source "${CUR_DIR}"/scripts/log_print.sh

init_log_path "/var/plog" "upgrade.log"

#***********************************************
#  Description: 添加OM进程非root用户
#    Parameter:
#        Input:
#              $1: 用户名、组名
#              $2：用户id
#              $3: 组id
# Return Value: 0   -- 升级成功
#               其他 -- 升级失败
#      Caution: NA
#***********************************************
function add_user()
{
    local user_name=$1
    local user_id=$2
    local group_id=$3
    local max_id=65535

    # MindXOM用户组存在则安装失败，提示已被占用
    if grep -w "^${user_name}" /etc/group &> /dev/null; then
        logger_error "The group name <MindXOM> has been occupied."
        return 1
    fi

    # MindXOM用户存在则安装失败，提示已被占用
    if grep -w "^${user_name}" /etc/passwd &> /dev/null; then
        logger_error "The user name <MindXOM> has been occupied."
        return 1
    fi

    # MindXOM用户组不存在则创建，创建时，如果ID被占用，则一直顺延到65535，
    while grep -w "${group_id}" /etc/group &> /dev/null; do
        group_id=$(( group_id + 1 ))
        if [[ ${group_id} -gt ${max_id} ]]; then
            logger_error "group id is full, add ${user_name} failed"
            return 1
        fi
        continue
    done

    if ! groupadd -g "${group_id}" "${user_name}" &> /dev/null; then
        logger_error "add group ${user_name} failed"
        return 1
    fi

    # MindXOM用户不存在则创建，创建时，如果ID被占用，则一直顺延到65535，
    while grep -w "${user_id}" /etc/passwd &> /dev/null; do
        user_id=$(( user_id + 1 ))
        if [[ ${user_id} -gt ${max_id} ]]; then
            logger_error "user id is full, add ${user_name} failed"
            return 1
        fi
        continue
    done
    OM_USER_ID="${user_id}"
    if ! useradd -u "${user_id}" -g "${user_name}" "${user_name}" -s "$(which nologin)" &> /dev/null; then
        logger_error "add user ${user_name} failed"
        return 1
    fi
    return 0
}

#*****************************************************************************
# Description: 安装OM前系统环境检查，以下目录均为如果不存在，需要创建并赋指定权限
# Globals: SYS_AREA_REST_CAPACITY
# Returns: 0 for success, 1 for fail
#*****************************************************************************
function sys_env_prepare_check()
{
    # 检查当前操作用户是否为root
    if [[ $(whoami) != "root" ]]; then
        logger_error "Current user is not root, please check."
        return 1
    fi

    # 根目录空间检查（需要创建新用户、记录日志）
    local cur_area=""
    cur_area=$(df -l -B 1M | grep "/dev/root")

    if [[ $(echo "${cur_area}" | awk '{print $NF}') != "/" ]]; then
        logger_error "Query system rest space failed."
        return 1
    fi

    if [[ $(echo "${cur_area}" | awk '{print $(NF-2)}') -lt "${SYS_AREA_REST_CAPACITY}" ]]; then
        logger_error "System rest space is insufficient failed."
        return 1
    fi

    # 判断是否已安装OM
    if [ -f "${OM_WORK_DIR}/version.xml" ]; then
        logger_error "MindXOM has been installed. Please uninstall first."
        return 1
    fi

    # 检查当前操作目录是否为根目录
    if [[ "${CUR_DIR}" == "/" ]]; then
        logger_error "The installation directory is '/'."
        return 1
    fi

    return 0
}

#*****************************************************************************
# Description: 安装OM失败的清理操作
# Returns: NA
#*****************************************************************************
function install_failed_clean_up()
{
    # 停止所有OM服务
    for (( i= 0; i < "${#OM_SERVICE_LIST[@]}"; i++)); do
        systemctl stop "${OM_SERVICE_LIST[i]}" > /dev/null 2>&1
    done

    # 杀掉MindXOM用户残留进程
    pgrep -U "${OM_USER_ID}" | while IFS=" " read -r line
    do
        kill -9 "${line}" > /dev/null 2>&1
    done

    # 清理目录
    local dir_list=("/usr/local/mindx/MindXOMUpgrade" "/usr/local/mindx/MindXOM" "/home/data/config" "/var/plog/redfish")
    for (( i= 0; i < "${#dir_list[@]}"; i++)); do
        if [[ -L "${dir_list[i]}" ]]; then
            continue
        fi
        chattr -i -R "${dir_list[i]}" > /dev/null 2>&1
        rm -rf "${dir_list[i]}" > /dev/null 2>&1
    done

    # 删除MindXOM用户及其家目录
    userdel -r "${OM_USER_NAME}" > /dev/null 2>&1

    # 删除MindXOM用户组
    groupdel "${OM_USER_NAME}" > /dev/null 2>&1
}

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


#***********************************************
#  Description: 安装OM
# Return Value: 0   -- 升级成功  1 -- 升级失败
#***********************************************
function install_om()
{
    # 添加 MindXOM 用户
    if ! add_user "${OM_USER_NAME}" "${OM_USER_ID}" "${OM_USER_ID}"; then
        return 1
    fi

    # 执行升级
    (
        export LD_LIBRARY_PATH="${CUR_DIR}"/lib:"${LD_LIBRARY_PATH}"
        export PYTHONPATH="${CUR_DIR}"/software/ibma:"${CUR_DIR}"/software/ibma/opensource/python:"${CUR_DIR}"/scripts/python
        python3 -u "${CUR_DIR}"/scripts/python/install.py
    )
    return "$?"
}

function main()
{
    if ! sys_env_prepare_check; then
        return 1
    fi

    # 添加 MindXOM 用户
    if ! install_om; then
        install_failed_clean_up
        return 1
    fi

}

record_op_log "${OP_LOG_DIR}" "${OP_LOG_FILE}" "Install MindXOM executing."
if ! main; then
    echo "Installation failed, please check </var/plog/upgrade.log> to see specific reason."
    record_op_log "${OP_LOG_DIR}" "${OP_LOG_FILE}" "Install MindXOM failed."
    clean_up_install_dir
    exit 1
fi
record_op_log "${OP_LOG_DIR}" "${OP_LOG_FILE}" "Install MindXOM success."
clean_up_install_dir
echo "Install MindXOM success, MindXOM service is ready."