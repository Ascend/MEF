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

umask 0022

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
SERVICE_PATH=$(realpath "${CURRENT_PATH}/../service")

function safe_cp() {
    if [ $# -lt 2 ]; then
        echo "safe cp parameter error"
        return 1
    fi
    local srcpath
    local dstpath

    srcpath=$(realpath -s "$1")
    dstpath=$(realpath -s "$2")

    # srcpath is link
    if [ -L "${srcpath}" ]; then
        echo "${srcpath} does not comply with security rules." && return 1
    fi

    # dstpath is link
    if [ -L "${dstpath}" ]; then
        echo "${dstpath} does not comply with security rules." && return 1
    fi

    cp "$@"
    return 0
}

function copy_service() {
    MNT_PATH=$1

    if [[ "$MNT_PATH" != $(realpath "$MNT_PATH") ]]; then
        return 1
    fi

    local MntLinkPath="${MNT_PATH}"/lib
    MntRealPath=$(readlink -s "$MntLinkPath")
    if ! safe_cp "${SERVICE_PATH}"/* "${MNT_PATH}/${MntRealPath}"/systemd/system -rf; then
          return 1
    fi

    ln -sf /lib/systemd/system/device-plugin.service "$MNT_PATH"/etc/systemd/system/device-plugin.service
    if [[ $? != 0 ]]; then
      return 1
    fi

    ln -sf /lib/systemd/system/edgecore.service "$MNT_PATH"/etc/systemd/system/edgecore.service
    if [[ $? != 0 ]]; then
      return 1
    fi

    ln -sf /lib/systemd/system/mef-edge-main.service "$MNT_PATH"/etc/systemd/system/mef-edge-main.service
    if [[ $? != 0 ]]; then
      return 1
    fi

    ln -sf /lib/systemd/system/mef-edge-om.service "$MNT_PATH"/etc/systemd/system/mef-edge-om.service
    if [[ $? != 0 ]]; then
      return 1
    fi

    ln -sf /lib/systemd/system/mef-edge-init.service "$MNT_PATH"/etc/systemd/system/mef-edge-init.service
    if [[ $? != 0 ]]; then
      return 1
    fi

    mkdir "$MNT_PATH"/etc/systemd/system/mef-edge.target.wants

    ln -sf /lib/systemd/system/device-plugin.service "$MNT_PATH"/etc/systemd/system/mef-edge.target.wants/device-plugin.service
    if [[ $? != 0 ]]; then
      return 1
    fi

    ln -sf /lib/systemd/system/edgecore.service "$MNT_PATH"/etc/systemd/system/mef-edge.target.wants/edgecore.service
    if [[ $? != 0 ]]; then
      return 1
    fi

    ln -sf /lib/systemd/system/mef-edge-main.service "$MNT_PATH"/etc/systemd/system/mef-edge.target.wants/mef-edge-main.service
    if [[ $? != 0 ]]; then
      return 1
    fi

    ln -sf /lib/systemd/system/mef-edge-om.service "$MNT_PATH"/etc/systemd/system/mef-edge.target.wants/mef-edge-om.service
    if [[ $? != 0 ]]; then
      return 1
    fi

    ln -sf /lib/systemd/system/mef-edge-init.service "$MNT_PATH"/etc/systemd/system/mef-edge.target.wants/mef-edge-init.service
    if [[ $? != 0 ]]; then
      return 1
    fi
}

function main()
{
  MNT_PATH=$1
  copy_service "$MNT_PATH"
  return $?
}

main "$@"
RESULT=$?
exit "${RESULT}"