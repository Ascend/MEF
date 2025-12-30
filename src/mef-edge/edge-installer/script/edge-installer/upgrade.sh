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