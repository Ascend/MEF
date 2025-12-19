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
# Description: redfish process start script

# set process umask right
umask 0077

SERVICE_NAME="RedfishServer"
OM_WORK_DIR="/usr/local/mindx/MindXOM"
REDFISH_PATH="${OM_WORK_DIR}/software/RedfishServer/ibma_redfish_main.py"

# Load log module
source /home/data/config/os_cmd.conf
source "${OM_WORK_DIR}"/software/RedfishServer/log_print.sh

# init log
init_log_path "/var/plog/redfish" "redfish_run.log"

chmod 400 "${OM_WORK_DIR}"/software/RedfishServer/user_manager/config/paw.ini

# start function entry.
function readFromConfig()
{
    INIFILE="$1"
    SECTION="$2"
    ITEM="$3"
    _readIni=$(${OS_CMD_AWK} -F '=' '/\['"${SECTION}"'\]/{a=1}a==1&&$1~/'"${ITEM}"'/{print $2;exit}' "${INIFILE}")
    echo "${_readIni}"
}

function checkConfigParam()
{
    local config_file="${OM_WORK_DIR}/software/RedfishServer/config/iBMA.ini"
    if [ ! -f "${config_file}" ]; then
        logger_error "The '$config_file' file does not exist."
        return 1
    fi

    local http_user
    local http_port
    http_user=$(readFromConfig "${config_file}" "iBMA_System" "iBMA_user")
    http_port=$(readFromConfig "${config_file}" "iBMA_System" "iBMA_http_server_port")

    # user should not be empty.
    if [ -z "${http_user}" ]; then
        logger_error "HTTP server user name is empty."
        return 1
    fi

    # check if http_user exist.
    if ! id "${http_user}" 2>/dev/null | grep "(${http_user})" -q; then
        logger_error "HTTP server user ${http_user} does not exist."
        return 1
    fi

    # http_port should between 1024~65535.
    if [ ! "${http_port}" -ge 1024 ] >& /dev/null || [ ! "${http_port}" -le 65535 ] >& /dev/null; then
        logger_error "HTTP server port ${http_port} is invalid."
        return 1
    fi

    return 0
}

function start()
{
    if pgrep -f "${REDFISH_PATH}" > /dev/null 2>&1; then
        logger_info "${SERVICE_NAME} service is already running"
        return 0
    fi

    (
        export LD_LIBRARY_PATH=/usr/lib64:${OM_WORK_DIR}/software/RedfishServer/lib/c:${LD_LIBRARY_PATH}
        export PYTHONPATH=${OM_WORK_DIR}/software/RedfishServer:${OM_WORK_DIR}/software/RedfishServer/lib/python:${PYTHONPATH}
        python3 -u ${REDFISH_PATH} 2>&1 &
    )

    # Waiting for the service to start up at most 30s
    local counter=0
    while [ ${counter} -lt 6 ]; do
        sleep 5
        # check if it's still running. if exit status is 0, it's still running.
        if pgrep -f "${REDFISH_PATH}" > /dev/null 2>&1; then
            logger_info "Success to start ${SERVICE_NAME} service."
            return 0
        fi

        logger_error "${SERVICE_NAME} service is not running, trying to wait ${counter}."
        (( counter++ )) || true
    done

    logger_info "Failed to start ${SERVICE_NAME} service."
    return 1
}

# stop function entry.
function stop()
{
    if ! pgrep -f "${REDFISH_PATH}" > /dev/null 2>&1; then
        logger_info "${SERVICE_NAME} is not running."
        return 0
    fi

    logger_info "${SERVICE_NAME} is running, will kill it."
    kill -9 "$(pgrep -f "${REDFISH_PATH}")"
    return 0
}

# restart function entry.
function main()
{   
    logger_info "Start redfish server..."

    if ! checkConfigParam; then
        return 1
    fi

    if ! stop; then
        return 1
    fi

    if ! start; then
        return 1
    fi

    logger_info "Start redfish done."
    return 0
}

main
exit $?
