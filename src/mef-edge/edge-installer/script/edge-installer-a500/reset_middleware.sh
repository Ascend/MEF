#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MEF is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

# 该文件在A500 A2恢复出厂设置|恢复最小系统的时候，放在p1|p7目录下，主要作用是用于注册reset_mefedge服务。

dev=$1
mode=$2
ResetFactory="reset_factory"
if [ "${dev}" = "/dev/mmcblk0" ];then
  if [ "${mode}" = "${ResetFactory}" ];then
    GoldDev="${dev}p1"
  else
    GoldDev="${dev}p7"
  fi
  P2="${dev}p2"
  P3="${dev}p3"
  P6="${dev}p6"
else
  if [ "${mode}" = "${ResetFactory}" ];then
    GoldDev="${dev}1"
  else
    GoldDev="${dev}7"
  fi
  P2="${dev}2"
  P3="${dev}3"
  P6="${dev}6"
fi

CheckGoldDevMount=$(df -h | awk -v dev="${GoldDev}" '{if($1==dev){print $6}}')
CheckP2Mount=$(df -h | awk -v dev="$P2" '{if($1==dev){print $6}}')
CheckP3Mount=$(df -h | awk -v dev="$P3" '{if($1==dev){print $6}}')
CheckP6Mount=$(df -h | awk -v dev="$P6" '{if($1==dev){print $6}}')

Gold="/mnt/mmc/p1"
P2Mount="/mnt/mmc/p2"
P3Mount="/mnt/mmc/p3"

EdgePath="/usr/local/mindx"
EdgeUnpackDir="MEFEdgePkg"
LogRootDir="/var/alog/MEFEdge_log/"
LogDir="${LogRootDir}/edge_installer/"
ResetLog="${LogDir}"/reset_mefedge.log

declare -i Err_OK=0
declare -i Err_EdgePathAlreadyExist=1
declare -i Err_MkdirEdgePathFailed=2
declare -i Err_MountP2Failed=3
declare -i Err_MountP3Failed=4
declare -i Err_MountP6Failed=5
declare -i Err_MkdirEdgeUnpackPathFailed=6
declare -i Err_UnpackEdgePkgFailed=7


function check_log_dir()
{
    local dir_path=$1
    local dir_mode=$2

    if [[ -L "${dir_path}" ]]; then
        unlink "${dir_path}"
    elif [[ -f "${dir_path}" ]]; then
        unlink "${dir_path}"
    fi

    if [[ ! -d "${dir_path}" ]]; then
      mkdir -p "${dir_path}"
    fi

    chmod "${dir_mode}" "${dir_path}"
}

function check_log_file()
{
    local log_path=$1
    if [[ -L "${log_path}" ]]; then
        unlink "${log_path}"
    elif [[ -d "${log_path}" ]]; then
        rm -r "${log_path}"
    fi

    if [[ ! -f "${log_path}" ]]; then
        touch "${log_path}"
    fi

    chmod 640 "${log_path}"
}

function log()
{
    level=$1
    shift 1
    check_log_file "${ResetLog}"
    echo "$(date) ${level}- $*" >> "${ResetLog}"
    echo "$*"
}

function logger_Info()
{
    log INFO "$@"
}

function logger_Error()
{
    log ERROR "$@"
}

function create_log_dir()
{
    check_log_dir "${LogRootDir}" 755
    check_log_dir "${LogDir}" 750
}

function set_env()
{
    Gold_Mount="${CheckGoldDevMount}"
    if [ "${Gold_Mount}" ]; then
        Gold="${Gold_Mount}"
    fi

    P6_Mount="${CheckP6Mount}"
    if [ "${P6_Mount}" ]; then
        EdgePath="${P6_Mount}"
        return "${Err_OK}"
    fi

    if [ -d "${EdgePath}" ];then
        logger_Error "${EdgePath} already existed"
        return "${Err_EdgePathAlreadyExist}"
    fi

    if ! mkdir -p "${EdgePath}";then
        logger_Error "mkdir ${EdgePath} failed"
        return "${Err_MkdirEdgePathFailed}"
    fi

    blockdev --setrw "${P6}"
    if ! mount "${P6}" "${EdgePath}";then
        logger_Error "mount ${P6} to ${EdgePath} failed"
        return "${Err_MountP6Failed}"
    fi

    return "${Err_OK}"
}

