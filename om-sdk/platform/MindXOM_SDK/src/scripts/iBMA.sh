#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2015-2023. All rights reserved.
# Description: monitor process start script.

CONF_DIR="/home/data/config"
LOG_SCRIPT="${OM_WORK_DIR}/scripts/log_print.sh"
OS_CMD_CONF="${CONF_DIR}/os_cmd.conf"
MONITOR_SCRIPT="${OM_WORK_DIR}/software/ibma/bin/monitor.py"
SERVICE_NAME="monitor"
INTERPRETER="python3"
RESTART_FLAG="/run/restart_flag"
RESTARTING_FLAG="/run/restarting_flag"

# Description: Start service
# Globals: MONITOR_SCRIPT SERVICE_NAME RESTART_FLAG RESTARTING_FLAG HDD_SN_INI INTERPRETER
# Returns: 0 for success, other nums for fail
function start()
{
    if pgrep -f "${MONITOR_SCRIPT}" > /dev/null; then
        logger_info "${SERVICE_NAME} service is already running."
        return 0
    fi

    if [ ! -s "${MONITOR_SCRIPT}" ] || [ ! -x "${MONITOR_SCRIPT}" ]; then
        logger_error "Check ${MONITOR_SCRIPT} is invalid."
        return 1
    fi

    local system_start_time=""
    # check system restart flag.
    if [ ! -e "${RESTART_FLAG}" ] && [ ! -e "${RESTARTING_FLAG}" ]; then
        logger_info "capture system restart."
        system_start_time=$(date +%s)
        echo "SYSTEM_START_TIME=${system_start_time}" 2> /dev/null > "${RESTARTING_FLAG}"
        chmod 644 ${RESTARTING_FLAG}
    fi

    # set process umask right
    umask 0077
    # start monitor
    local start_log_file="/var/plog/ibma_edge/om_scripts_run.log"
    ${INTERPRETER} -u "${MONITOR_SCRIPT}" 2>&1 | awk '{now=strftime("%Y-%m-%d %H:%M:%S"); print now,$0;system("")}' >> "${start_log_file}" &

    # wait for 30 seconds at most.
    local counter=0
    while [ ${counter} -lt 60 ]; do
        sleep 0.5
        # check if it's still running. if exit status is 0, it's still running.
        if pgrep -f "${MONITOR_SCRIPT}" > /dev/null; then
            return 0
        fi

        (( counter++ )) || true
    done

    # failed, timeout or just exited.
    logger_error "Failed to start ${SERVICE_NAME} service."
    return 1
}

# Description: Stop service
# Globals: MONITOR_SCRIPT SERVICE_NAME
# Returns: 0 for success, other nums for fail
function stop()
{
    local pid=""
    pid=$(pgrep -f "${MONITOR_SCRIPT}")
    if [ -z "${pid}" ]; then
        logger_info "${SERVICE_NAME} service is not running."
        return 0
    fi

    # send SIGTERM signal, and wait.
    if kill -TERM "${pid}" 2>&1 || kill -SIGTERM "${pid}" 2>&1; then
        logger_info "Sent stop signal to ${SERVICE_NAME} process, now wait." "HIDE"
        return 0
    fi

    logger_error "Failed to send stop signal to ${SERVICE_NAME} process, try again later."

    # wait for 60 seconds at most.
    local counter=0
    while [ ${counter} -lt 60 ]; do
        # check if it's still running. no process found, process stopped.
        if ! pgrep -f "${MONITOR_SCRIPT}" > /dev/null; then
            logger_info "${SERVICE_NAME} service stopped successfully."
            return 0
        fi

        (( counter++ )) || true
        sleep 1
    done

    logger_error "Failed to stop ${SERVICE_NAME} service."
    return 1
}

# Description: Restart service
# Globals: MONITOR_SCRIPT SERVICE_NAME
# Returns: 0 for success, other nums for fail
function restart()
{
    if ! stop; then
        logger_error "Failed to restart ${SERVICE_NAME} service."
        return 1
    fi

    if ! start; then
        logger_error "Failed to restart ${SERVICE_NAME} service."
        return 1
    fi

    if ! pgrep -f "${MONITOR_SCRIPT}" > /dev/null; then
        logger_error "Failed to restart ${SERVICE_NAME} service."
        return 1
    fi

    logger_info "Restart ${SERVICE_NAME} service successfully."
    return 0
}

function main()
{
    if [ ! -s "${LOG_SCRIPT}" ]; then
        echo "[ERROR] Check ${LOG_SCRIPT} is invalid."
        return 1
    fi

    # shellcheck disable=SC1090
    source "${LOG_SCRIPT}"
    init_log_path "/var/plog/ibma_edge" "om_scripts_run.log"

    if [ ! -s "${OS_CMD_CONF}" ]; then
        logger_error "Check ${OS_CMD_CONF} is invalid."
        return 1
    fi

    # shellcheck disable=SC1090
    source "${OS_CMD_CONF}"

    # check if this is the root
    if [ "$(${OS_CMD_ID} -u)" != 0 ]; then
        logger_error "$0: This script must be run as root."
        return 1
    fi

    if ! restart; then
        return 1
    fi

    return 0
}

main "$@"
exit $?
