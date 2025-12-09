#!/bin/bash

#add copy web-code and start nginx
target_cert_dir=/home/data/config/default
source /home/data/config/os_cmd.conf
source "${OM_WORK_DIR}"/scripts/safe_common.sh
source "${OM_WORK_DIR}"/scripts/comm_checker.sh
source "${OM_WORK_DIR}"/scripts/log_print.sh

# init log
init_log_path "/var/plog/ibma_edge" "om_scripts_run.log"

function copy_nginx_cert_to_redfish()
{
    # nginx证书
    local redfish_cert_dir="/home/data/config/redfish"
    local NGINX_CERT_ARRAY=("om_alg.json" "om_cert.keystore" "om_cert_backup.keystore" "server_kmc.cert" \
    "server_kmc.priv" "server_kmc.psd")
    if [[ -d "${redfish_cert_dir}" ]]; then
        for ((i = 0; i < ${#NGINX_CERT_ARRAY[@]}; i++)); do
            unlink "${redfish_cert_dir}/${NGINX_CERT_ARRAY[i]}"
        done
    else
        mkdir "${redfish_cert_dir}"
        chmod 700 "${redfish_cert_dir}"
        safe_change_owner MindXOM MindXOM "${redfish_cert_dir}"
    fi

    # 拷贝nginx证书给redfish
    for ((i = 0; i < ${#NGINX_CERT_ARRAY[@]}; i++)); do
        if ! safe_cp "${target_cert_dir}/${NGINX_CERT_ARRAY[i]}" "${redfish_cert_dir}"/ -f; then
            return 1
        fi
        safe_change_owner MindXOM MindXOM "${redfish_cert_dir}/${NGINX_CERT_ARRAY[i]}"
        su - MindXOM -s "$(which bash)" -c "chmod 600 ${redfish_cert_dir}/${NGINX_CERT_ARRAY[i]}"
    done

    logger_info "Copy nginx cert success."
    return 0
}

function copy_nginx_conf()
{
    if ! safe_cp "${OM_WORK_DIR}/config/nginx.conf" "${OM_WORK_DIR}/software/nginx/conf/nginx.conf"; then
        return 1
    fi

    if ! safe_change_owner nobody nobody "${OM_WORK_DIR}/software/nginx/conf/nginx.conf"; then
        return 1
    fi

    return 0
}

function create_cert()
{
    # 生成证书
    (
        export LD_LIBRARY_PATH="${OM_WORK_DIR}"/lib:"${LD_LIBRARY_PATH}"
        export PYTHONPATH="${OM_WORK_DIR}"/software/ibma:"${OM_WORK_DIR}"/software/ibma/opensource/python:${PYTHONPATH}
        python3 "${OM_WORK_DIR}"/scripts/python/create_server_certs.py "$(id -u nobody)" "$(id -g nobody)" "${target_cert_dir}" "normal"
    )
    return $?
}

function check_cert_existent()
{
    local i
    local cert_files=("server_kmc.cert" "server_kmc.priv" "server_kmc.psd" "om_cert.keystore" "om_cert_backup.keystore" "om_alg.json")
    for (( i= 0; i < "${#cert_files[@]}"; i++)); do
        if [[ ! -s "${target_cert_dir}/${cert_files[i]}" ]]; then
            return 1
        fi
    done
    return 0
}

# 如果nginx进程已启动，则删除其父子进程
NGINX_NAME="${OM_WORK_DIR}/software/nginx/sbin/nginx"
if pgrep -f "${NGINX_NAME}" > /dev/null; then
    pstree -p "$(pgrep -f "${NGINX_NAME}")" | awk 'BEGIN{ FS="(" ; RS=")" } NF>1 { print $NF }' | xargs kill -9 &> /dev/null
fi

if ! check_soft_link "${OM_WORK_DIR}"/software/nginx/conf/nginx.conf; then
    logger_error "Nginx config file contain soft links!"
    exit 1
fi

if ! create_cert; then
    logger_error "Create https server cert failed."
    exit 1
fi

if ! check_cert_existent; then
    logger_error "server_kmc.cert server_kmc.psd server_kmc.priv not exist."
    exit 1
fi

# 拷贝一份证书给redfish_Main进程使用
if ! copy_nginx_cert_to_redfish; then
    logger_error "Copy https server cert failed."
    exit 1
fi

# 拷贝nginx配置文件
if ! copy_nginx_conf; then
    logger_error "Copy nginx conf failed."
    exit 1
fi

# 修改nginx二进制权限
mkdir -p /run/nginx && chmod 700 /run/nginx
safe_change_owner nobody nobody /run/nginx
setcap 'cap_net_bind_service=+ep' "${OM_WORK_DIR}"/software/nginx/sbin/nginx
(
    # 需要以nobody用户启动脚本执行
    su - nobody -s /bin/bash -c "LD_LIBRARY_PATH=${OM_WORK_DIR}/software/nginx/lib PYTHONPATH=${OM_WORK_DIR}/software/nginx/python python3 -u ${OM_WORK_DIR}/software/nginx/python/start_nginx.py"
)
ret=$?
if [ ${ret} -ne 0 ]; then
    logger_error "Start nginx failed, ret ${ret}."
    exit 1
fi
logger_info "start_nginx.sh done "
