#!/bin/bash

# set process umask right
umask 0077
OS_CMD_CONF_FILE=/home/data/config/os_cmd.conf
if [ ! -f "${OS_CMD_CONF_FILE}" ]; then
    echo "${OS_CMD_CONF_FILE} not exit or is symbolic link"
    exit 1
else
    source "${OS_CMD_CONF_FILE}"
fi

function check_root_env_path()
{
    if [[ $# -ne 1 ]]; then
        logger_error "parameter error."
        return 1
    fi

    local path="$1"
    local format_name
    local real_name
    local user_id
    local group_id

    # 校验软链接
    format_name=$(${OS_CMD_REALPATH} -ms "${path}")
    real_name=$(${OS_CMD_READLINK} -m "${path}")
    if [[ "${format_name}" != "${real_name}" ]]; then
        logger_error "path is soft link!"
        return 2
    fi

    # 校验属主
    if [[ ! -e "${path}" ]]; then
        return 3
    fi

    user_id=$(${OS_CMD_STAT} -c %u "${path}")
    group_id=$(${OS_CMD_STAT} -c %g "${path}")
    if [[ -z "${user_id}" ]] || [[ -z "${group_id}" ]]; then
        logger_error "user or group not exist"
        return 5
    fi

    if [[ "${user_id}" != "0" ]] && [[ "${group_id}" != "0" ]]; then
        logger_error "The path owner check failed. path=${path}"
        return 6
    fi

    # 校验路径是否包含other的写权限
    local other_write_mod
    other_write_mod=$(${OS_CMD_STAT} -c %A "${path}" | ${OS_CMD_AWK} '{print substr($1,9,1)}')
    if [[ "${other_write_mod}" != '-' ]]; then
        logger_error "other write mod is illegal."
        return 7
    fi

    return 0
}

function check_root_path_and_exit()
{
    if [[ $# -ne 1 ]]; then
        logger_error "parameter error."
        exit 1
    fi

    local path=$1
    check_root_env_path "${path}"
    local ret=$?
    if [[ ${ret} -ne 0 ]]; then
        logger_error "${path} check_root_env_path failed! ${ret}"
        exit ${ret}
    fi
    return 0
}

function prepare_recover_dir() {
    local recover_config_path="/home/package/config"
    if [ ! -d "${recover_config_path}" ]; then
      mkdir -p "${recover_config_path}"
    fi
    chmod 700 "${recover_config_path}"
    safe_change_owner MindXOM MindXOM "${recover_config_path}"
}

# 入口先校验服务环境变量文件是否合法
check_root_path_and_exit "/home/data/config/mindx_om_env.conf"
check_root_path_and_exit "/home/data/config"

# Load log module
if ! check_root_env_path "${OM_WORK_DIR}"/scripts/log_print.sh; then
    echo "${OM_WORK_DIR}/scripts/log_print.sh check failed!"
    exit 1
fi
source "${OM_WORK_DIR}"/scripts/log_print.sh
init_log_path "/var/plog/ibma_edge" "om_scripts_run.log"

if [ -f "${OM_WORK_DIR}/lib/liblpeblock.so" ];then
    check_root_path_and_exit "${OM_WORK_DIR}/lib/liblpeblock.so"
    export LD_PRELOAD="${OM_WORK_DIR}/lib/liblpeblock.so"
fi
check_root_path_and_exit "${OM_WORK_DIR}/scripts/safe_common.sh"
source "${OM_WORK_DIR}"/scripts/safe_common.sh
check_root_path_and_exit "${OM_WORK_DIR}/scripts/comm_checker.sh"
source "${OM_WORK_DIR}"/scripts/comm_checker.sh
UDS_CERT_PATH="${OM_WORK_DIR}"/software/ibma/cert

#add for start IBMA-EDGE
cd "${OM_WORK_DIR}"/software/ibma || exit 3

prepare_recover_dir

# 判断 ibma_edge_service.ini是否存在，并且不为空
file_path=/home/data/ies
if [ ! -s ${file_path}/ibma_edge_service.ini ]; then
    safe_cp "${OM_WORK_DIR}/software/ibma/lib/Linux/config/ibma_edge_service.ini" "${file_path}/ibma_edge_service.ini"
    echo create work file ibma_edge_service.ini!
fi

# 创建shield_alarm.ini文件
shield_alarm_ini_path=/home/data/ies/shield_alarm.ini
if ! check_soft_link "${shield_alarm_ini_path}"; then
    echo "shield_alarm.ini is soft links, unlink it"
    unlink "${shield_alarm_ini_path}"
fi
if [[ ! -f "${shield_alarm_ini_path}" ]]; then
    echo "shield_alarm.ini is not exist, create it"
    touch "${shield_alarm_ini_path}"
fi

ldconfig

# 安全整改涉及项
mkdir -p /run/web && chmod 700 /run/web
safe_change_owner MindXOM MindXOM /run/web

function chmod_ies_config()
{
    # 配置文件
    local INI_ARRAY=("access_control.ini" "certWarnTime.ini" "ibma_edge_service.ini" "mountCnf_site.ini" \
    "NTPEnable.ini" "shield_alarm.ini" "tag.ini")

    for ((i = 0; i < ${#INI_ARRAY[@]}; i++)); do
        if [[ ! -f "/home/data/ies/${INI_ARRAY[i]}" ]]; then
            continue
        fi

        chmod 640 "/home/data/ies/${INI_ARRAY[i]}"
    done
    logger_info "chmod ies config success."
    return 0
}

ABILITY_POLICY_FILE="/home/data/config/ability_policy.json"

if [ ! -f "${ABILITY_POLICY_FILE}" ]; then
    safe_cp "${OM_WORK_DIR}/software/ibma/config/ability_policy_allow_all.json" "${ABILITY_POLICY_FILE}"
fi

chmod 644 "${ABILITY_POLICY_FILE}"
chmod_ies_config

SOFTWARE_PATH="${OM_WORK_DIR}/software"
if [ ! -d "${UDS_CERT_PATH}" ]; then
    mkdir -p "${UDS_CERT_PATH}"
    chmod 700 "${UDS_CERT_PATH}"
fi

if [ ! -e "${UDS_CERT_PATH}/om_alg.json" ]; then
    safe_cp "${SOFTWARE_PATH}/ibma/config/nginx_default/om_alg.json" "${UDS_CERT_PATH}"
    chmod 600 "${UDS_CERT_PATH}"/om_alg.json
fi
uid=$(id -u root)
gid=$(id -g root)
# create uds server cert
(
    check_root_path_and_exit "${SOFTWARE_PATH}/ibma"
    check_root_path_and_exit "${SOFTWARE_PATH}/ibma/opensource/python"
    export PYTHONPATH="${SOFTWARE_PATH}"/ibma:"${SOFTWARE_PATH}"/ibma/opensource/python
    check_root_path_and_exit "${OM_WORK_DIR}/scripts/python/create_server_certs.py"
    check_root_path_and_exit "${OM_WORK_DIR}/lib"
    export LD_LIBRARY_PATH="${OM_WORK_DIR}"/lib:${LD_LIBRARY_PATH}
    python3 "${OM_WORK_DIR}"/scripts/python/create_server_certs.py "${uid}" "${gid}" "${UDS_CERT_PATH}" "normal" "client"
)
ret=$?
if [ "${ret}" -ne 0 ]; then
    logger_error "Create uds server cert failed."
    return 1
fi

REDFISH_ROOT_PATH="${SOFTWARE_PATH}/RedfishServer/cert"
if [ ! -d "${REDFISH_ROOT_PATH}" ]; then
    su - MindXOM -s "$(which bash)" -c "mkdir -p ${REDFISH_ROOT_PATH}"
    su - MindXOM -s "$(which bash)" -c "chmod 700 ${REDFISH_ROOT_PATH}"
fi

REDFISH_BACKUP_DIR="/home/data/config/redfish_backup"
if [[ ! -d "${REDFISH_BACKUP_DIR}" ]]; then
    mkdir "${REDFISH_BACKUP_DIR}"
    safe_change_owner MindXOM MindXOM "${REDFISH_BACKUP_DIR}"
fi

# 拷贝一份client证书给MindXOM
function copy_client_cert_to_redfish()
{
    # uds证书
    local CERT_ARRAY=("om_alg.json" "om_cert.keystore" "om_cert_backup.keystore" "root_ca.cert" "client_kmc.cert" \
    "client_kmc.priv" "client_kmc.psd")
    # 拷贝client证书给MindXOM
    for ((i = 0; i < ${#CERT_ARRAY[@]}; i++)); do
        if ! check_soft_link "${REDFISH_ROOT_PATH}/${CERT_ARRAY[i]}"; then
            logger_error "redsifh contain soft link."
            continue
        fi

        if ! safe_cp "${UDS_CERT_PATH}/${CERT_ARRAY[i]}" "${REDFISH_ROOT_PATH}"/ -f; then
            continue
        fi
        safe_change_owner MindXOM MindXOM "${REDFISH_ROOT_PATH}/${CERT_ARRAY[i]}"
        su - MindXOM -s "$(which bash)" -c "chmod 600 ${REDFISH_ROOT_PATH}/${CERT_ARRAY[i]}"
    done

    logger_info "Copy client cert success."
    return 0
}

copy_client_cert_to_redfish

(
    check_root_path_and_exit "${OM_WORK_DIR}/lib"
    check_root_path_and_exit "${SOFTWARE_PATH}/ibma"
    check_root_path_and_exit "${SOFTWARE_PATH}/ibma/opensource/python"
    export LD_LIBRARY_PATH="${OM_WORK_DIR}"/lib:"${LD_LIBRARY_PATH}"
    export PYTHONPATH="${SOFTWARE_PATH}"/ibma:"${SOFTWARE_PATH}":"${SOFTWARE_PATH}"/ibma/opensource/python
    check_root_path_and_exit "${OM_WORK_DIR}/scripts/iBMA.sh"
    if ! /bin/bash "${OM_WORK_DIR}"/scripts/iBMA.sh restart; then
        logger_error "restart iBMA failed."
        exit 1
    fi
)

redfish_server_run="${SOFTWARE_PATH}"/RedfishServer/redfish_server.sh
if ! su - MindXOM -s "$(which bash)" -c "/bin/bash ${redfish_server_run}"; then
    logger_error "restart redfish server failed."
    exit 1
fi

