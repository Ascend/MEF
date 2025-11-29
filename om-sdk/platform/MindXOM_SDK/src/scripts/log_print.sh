#!/bin/bash

# 默认参数
LOG_DIR="/var/plog"
LOG_FILE="/var/plog/upgrade.log"
DATA_FORMAT="+%Y-%m-%d %T"
LOG_MAX_FILE_SIZE=31457280

init_log_path()
{
    local log_name
    # 参获取日志保存路径
    if [[ "$#" == 2 ]]; then
        LOG_DIR="$1"
        log_name="$2"
        LOG_FILE="${LOG_DIR}/${log_name}"
    fi

    # 检测并创建日志
    if [[ -L "${LOG_DIR}" ]]; then
        unlink "${LOG_DIR}"
        mkdir -p "${LOG_DIR}" && chmod 700 "${LOG_DIR}"
    else
        if [[ ! -d "${LOG_DIR}" ]]; then
            mkdir -p "${LOG_DIR}" && chmod 700 "${LOG_DIR}"
        fi
    fi

    if [[ -L "${LOG_FILE}" ]]; then
        unlink "${LOG_FILE}"
        touch "${LOG_FILE}" && chmod 600 "${LOG_FILE}"
    else
        if [[ ! -f "${LOG_FILE}" ]]; then
            touch "${LOG_FILE}" && chmod 600 "${LOG_FILE}"
        fi
    fi
}

color_red=""
color_yellow=""
color_norm=""
if [[ -t 1 ]] && [[ $((1$(tput colors 2> /dev/null))) -ge 18 ]]; then
    color_red="$(tput setaf 1)"
    color_yellow="$(tput setaf 3)"
    color_norm="$(tput sgr0)"
fi

readonly color_red
readonly color_yellow
readonly color_norm

if command -v caller >/dev/null 2>&1; then
    # NOTE: skip 2-level inner frame
    _caller() { caller 2| awk '{sub(/.*\//,e,$3);print $2"("$3":"$1")"}'; }
else
    _caller() { :; }
fi

check_log_file()
{
    local real_name
    real_name=$(readlink -f "${LOG_FILE}")
    if [[ "${LOG_FILE}" != "${real_name}" ]]; then
        echo "log file is softlink!"
        return 1
    fi

    return 0
}

_log()
{
    LEVEL="$1"
    shift 1

    check_log_file
    local ret=$?
    if [[ -f "${LOG_FILE}" && "${ret}" -eq 0 ]]; then
        if [[ $(stat -c %s "${LOG_FILE}") -gt "${LOG_MAX_FILE_SIZE}" ]]; then
            cat "/dev/null" > "${LOG_FILE}"
        fi
        echo "[$(date "${DATA_FORMAT}")] [${LEVEL}] [$(_caller)]: $*" >> "${LOG_FILE}"
    fi
}

logger_debug()
{
    echo "$(date ${DATA_FORMAT}) [${LEVEL}][$(_caller)]: $*"
}

logger_info()
{
    _log INFO "$@"
}

logger_warn()
{
    _log WARN "${color_yellow}$*${color_norm}"
}

logger_error()
{
    _log ERROR "${color_red}$*${color_norm}"
}

#***********************************************
#  Description: 用于记录操作日志
#  Parameter: $1：日志目录， $2：日志文件名，$3：日志信息
#***********************************************
function record_op_log()
{
    local op_dir="$1"
    local op_log="${op_dir}/$2"
    local op_msg="$3"

    # 检查操作日志目录和日志文件
    if [[ ! -d "${op_dir}" ]]; then
        mkdir "${op_dir}" -p
        chmod 700 "${op_dir}"
    fi

    if [ -L "${op_log}" ]; then
        unlink "${op_log}"
    fi

    if [ ! -f "${op_log}" ]; then
        touch "${op_log}" && chmod 640 "${op_log}"
    fi

    local local_ip=""
    local_ip=$(/usr/bin/who am i | cut -d \( -f 2 | cut -d \) -f 1)
    if [[ -z "${local_ip}" ]]; then
        local_ip="LOCAL"
    fi

    local local_name=""
    local_name=$(/usr/bin/whoami)
    if [[ -z "${local_name}" ]]; then
        local_name="root"
    fi

    local cur_date=""
    cur_date=$(date +"%Y-%m-%d %H:%M:%S,%3N")
    echo "[${cur_date}] [${local_name}@${local_ip}] $(echo "${op_msg}" | sed 's/\(\(\/\w\+\)\+\/\)/****\//g')" >> "${op_log}"
}
