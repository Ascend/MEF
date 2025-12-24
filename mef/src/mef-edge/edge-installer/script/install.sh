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
declare -i ret_ok=0  # success
declare -i ret_flag_parse_error=2
declare -i ret_print_info=3

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
INSTALL_PROGRAM_PATH="${CURRENT_PATH}/software/edge_installer/bin"
export LD_LIBRARY_PATH=${CURRENT_PATH}/software/lib:$LD_LIBRARY_PATH

function main()
{
    local os_arch
    os_arch=$(arch)
    local install_bin_arch
    install_bin_arch=$(file "${INSTALL_PROGRAM_PATH}"/install)
    if [[ ! (${install_bin_arch} =~ ${os_arch}) ]]; then
        echo "the device is ${os_arch} system, upgrade package is not, please check."
        return 1
    fi

    local ret
    "${INSTALL_PROGRAM_PATH}"/install "$@"
    ret=$?
    if [ "${ret}" == $ret_flag_parse_error ];then
        echo "input parameters error"
        return ${ret}
    fi
    if [ "${ret}" == $ret_print_info ];then
        return ${ret_ok}
    fi
    if [[ "${ret}" != 0 ]];then
        echo "install MEFEdge failed"
        return ${ret}
    fi

    echo "install MEFEdge success"
    return ${ret_ok}
}

main "$@"
RESULT=$?
exit ${RESULT}