function  del_residual_files() {
    rm -rf "${EdgePath:?}"/"${EdgeUnpackDir}"
    rm -rf "${EdgePath}"/MEFEdge/
}

function unpack_software_pkg()
{
    if [ "${mode}" = "${ResetFactory}" ];then
      MEFEdgeTar=$(ls "${Gold}"/Ascend-mindxedge-mefedge*.tar.gz )
    else
      MEFEdgeTar=$(ls "${Gold}"/firmware/Ascend-mindxedge-mefedge*.tar.gz )
    fi

    if ! mkdir -p "${EdgePath}"/"${EdgeUnpackDir}";then
      return "${Err_MkdirEdgeUnpackPathFailed}"
    fi

    if ! tar xzf "${MEFEdgeTar}" -C "${EdgePath}"/"${EdgeUnpackDir}";then
      return "${Err_UnpackEdgePkgFailed}"
    fi
}

function mount_partition() {
    if [[ ! -d "$2" ]]; then
        rm -rf "$2"
        mkdir -p "$2"
    fi
    if mount -t ext4 "$1" "$2"; then
        logger_Info "mount $1 to $2 success"
        return "${Err_OK}"
    fi
    logger_Error "mount $1 to $2 failed"
    return 1
}

function register_service()
{
    P2_Mount="${CheckP2Mount}"
    if [ "${P2_Mount}" ]; then
        P2Mount="${P2_Mount}"
    else
      if ! mount_partition "${P2}" "${P2Mount}";then
        return "${Err_MountP2Failed}"
      fi
    fi
    logger_Info "partition ${P2} is mounted to ${P2Mount}"

    P3_Mount="${CheckP3Mount}"
    if [ "${P3_Mount}" ]; then
        P3Mount="${P3_Mount}"
    else
      if ! mount_partition "${P3}" "${P3Mount}";then
        return "${Err_MountP3Failed}"
      fi
    fi
    logger_Info "partition ${P3} is mounted to ${P3Mount}"

    ServicePath="${EdgePath}"/"${EdgeUnpackDir}"/software/edge_installer/service/reset_mefedge.service
    sed -i "s|{pkg_dir}|${EdgePath}/${EdgeUnpackDir}|g" "${ServicePath}"
    cp -f "${ServicePath}" "${P2Mount}"/lib/systemd/system/reset_mefedge.service
    cp -f "${ServicePath}" "${P3Mount}"/lib/systemd/system/reset_mefedge.service
    ln -sf /lib/systemd/system/reset_mefedge.service "${P2Mount}"/etc/systemd/system/multi-user.target.wants/reset_mefedge.service
    ln -sf /lib/systemd/system/reset_mefedge.service "${P3Mount}"/etc/systemd/system/multi-user.target.wants/reset_mefedge.service
}

function main()
{
    create_log_dir
    logger_Info "start reset MEFEdge"

    set_env
    ret=$?
    if [ "${ret}" -ne 0 ];then
      logger_Error "set env failed:${ret}"
      return "${ret}"
    fi
    logger_Info "set env success"

    del_residual_files
    unpack_software_pkg
    ret=$?
    if [ "${ret}" -ne 0 ];then
      logger_Error "unpack software package failed:${ret}"
      return "${ret}"
    fi
    logger_Info "unpack software package success"

    register_service
    ret=$?
    if [ "${ret}" -ne 0 ];then
      logger_Error "register service failed:${ret}"
      return "${ret}"
    fi
    logger_Info "register service success"
    logger_Info "reset MEFEdge success"
    return "${Err_OK}"
}

main "$@"
RESULT=$?
exit "${RESULT}"
