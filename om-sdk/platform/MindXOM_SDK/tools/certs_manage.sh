#!/bin/bash
#
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
#

source /home/data/config/mindx_om_env.conf
source /home/data/config/os_cmd.conf
source "${OM_WORK_DIR}"/scripts/log_print.sh
source "${OM_WORK_DIR}"/scripts/comm_checker.sh
LOG_FILE_DIR="/var/plog/manager"
OP_LOG_FILE="${LOG_FILE_DIR}/manager_operate.log"

support_component=("fd-ccae" "web")
support_actions=("getunusedcert" "restorecert" "deletecert")

function op_log() {
    local local_ip
    local_ip=$(${OS_CMD_WHO} am i | cut -d \( -f 2 | cut -d \) -f 1)
    if [[ -z "${local_ip}" ]]; then
        local_ip="LOCAL"
    fi

    local local_name
    local_name=$(${OS_CMD_WHO} am i | awk '{print $1}' | awk '{gsub(/^\s+|\s+$/, "");print}')
    if [[ -z "${local_name}" ]]; then
        local_name="root"
    fi

    if [[ ! -d "${LOG_FILE_DIR}" ]]; then
        mkdir "${LOG_FILE_DIR}" -p
        chmod "${LOG_FILE_DIR}" 700
    fi

    if [[ -L "${OP_LOG_FILE}" ]]; then
        unlink "${OP_LOG_FILE}"
    fi

    if [[ ! -f "${OP_LOG_FILE}" ]]; then
        touch "${OP_LOG_FILE}" && chmod 640 "${OP_LOG_FILE}"
    fi

    cur_date=$(date +"%Y-%m-%d %H:%M:%S,%3N")
    temp=$(echo "$@" | sed 's/\(\(\/\w\+\)\+\/\)/****\//g')
    echo "[${cur_date}] [${local_name}@${local_ip}] ${temp}" >> "${OP_LOG_FILE}"
}

function check_device_500a2() {
    if npu-smi info -t product -i 0 | grep 'Atlas 500 A2' > /dev/null 2>&1; then
        return 0
    fi
    return 1
}

function check_name() {
    if [[ "$1" == *".."* ]]; then
        return 1
    fi
    if [[ ! "$1" =~ ^[0-9a-zA-Z_.]{4,64}$ ]]; then
        return 1
    fi
}

function conflict_check()
{
    local count=0
    local temp_file=/run/certsmanage.log

    if ! check_soft_link "${temp_file}"; then
        echo "certsmanage.log is soft link, unlink it"
        unlink "${temp_file}"
    fi

    ps -elf > "${temp_file}"
    count=$(< "${temp_file}" grep "$(basename "$0")" -c)
    if [[ "${count}" -ge 2 ]]; then
        return 1
    fi

    rm -f "${temp_file}"
    return 0
}

function param_check() {
    if [[ "$#" -gt 4 ]]; then
        return 1
    fi
    if check_device_500a2; then
        support_component=("fd-ccae" "web" "https")
    fi

    local action_check_flag=1
    local component_check_flag=1
    for item in "${support_actions[@]}"; do
        if [[ "$1" == "$item" ]]; then
            action_check_flag=0
            break
        fi
    done
    for item in "${support_component[@]}"; do
        if [[ "$2" == "$item" ]]; then
            component_check_flag=0
            break
        fi
    done

    if [[ "$action_check_flag" -ne 0 ]] || [[ "$component_check_flag" -ne 0 ]]; then
        return 1
    fi
    if [[ -n "$3" ]] && ! check_name "$3"; then
        echo "Parameter name is invalid."
        return 1
    fi
}

function print_help() {
    local formatted_string="{"
    for item in "${support_component[@]}"; do
        if [[ -n "${formatted_string}" ]] && [[ "${formatted_string: -1}" != "{" ]]; then
            formatted_string+="|"
        fi
        formatted_string+="${item}"
    done
    formatted_string+="}"
    echo "Usage: certs_manage.sh {getunusedcert|restorecert|deletecert} "${formatted_string}" [OPTION]"
    echo "    OPTION: Cert name to be deleted when [certs_manage.sh deletecert fd-ccae] is called."
}

function confirm() {
    echo "The specified certificate will be deleted. Are you sure you want to continue? (yes/no)"
    read -r response
    case "$response" in
        yes)
            return 0
            ;;
        no)
            return 1
            ;;
        *)
            echo "Invalid input. Please enter again.(yes/no)"
            confirm
            ;;
    esac
}

function main() {
    if ! param_check "$@"; then
        print_help
        return 1
    fi

    if ! conflict_check; then
        echo "Someone else is working, please wait."
        return 1
    fi

    local redfish_script="${OM_WORK_DIR}"/software/RedfishServer/certs_manage/redfish_cert_manage.py
    local nginx_script="${OM_WORK_DIR}"/scripts/python/nginx_cert_manage.py

    local action="$1"
    local component="$2"
    if [[ -z "$3" ]]; then
        local name="default"
    else
        local name="$3"
    fi

    if [[ "${action}" == "deletecert" ]]; then
        if [[ "${component}" == "fd-ccae" ]] && [[ -z "$3" ]]; then
            echo "Parameter name is null!"
            print_help
            return 1
        fi
        confirm
        if [[ $? -ne 0 ]];then
            echo "Delete cert canceled!"
            return 1
        fi
    fi

    case "${component}" in
        fd-ccae|https)
            (
                su - MindXOM -s /bin/bash -c "LD_LIBRARY_PATH="${OM_WORK_DIR}"/software/RedfishServer/lib/c PYTHONPATH="${OM_WORK_DIR}"/software/RedfishServer/lib/python:"${OM_WORK_DIR}"/software/RedfishServer python3 -u "${redfish_script}" --action "${action}" --component "${component}" --name "${name}""
            )
            ret=$?
            if [[ "${ret}" -ne 0 ]]; then
                op_log ""${action}" "${component}" failed."
                return 1
            fi
            op_log ""${action}" "${component}" successfully."
            return 0
            ;;
        web)
            (
                export LD_LIBRARY_PATH="${OM_WORK_DIR}"/lib:"${LD_LIBRARY_PATH}"
                export PYTHONPATH="${OM_WORK_DIR}"/software/ibma:"${OM_WORK_DIR}"/software/ibma/opensource/python:"${OM_WORK_DIR}"/scripts/python
                python3 "${nginx_script}" --action "${action}" --component "${component}"
            )
            ret=$?
            if [[ "${ret}" -ne 0 ]]; then
                op_log ""${action}" "${component}" failed."
                return 1
            fi
            op_log ""${action}" "${component}" successfully."
            return 0
            ;;
        *)
            echo "Unsupported action: "${action}""
            print_help
            return 1
            ;;
    esac
}

main "$@"
exit $?