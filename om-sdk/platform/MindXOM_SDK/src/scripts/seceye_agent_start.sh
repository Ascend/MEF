#!/bin/bash

# get current path
CURRENT_PATH=$(cd "$(dirname "$0")" || exit 3; pwd)

SECEYE_WORK_DIR="${OM_WORK_DIR}/software/sec_agent"
SECEYE_AGENT_FILE="${SECEYE_WORK_DIR}/seceye-agent"
SECEYE_AGENT_CONFIG_FILE="${OM_WORK_DIR}/software/ibma/config/SecAgent.json"
SECEYE_AGENT_MIN_CPU=1
SECEYE_AGENT_MAX_CPU=20
SECEYE_AGENT_DEFAULT_CPU=5
SECEYE_AGENT_MIN_MEMORY=12000
SECEYE_AGENT_MAX_MEMORY=256000
SECEYE_AGENT_DEFAULT_MEMORY=65000

# Load log module
source "${CURRENT_PATH}"/log_print.sh
# init log
init_log_path "/var/plog/ibma_edge" "seceye_agent_run.log"

if [ ! -f "${SECEYE_AGENT_FILE}" ]; then
    logger_warn "Not found SecEye-Agent service"
    exit 1
fi

if [ ! -x "${SECEYE_AGENT_FILE}" ]; then
    chmod u+x "${SECEYE_AGENT_FILE}"
fi

export LD_LIBRARY_PATH="${OM_WORK_DIR}"/lib/:"${SECEYE_WORK_DIR}":"${LD_LIBRARY_PATH}"

function check_cpu_param()
{
    if [[ -z "$1" ]]; then
        return 1
    fi

    if [[ "$1" =~ ^[0-9]{1,2}$ ]] && [[ "$1" -ge "${SECEYE_AGENT_MIN_CPU}" ]] && [[ "$1" -le "${SECEYE_AGENT_MAX_CPU}" ]]; then
        return 0
    fi
    return 1
}

function check_memory_param()
{
    if [[ -z "$1" ]]; then
        return 1
    fi

    if [[ "$1" =~ ^[0-9]{1,15}$ ]] && [[ "$1" -ge "${SECEYE_AGENT_MIN_MEMORY}" ]] && [[ "$1" -le "${SECEYE_AGENT_MAX_MEMORY}" ]]; then
        return 0
    fi
    return 1
}

function get_seceye_agent_cpu_config()
{
    if [[ ! -f "${SECEYE_AGENT_CONFIG_FILE}" ]]; then
        cpu_config="${SECEYE_AGENT_DEFAULT_CPU}"
    else
        cpu_config=$(< "${SECEYE_AGENT_CONFIG_FILE}" grep "SecAgentCPU" | awk -F ':' '{print $2}' | sed 's/[",]//g')

        if ! check_cpu_param "${cpu_config}"; then
            logger_warn "current cpu config: ${cpu_config}"
            cpu_config="${SECEYE_AGENT_DEFAULT_CPU}"
        fi
    fi
}

function get_seceye_agent_memory_config()
{
    if [[ ! -f "${SECEYE_AGENT_CONFIG_FILE}" ]]; then
        memory_config="${SECEYE_AGENT_DEFAULT_MEMORY}"
    else
        memory_config=$(< "${SECEYE_AGENT_CONFIG_FILE}" grep "SecAgentMemory" | awk -F ':' '{print $2}' | sed 's/[",]//g')

        if ! check_memory_param "${memory_config}"; then
            logger_warn "current memory config: ${memory_config}"
            memory_config="${SECEYE_AGENT_DEFAULT_MEMORY}"
        fi
    fi
}

function monitor_seceye_agent()
{
    while true
    do
        sleep 5
        if [ "$(ps -ef | grep -v grep | grep "${SECEYE_AGENT_FILE}" -c)" -eq 0 ]; then
            cd "${SECEYE_WORK_DIR}" || { logger_error "${SECEYE_WORK_DIR} does not exist"; exit 3; }
            (
                export PATH="${PATH}":/usr/local/scripts
                "${SECEYE_WORK_DIR}"/seceye-agent --mgmt-socket-path "${SECEYE_WORK_DIR}"/server.socket --slot-id 1 --cpu-id 1 --path "${SECEYE_WORK_DIR}" &
            )
        fi

        # 查询进程id
        pid=$(pgrep -f "${SECEYE_AGENT_FILE}")
        if [ -z "${pid}" ]; then
            logger_info "SecEye-Agent service is not running."
            continue
        fi

        # 查询CPU占用率
        cpu_usage=$(ps -aux | grep "${SECEYE_AGENT_FILE}" | grep -v grep | awk '{print $3}')
        if [ -z "${cpu_usage}" ]; then
            logger_warn "Query SecEye-Agent service cpu usage error."
            continue
        fi

        # 查询内存占用
        mem_usage=$(< /proc/"${pid}"/status grep "VmRSS" | awk '{print $2}')
        if [ -z "${mem_usage}" ]; then
            logger_warn "Query SecEye-Agent service memory usage error."
            continue
        fi

        # 获取配置文件中的阈值
        get_seceye_agent_cpu_config
        cpu_threshold="${cpu_config}"
        get_seceye_agent_memory_config
        memory_threshold="${memory_config}"

        # CPU峰值超阈值, 杀掉进程
        ret=$(awk -v num1="${cpu_usage}" -v num2="${cpu_threshold}" 'BEGIN{print(num1<num2)?"0":"1"}')
        if [[ "${ret}" -eq 1 ]]; then
            logger_warn "Over MAX CPU usage."
            kill "${pid}"
        fi

        if [[ ${mem_usage} -gt ${memory_threshold} ]]; then
            logger_warn "Over MAX Memory usage"
            kill "${pid}"
        fi
    done
}

monitor_seceye_agent
