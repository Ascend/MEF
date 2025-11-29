#!/bin/bash

# set process umask right
umask 0077

# get current path
CURRENT_PATH=$(cd "$(dirname "$0")" || exit 3; pwd)
ROOT_USER_ID="0"
ROOT_GROUP_ID="0"
ntp_etc_file="/etc/ntp.conf"
ntpd_file="/etc/sysconfig/ntpd"
ntp_enable_file="/home/data/ies/NTPEnable.ini"
NGINX_CONFIG_PATH="/usr/local/mindx/MindXOM/config/nginx.conf"
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

    local user_id
    local group_id
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

function check_config_file()
{
    local config_file=(
        "/home/data/ies/NTPEnable.ini"
        "/home/data/ies/mountCnf_site.ini"
    )

    for path in "${config_file[@]}"; do
        if ! check_soft_link "${path}"; then
            logger_error "${path} contain soft links!"
            return 1
        fi

        # 不存在则跳过属主校验
        if [[ ! -e "${path}" ]]; then
            continue
        fi

        if ! check_user_and_group_owner "${path}" "${ROOT_USER_ID}" "${ROOT_GROUP_ID}"; then
            logger_error "the ${path} owner is not right!"
            return 1
        fi
    done
}

function get_tec_presence()
{
    local i=0
    for ((i=0;i<10;i++))
    do
        tecTemp=$(npu-smi info -t temp -i 0 | grep "TEC_TEM" | awk '{print $4}')
        if [[ "${tecTemp}" == "" ]];then
            logger_error "Cold Start: get_tec_temp,retry $i times, get tec temp failed"
            sleep 1
        else
            logger_info "Cold Start: tectemp=$tecTemp"
            if [[ "${tecTemp}" != "NA" ]];then
                return 0
            else
                return 1
            fi
        fi
    done

    if [ "$i" -eq 10 ];then
        logger_error "Cold Start: get tec temp failed"
        return 2
    fi
}

function mount_config()
{
    local path="$1"
    local dev="$2"
    if [ -e "${dev}" ];then
        mount -t ext4 "${dev}" "${path}"
        logger_info "Cold Start: dev $dev found, mount $dev $path"
    else
        get_tec_presence
        tec_presence=$?
        # 若TEC不存在，则认为硬盘不存在，直接退出
        if [[ "${tec_presence}" -ne 0 ]];then
            logger_info "Cold Start: tec is not exists"
            mount -t ext4 "${dev}" "${path}"
            return
        fi
    fi
}

function check_and_restart_ntp_service()
{
    local ntp_client_enable
    local ntp_server_enable
    if [[ ! -f "${ntp_enable_file}" ]]; then
      return
    fi

    ntp_client_enable=$(${OS_CMD_CAT} ${ntp_enable_file} | awk -F ',' '{print $1}' | awk -F ':' '{print $2}' | awk '{gsub(/^\s+|\s+$/, "");print}')
    ntp_server_enable=$(${OS_CMD_CAT} ${ntp_enable_file} | awk -F ',' '{print $2}' | awk -F ':' '{print $2}' | awk '{gsub(/^\s+|\s+$/, "");print}')
    logger_info "Check ntp client enable state is [${ntp_client_enable}]."
    logger_info "Check ntp server enable state is [${ntp_server_enable}]."
    if [ "${ntp_server_enable}" == "True" ]; then
        sed -i 's/True/False/2' "${ntp_enable_file}"
    fi

    if [ "${ntp_client_enable}" != "True" ]; then
        logger_info "Get ntp client enable is False, stop ntp service."
        systemctl stop ntpd > /dev/null 2>&1
        return
    fi

    local remote_servers_array=()
    IFS=" " read -r -a remote_servers_array <<< "$(< "${ntp_etc_file}" grep '^server' | awk '{print $2}' | tr '\n' ' ')"
    if [[ "${#remote_servers_array[*]}" -gt 3 ]]; then
        logger_error "Get ntp remote servers num be more three, stop ntp service."
        ${OS_CMD_ECHO} "{'NTPClientEnable': False, 'NTPServerEnable': False, 'NTPLocalServers': ''}" > "${ntp_enable_file}"
        chmod 640 "${ntp_enable_file}"
        systemctl stop ntpd > /dev/null 2>&1
        return
    fi

    IFS=" " read -r -a remote_servers_array <<< "${remote_servers_array[*]/"127.127.1.0"}"
    local ntp_remote_servers="${remote_servers_array[*]}"
    if [[ "${#remote_servers_array[*]}" -eq 2 ]]; then
        ntp_remote_servers="${remote_servers_array[0]};${remote_servers_array[1]}"
    fi

    logger_info "Get ntp remote_servers is [${ntp_remote_servers}]."
    if [[ -z "${ntp_remote_servers}" ]]; then
        logger_error "Get ntp remote servers is null for ${ntp_etc_file}, stop ntp service."
        ${OS_CMD_ECHO} "{'NTPClientEnable': False, 'NTPServerEnable': False, 'NTPLocalServers': ''}" > "${ntp_enable_file}"
        chmod 640 "${ntp_enable_file}"
        systemctl stop ntpd > /dev/null 2>&1
        return
    fi

    local ntp_local_servers
    local nginx_listen_ip
    ntp_local_servers=$(${OS_CMD_CAT} ${ntp_enable_file} | awk -F "," '{print $3}' | awk -F ":" '{print $2}' | tr -d "'}" | awk '{gsub(/^\s+|\s+$/, "");print}')
    nginx_listen_ip=$(${OS_CMD_CAT} ${NGINX_CONFIG_PATH} | grep "listen" | head -n 1 | awk '{print $2}' | awk -F ":" '{print $1}' | awk '{gsub(/^\s+|\s+$/, "");print}')
    logger_info "Get ntp local servers ip is [${ntp_local_servers}]."
    logger_info "Get nginx listen ip is [${nginx_listen_ip}]."
    if [ "${ntp_local_servers}" != "${nginx_listen_ip}" ]; then
        sed -i "s/${ntp_local_servers}/${nginx_listen_ip}/" "${ntp_enable_file}"
        ntp_local_servers="${nginx_listen_ip}"
    fi

    if ! grep "interface" "${ntp_etc_file}" > /dev/null; then
        echo "interface ignore wildcard" >> "${ntp_etc_file}"
        echo "interface listen ${ntp_local_servers}" >> "${ntp_etc_file}"
    else
        sed -i '/interface listen/d' "${ntp_etc_file}"
        echo "interface listen ${ntp_local_servers}" >> "${ntp_etc_file}"
    fi

    local ntpd_option
    ntpd_option=$(< "${ntpd_file}" grep "OPTIONS" | awk -F "=" '{ print $2 }' | tr -d '"')
    if [ "${ntpd_option}" = "-g" ]; then
        sed -i 's/-g/-4 -g/g' "${ntpd_file}"
    fi

    # 重启ntpd服务
    if ! "${OM_WORK_DIR}"/scripts/ntp_service.sh "start" "${ntp_remote_servers}" "${ntp_local_servers}" "-f"; then
        logger_error "Exec ntp service scripts failed, stop ntp service."
        ${OS_CMD_ECHO} "{'NTPClientEnable': False, 'NTPServerEnable': False, 'NTPLocalServers': ''}" > "${ntp_enable_file}"
        chmod 640 "${ntp_enable_file}"
        systemctl stop ntpd > /dev/null 2>&1
        return
    fi

    systemctl restart ntpd > /dev/null 2>&1
    logger_info "Restart ntpd service finish."
}

