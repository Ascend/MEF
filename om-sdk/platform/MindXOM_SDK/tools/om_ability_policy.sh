#!/bin/bash
#
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
#
source /home/data/config/mindx_om_env.conf
source /home/data/config/os_cmd.conf
OP_LOG_FILE="/var/plog/manager/manager_operate.log"

function op_log() {
    local local_ip
    local_ip=$("${OS_CMD_WHO}" am i | cut -d \( -f 2 | cut -d \) -f 1)
    if [[ -z "${local_ip}" ]]; then
        local_ip="LOCAL"
    fi

    local local_name
    local_name=$("${OS_CMD_WHO}" am i | awk '{print $1}' | awk '{gsub(/^\s+|\s+$/, "");print}')
    if [[ -z "${local_name}" ]]; then
        local_name=${LOGNAME}
    fi

    str="$(date +"%Y-%m-%d %H:%M:%S"): User=${local_name}, Location=${local_ip}, $*"
    logger -p local3.info "${str}"

    #检查操作文件是否存在
    if [ ! -f "${OP_LOG_FILE}" ]; then
        touch "${OP_LOG_FILE}"
    fi

    chmod 640 "${OP_LOG_FILE}"
    cur_date=$(date +"%Y-%m-%d %H:%M:%S,%3N")
    temp=$(echo "$@" | sed 's/\(\(\/\w\+\)\+\/\)/****\//g')
    echo "${cur_date} [${local_name}@${local_ip}] ${temp}." >> ${OP_LOG_FILE}
}

function main() {
    export PYTHONPATH=$PYTHONPATH:${OM_WORK_DIR}/software/ibma:${OM_WORK_DIR}/software/ibma/opensource/python:"${OM_WORK_DIR}"/scripts/python
    python3 "${OM_WORK_DIR}/scripts/python/om_ability_policy.py" "$@"
    ret=$?
    if [ "${ret}" -eq 0 ]; then
        op_log "Modify ability policy file successfully, actions: $*"
    elif [ "${ret}" -eq 1 ]; then
        echo ""
        ret=0
    elif [ "${ret}" -eq 2 ]; then
        op_log "Modify ability policy file failed, actions: $*"
    else
        op_log "Modify ability policy file failed"
    fi

    return "${ret}"
}

main "$@"
exit $?