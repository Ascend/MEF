#!/bin/bash

CUR_DIR=$(dirname "$(readlink -f "$0")")
source "${CUR_DIR}"/../safe_common.sh
MOUNT_INFO_PATH=${CUR_DIR}/"mount_info.txt"
DOCKER_SERVICE_BACKUP_FILE="docker.service.bak"

function log_info()
{
    printf "[info]:%s \n" "$*"
}

function log_error()
{
    printf "[error]:%s \n" "$*"
}

function umount_docker_isolation_path()
{
    local mount_info_list=($(awk '{print $NF}' "${MOUNT_INFO_PATH}"))
    for f in "${mount_info_list[@]}"; do
        if [ -z "${f}" ]; then
            continue
        fi
        umount -f -R "/var/lib/docker/root_dir/${f}" > /dev/null 2>&1 || true
    done
    log_info "umount docker isolation path success"
}

function stop_docker()
{
    systemctl stop docker
    log_info "stop docker"
}

function is_need_restore_service()
{
    local edge_installer_config_path="$1"
    cat "${edge_installer_config_path}/${DOCKER_SERVICE_BACKUP_FILE}" | grep -E "docker_entrypoint.sh"
    local ret=$?
    if [ "${ret}" -eq 0 ]; then
        log_info "service is not primal docker service, no need to restore"
        return 1
    fi
    return 0
}

function restore_docker_service() {
    local edge_installer_config_path="$1"
    local docker_service_path="/usr/lib/systemd/system/docker.service"
    if ! safe_cp "${edge_installer_config_path}/${DOCKER_SERVICE_BACKUP_FILE}" "${docker_service_path}" -f; then
        log_error "cp [${edge_installer_config_path}/${DOCKER_SERVICE_BACKUP_FILE}] failed"
        return 1
    fi
    systemctl stop docker.socket > /dev/null 2>&1
    unlink /run/docker.sock
    systemctl daemon-reload
    systemctl restart docker
    local ret=$?
    if [ "${ret}" -ne 0 ]; then
        log_info "restore docker service failed, try again"
        systemctl daemon-reload
        systemctl restart docker
        ret=$?
    fi
    if [ "${ret}" -ne 0 ]; then
        log_error "restore docker service failed"
    else
        log_info "restore docker service success"
    fi
    return ${ret}
}

function main()
{
    local edge_installer_config_path="$1"
    local ret=0
    is_need_restore_service "${edge_installer_config_path}"
    ret=$?
    if [ "${ret}" -ne 0 ]; then
        log_info "no need to restore docker service"
        return 0
    fi
    stop_docker
    umount_docker_isolation_path
    restore_docker_service "${edge_installer_config_path}"
    ret=$?
    return ${ret}
}

main "$@"
exit $?