function check_init_root_path()
{
    check_root_env_path=(
            "${CURRENT_PATH}"
            "${OM_WORK_DIR}"
            "/var/plog/ibma_edge"
            "/var/plog"
            "/usr/lib64"
            "/home/data/config"
            "/home/data/config/mindx_om_env.conf"
        )

    for path in "${check_root_env_path[@]}"; do
        check_root_env_path "${path}"
        check_root_env_ret=$?
        if [[ "${check_root_env_ret}" -ne 0 ]]; then
            echo "${path} check_root_env_path failed! ${check_root_env_ret}"
            exit "${check_root_env_ret}"
        fi
    done
}

check_init_root_path

if [[ -f "${OM_WORK_DIR}/lib/liblpeblock.so" ]]; then
    export LD_PRELOAD="${OM_WORK_DIR}/lib/liblpeblock.so"
fi
# Load log module
source "${CURRENT_PATH}"/log_print.sh
# Load comm_checker module
source "${CURRENT_PATH}"/comm_checker.sh
# init log
init_log_path "/var/plog/ibma_edge" "om_scripts_run.log"
logger_info "start the platform init!"

if ! check_config_file; then
    logger_error "check config file fail."
    exit 1
fi

# prepare docker root dir
(
    export PYTHONPATH="${OM_WORK_DIR}"/software/ibma:"${OM_WORK_DIR}"/software/ibma/opensource/python:"${OM_WORK_DIR}"/scripts/python
    python3  -u "${OM_WORK_DIR}"/scripts/python/prepare_docker_root_dir.py
)
ret=$?
if [ "${ret}" != 0 ];then
    logger_error "prepare docker root dir fail."
fi

# create mount dir
if [[ -f "/home/data/ies/mountCnf_site.ini" ]]; then
    < /home/data/ies/mountCnf_site.ini grep /dev | while read -r line
    do
        dev=$(echo "${line}" | cut -d "=" -f1 | awk '{gsub(/^\s+|\s+$/, "");print}')
        host_path=$(echo "${line}" | cut -d "=" -f2 | awk '{gsub(/^\s+|\s+$/, "");print}')
        mkdir -p "${host_path}" && chmod 1755 "${host_path}"
        if ! mount_config "${host_path}" "${dev}"; then
            logger_error "mount ${dev} ${host_path} failed"
        fi
    done
fi

# set time_zone
(
    export PYTHONPATH="${OM_WORK_DIR}"/software/ibma:"${OM_WORK_DIR}"/software/ibma/opensource/python:"${OM_WORK_DIR}"/scripts/python
    python3  -u "${OM_WORK_DIR}"/scripts/python/modify_time_zone.py
)
ret=$?
if [ "${ret}" != 0 ];then
    logger_error "set time_zone fail."
fi

# 初始化扩展存储配置，用于启动扩展存储告警检测
(
    export PYTHONPATH="${OM_WORK_DIR}"/software/ibma:"${OM_WORK_DIR}"/software/ibma/opensource/python:"${OM_WORK_DIR}"/scripts/python
    python3  -u "${OM_WORK_DIR}"/scripts/python/extend_storage_config.py
)
ret=$?
if [ "${ret}" != 0 ];then
    logger_error "load extend storage config fail."
fi

# 启动ensd进程
"${OM_WORK_DIR}"/software/ens/bin/start_ensd.sh

# ntpd服务
check_and_restart_ntp_service

# 启动seceye-agent进程
"${OM_WORK_DIR}"/scripts/seceye_agent_start.sh &

logger_info "finish the platform init!"
