#!/bin/bash
source /home/data/config/mindx_om_env.conf
source /home/data/config/os_cmd.conf
SOFTWARE_PATH="${OM_WORK_DIR}"/software/ibma/
MONITOR_UPDATE_KEY_PATH="${OM_WORK_DIR}"/software/ibma/monitor_kmc/kmc_updater.py
REDFISH_UPDATE_KEY_PATH="${OM_WORK_DIR}"/software/RedfishServer/redfish_kmc/kmc_updater.py
LOG_FILE_DIR="/var/plog/manager"
RUN_LOG_FILE="${LOG_FILE_DIR}/manager_run.log"
OP_LOG_FILE="${LOG_FILE_DIR}/manager_operate.log"

function op_log() {
    local local_ip
    local_ip=$(${OS_CMD_WHO} am i | cut -d \( -f 2 | cut -d \) -f 1)
    if [[ -z "${local_ip}" ]]; then
        local_ip="LOCAL"
    fi

    local local_name
    local_name=$(${OS_CMD_WHO} am i | awk '{print $1}' | awk '{gsub(/^\s+|\s+$/, "");print}')
    if [[ -z "${local_name}" ]]; then
        local_name=${LOGNAME}
    fi

    # 检查操作日志目录和日志文件
    if [[ ! -d "${LOG_FILE_DIR}" ]]; then
        mkdir "${LOG_FILE_DIR}" -p
        chmod "${LOG_FILE_DIR}" 700
    fi

    if [ -L "${OP_LOG_FILE}" ]; then
        unlink "${OP_LOG_FILE}"
    fi

    if [ ! -f "${OP_LOG_FILE}" ]; then
        touch "${OP_LOG_FILE}" && chmod 640 "${OP_LOG_FILE}"
    fi

    if [ -L "${RUN_LOG_FILE}" ]; then
        unlink "${RUN_LOG_FILE}"
    fi

    if [ ! -f "${RUN_LOG_FILE}" ]; then
        touch "${RUN_LOG_FILE}" && chmod 640 "${RUN_LOG_FILE}"
    fi

    cur_date=$(date +"%Y-%m-%d %H:%M:%S,%3N")
    temp=$(echo "$@" | sed 's/\(\(\/\w\+\)\+\/\)/****\//g')
    echo "${cur_date} [${local_name}@${local_ip}] ${temp}" >> ${OP_LOG_FILE}
}

function main() {
    # 执行monitor的密钥更新
    (
        export LD_LIBRARY_PATH="${OM_WORK_DIR}"/lib:${LD_LIBRARY_PATH}
        export PYTHONPATH="${SOFTWARE_PATH}":"${SOFTWARE_PATH}"/opensource/python:"${PYTHONPATH}"
        python3 -u "${MONITOR_UPDATE_KEY_PATH}" > /dev/null 2>&1
    )
    ret=$?
    if [ ${ret} -eq 0 ]; then
        op_log "Manual update monitor keys successfully."
        echo "Manual update monitor keys successfully."
    else
        op_log "Manual update monitor keys failed."
        echo "Manual update monitor keys failed."
    fi

    # 执行Redfish的密钥更新
    (
        su - MindXOM -s /bin/bash -c "LD_LIBRARY_PATH=${OM_WORK_DIR}/software/RedfishServer/lib/c PYTHONPATH=${OM_WORK_DIR}/software/RedfishServer/lib/python:${OM_WORK_DIR}/software/RedfishServer python3 -u ${REDFISH_UPDATE_KEY_PATH}"
    )
    ret=$?
    if [ ${ret} -eq 0 ]; then
        op_log "Manual update redfish keys successfully."
        echo "Manual update redfish keys successfully."
    else
        op_log "Manual update redfish keys failed."
        echo "Manual update redfish keys failed."
    fi

    return ${ret}
}

main