#!/bin/bash

OM_ALG_JSON="/home/data/config/default/om_alg.json"
OP_LOG_DIR="/var/plog/manager"
OP_LOG_FILE="${OP_LOG_DIR}/manager_operate.log"

function record_op_log()
{
    local op_msg="$1"

    # 检查操作日志目录和日志文件
    if [[ ! -d "${OP_LOG_DIR}" ]]; then
        mkdir "${OP_LOG_DIR}" -p
        chmod 700 "${OP_LOG_DIR}"
    fi

    if [ -L "${OP_LOG_FILE}" ]; then
        unlink "${OP_LOG_FILE}"
    fi

    if [ ! -f "${OP_LOG_FILE}" ]; then
        touch "${OP_LOG_FILE}" && chmod 640 "${OP_LOG_FILE}"
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
    echo "[${cur_date}] [${local_name}@${local_ip}] $(echo "${op_msg}" | sed 's/\(\(\/\w\+\)\+\/\)/****\//g')" >> "${OP_LOG_FILE}"
}

function main() {
    # 解析命令行参数
    while getopts "s:h:" opt; do
        case $opt in
            s)
              local sdp_alg_id=$OPTARG
              ;;
            h)
              local hmac_alg_id=$OPTARG
              ;;
            # 当 $opt 的值既不是 s 也不是 h 时，打印错误信息并退出脚本。
            \?)
              return 1
              ;;
            # 当 $opt 的值为 s 或 h 时，但没有提供相应的参数值时，打印错误信息并退出脚本。
            :)
              return 1
              ;;
          esac
    done

    # 检查当前操作用户是否为root
    if [[ $(whoami) != "root" ]]; then
        echo "Current user is not root, please check."
        return 1
    fi

    # 检查目标文件是否存在
    if [[ ! -f "${OM_ALG_JSON}" ]]; then
        echo "File not found: ${OM_ALG_JSON}" >&2
        return 1
    fi

    # 检查参数是否合法
    if [[ "${sdp_alg_id}" != "8" && "${sdp_alg_id}" != "9" ]]; then
        echo "Invalid sdp_alg_id: $sdp_alg_id" >&2
        return 1
    fi
    if [[ "${hmac_alg_id}" != "2053" && "${hmac_alg_id}" != "2054" ]]; then
        echo "Invalid hmac_alg_id: $hmac_alg_id" >&2
        return 1
    fi

    # 修改配置文件
    if !  su - nobody -s "$(which bash)" -c "sed -i.bak 's/\"sdp_alg_id\": \".*\"/\"sdp_alg_id\": \"${sdp_alg_id}\"/' ${OM_ALG_JSON}" &> /dev/null; then
        su - nobody -s "$(which bash)" -c "mv ${OM_ALG_JSON}.bak ${OM_ALG_JSON}" &> /dev/null
        echo "Config update failed"
        return 1
    fi
    if ! su - nobody -s "$(which bash)" -c "sed -i.bak 's/\"hmac_alg_id\": \".*\"/\"hmac_alg_id\": \"${hmac_alg_id}\"/' ${OM_ALG_JSON}" &> /dev/null; then
        su - nobody -s "$(which bash)" -c "mv ${OM_ALG_JSON}.bak ${OM_ALG_JSON}" &> /dev/null
        echo "Config update failed"
        return 1
    fi
    su - nobody -s "$(which bash)" -c "rm -f ${OM_ALG_JSON}.bak" &> /dev/null
    echo "Config updated success."
    return 0
}

if main "$@"; then
    record_op_log "Update om_alg.json success"
    exit 0
else
    record_op_log "Update om_alg.json failed"
    exit 1
fi

