#!/bin/bash
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

umask 0022

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
UPGRADE_PATH="${CURRENT_PATH}/../bin"
export LD_LIBRARY_PATH=${CURRENT_PATH}/../../lib:$LD_LIBRARY_PATH

function main()
{
    local os_arch
    os_arch=$(arch)
    local upgrade_bin_arch
    upgrade_bin_arch=$(file "${UPGRADE_PATH}"/upgrade)
    if [[ ! (${upgrade_bin_arch} =~ ${os_arch}) ]]; then
        echo "the device is ${os_arch} system, upgrade package is not ${os_arch}, please check."
        return 1
    fi
    local ret
    "${UPGRADE_PATH}"/upgrade "$@"
    ret=$?
    return ${ret}
}

main "$@"
RESULT=$?
exit ${RESULT}