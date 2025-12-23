#!/bin/bash

CUR_DIR=$(dirname "$(readlink -f "$0")")
source "${CUR_DIR}"/../safe_common.sh
# 这个路径和docker.service 里面的跟路径要匹配
Default_Path="/var/lib/docker"
Root_dir=${Default_Path}/root_dir
Mount_info_path=$CUR_DIR/"mount_info.txt"
Dev_info_path=$CUR_DIR/"dev_info.txt"
OUT_FILE=$CUR_DIR/"dev_out_list.txt"
declare -a Mount_info_list

function log_info() {
    printf "[info]:%s \n" "$*"
}

function log_error() {
    printf "[error]:%s \n" "$*"
}

function log_warn() {
    printf "[warn]:%s \n" "$*"
}

function record_one() {
    local input_dev_path=$1
    local real_dev_path
    local out_file=$2
    local st_mode st_type st_uid st_gid type_name major_id minor_id
    if [ -z "$input_dev_path" ]; then
        return 0
    fi

    real_dev_path=$(realpath "$input_dev_path")
    if [ ! -e "$real_dev_path" ]; then
        log_warn "the dev file path [$input_dev_path] not exist, so skipped"
        return 0
    fi

    st_mode=$(stat -c %a "$real_dev_path")
    st_uid=$(stat -c %u "$real_dev_path")
    st_gid=$(stat -c %g "$real_dev_path")
    major_id=$(stat -c %t "$real_dev_path")
    minor_id=$(stat -c %T "$real_dev_path")
    type_name=$(stat -c %F "$real_dev_path")
    if [ "$type_name" == "character special file" ]; then
        st_type="c"
    elif [ "$type_name" == "block special file" ]; then
        st_type="b"
    elif [ "$type_name" == "fifo" ]; then
        st_type="p"
    else
        log_warn "the device path [$input_dev_path] with type [$type_name] not be supported"
        return 1
    fi
    echo "$input_dev_path" "$st_type" "0x$major_id" "0x$minor_id" "$st_mode" "$st_uid" "$st_gid" >>"$out_file"
    return 0
}

function record() {
    local input_white_list_file=$1
    local one_dev_path
    rm -f "$OUT_FILE"
    touch "$OUT_FILE"
    if ! safe_chmod 600 "$OUT_FILE"; then
        log_error "chmod [$OUT_FILE] failed"
        return 1
    fi
    if [ ! -f "$input_white_list_file" ]; then
        log_error "the white list file [$input_white_list_file] not exists"
        return 1
    fi

    while IFS= read -r one_dev_path; do
        record_one "$one_dev_path" "$OUT_FILE"
    done <"$input_white_list_file"
    log_info "record the device path info success"
    return 0
}

function backup_primal_docker_service() {
    local edge_installer_config_path="$1"
    local primal_docker_service_path="/usr/lib/systemd/system/docker.service"
    local docker_service_backup_file="docker.service.bak"

    if [ -f "${edge_installer_config_path}/${docker_service_backup_file}" ]; then
        log_info "the docker.service file has been backed up."
        return
    fi

    if ! safe_cp "${primal_docker_service_path}" "${edge_installer_config_path}/${docker_service_backup_file}"; then
        log_error "cp [${primal_docker_service_path}] failed"
        return
    fi

    if ! safe_chmod 600 "${edge_installer_config_path}/${docker_service_backup_file}"; then
        log_error "chmod [${docker_service_backup_file}] failed"
        return 1
    fi

    log_info "primal docker.service backup success"
}

function copy_service() {
    local docker_service=$1
    local real_docker
    real_docker=$(realpath "$docker_service")
    if ! safe_cp "$real_docker" /usr/lib/systemd/system/docker.service -f; then
        log_error "cp [${real_docker}] failed"
        return 1
    fi
    systemctl daemon-reload
    log_info "copy docker service [$real_docker] to system success"
}

function stop_docker() {
    systemctl stop docker
    log_info "stop docker service"
}

function restart_docker() {
    systemctl restart docker
    log_info "restart docker service success"
    return 0
}

function do_mount() {
    local mount_type=$1
    local mount_src_path=$2
    local mount_dst_path=$3

    if ! mountpoint -q "${mount_dst_path}"; then
        if ! mount "$mount_type" "$mount_src_path" "$mount_dst_path"; then
            log_warn "mount [${mount_dst_path}] failed"
        fi
    else
        log_warn "the path [${mount_dst_path}] already mounted"
    fi
}

function get_mount_info() {
    Mount_info_list=($(awk '{print $NF}' "$Mount_info_path"))
}

function umount_files() {
    for f in "${Mount_info_list[@]}"; do
        if [ -z "$f" ]; then
            continue
        fi
        umount --recursive "${Root_dir}"/"${f}" > /dev/null 2>&1 || true
    done
    log_info "umount files success "
}

function clean_root_dir() {
    for f in "${Mount_info_list[@]}"; do
        if [ -z "$f" ]; then
            continue
        fi
        if ! mountpoint -q "${Root_dir}"/"${f}"; then
            rm -rf "${Root_dir:?}"/"${f}"
        fi
    done
}

function create_mount_point() {
    for f in "${Mount_info_list[@]}"; do
        if [ -d "${f}" ]; then
            mkdir -p "${Root_dir}"/"${f}"
        elif [ -f "${f}" ]; then
            mkdir -p "$(dirname "${Root_dir}"/"${f}")"
            touch "${Root_dir}"/"${f}"
        fi
    done
}

function mount_files() {
    local file_line mount_info_arr
    while IFS= read -r file_line; do
        read -r -a mount_info_arr <<<"$file_line"
        if [ "${mount_info_arr[1]}" == "/var/lib/docker" ]; then
            do_mount "${mount_info_arr[0]}" "${mount_info_arr[1]}/docker_path" "${Root_dir}"/"${mount_info_arr[1]}"
            continue
        fi
        do_mount "${mount_info_arr[0]}" "${mount_info_arr[1]}" "${Root_dir}"/"${mount_info_arr[1]}"
    done <"$Mount_info_path"
    log_info "mount files from mount_info.txt [$Mount_info_path] success"
}

function prepare_root_dir() {
    local real_path link
    mkdir -p "/var/lib/docker/kubelet" && chmod 750 "/var/lib/docker/kubelet"
    mkdir -p "/var/lib/docker/modelfile" && chmod 700 "/var/lib/docker/modelfile"
    mkdir -p "/var/lib/docker/docker_path" && chmod 755 "/var/lib/docker/docker_path"
    mkdir -p "/var/lib/docker/docker_path/kubelet" && chmod 755 "/var/lib/docker/docker_path/kubelet"
    mkdir -p "/var/lib/docker/docker_path/modelfile" && chmod 755 "/var/lib/docker/docker_path/modelfile"
    mkdir -p "/etc/docker/certs.d" && chmod 750 "/etc/docker/certs.d"
    clean_root_dir
    create_mount_point
    record "$Dev_info_path"

    local resolv_link_path=/etc/resolv.conf
    resolv_real_path=$(readlink -f "$resolv_link_path")
    if ! safe_cp "${resolv_real_path}" "${Root_dir}"/etc/resolv.conf; then
      log_error "cp /etc/resolv.conf failed"
      return 1
    fi
    if ! safe_cp "$CUR_DIR"/docker_entrypoint.sh "${Root_dir}"/; then
      log_error "cp docker_entrypoint.sh failed"
      return 1
    fi
    if ! safe_cp "$OUT_FILE" "${Root_dir}"/; then
      log_error "cp [${OUT_FILE}] failed"
      return 1
    fi
    if ! safe_chmod 500 "${Root_dir}"/docker_entrypoint.sh; then
      log_error "chmod docker_entrypoint.sh failed"
      return 1
    fi
    rm -f "$OUT_FILE"

    if [ -L /run/docker.sock ]; then
      rm /run/docker.sock
    fi
    ln -sf "${Root_dir}"/var/run/docker.sock /run/docker.sock
    pushd "$Root_dir" >/dev/null || return 1
    find / -maxdepth 1 -type l -print0 | while IFS= read -r -d '' link
    do
        real_path=$(realpath "$link")
        if [[ ${real_path} == /usr/* ]]; then
            [ ! -L "${link#/}" ] && ln -sf "${real_path#/}" "${link#/}" || true
        fi
    done
    popd >/dev/null || return 1
    log_info "prepare root dir success"
}

function check_docker() {
    local mount_path docker_status
    for mount_path in "${Mount_info_list[@]}"; do
        if [ -z "$mount_path" ]; then
            continue
        fi
        if ! mountpoint -q "${Root_dir}"/"${mount_path}"; then
            log_error "check $mount_path mount point failed, it is not mounted"
            return 1
        fi
    done

    docker_status=$(systemctl is-active docker)
    if [ "$docker_status" != "active" ]; then
        log_error "docker service is not running"
        return 1
    fi

    if [ -L /run/docker.sock ]; then
      rm /run/docker.sock
    fi
    ln -sf "${Root_dir}"/var/run/docker.sock /run/docker.sock
    return 0
}

function main() {
    local docker_service_path=$1
    local edge_installer_config_path=$2
    local mount_info_path=$3
    local dev_info_path=$4
    local install_path=$5

    if [ -z "$docker_service_path" ]; then
        log_error "the input docker service path is null"
        return 1
    fi

    if [ -z "${edge_installer_config_path}" ]; then
        log_error "the input edge om config path is null"
        return 1
    fi

    if [ -n "${mount_info_path}" ]; then
        log_info "use the input mount info path [$mount_info_path]"
        Mount_info_path=$mount_info_path
    fi

    if [ -n "${dev_info_path}" ]; then
        log_info "use the input dev info path [$dev_info_path]"
        Dev_info_path=$dev_info_path
    fi

    if [ -n "${install_path}" ]; then
        Root_dir=$install_path/root_dir
        log_info "using input root_dir [${Root_dir}]"
    else
        log_info "using default root_dir [${Root_dir}]"
    fi

    get_mount_info

    if check_docker; then
        # mode of model file dir need to reset to 700 for keeping the verifications be same
        chmod 700 "/var/lib/docker/modelfile"
        log_info "docker is new namespace status, no need to create again"
        return 0
    fi
    backup_primal_docker_service "${edge_installer_config_path}"
    copy_service "$docker_service_path"
    stop_docker
    umount_files
    prepare_root_dir
    mount_files
    restart_docker
    return $?
}
chattr -i "$CUR_DIR"
main "$@"
chattr +i "$CUR_DIR"
exit $?
