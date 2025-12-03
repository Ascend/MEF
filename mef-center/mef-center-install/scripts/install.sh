#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
umask 077
declare -i ret_ok=0  # success

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
export LD_LIBRARY_PATH=${CURRENT_PATH}/lib/kmc-lib:$LD_LIBRARY_PATH
INSTALL_PROGRAM_PATH="${CURRENT_PATH}/bin"
INSTALL_BIN="MEF-center-installer"

function main()
{
    "${INSTALL_PROGRAM_PATH}"/"${INSTALL_BIN}" "$@"
    ret=$?
    # help/version时不打印安装失败信息
    if [[ "${ret}" == 3 ]];then
        return ${ret}
    elif [[ "${ret}" != 0 ]];then
        echo "install MEF center failed"
        return ${ret}
    fi

    echo "install MEF center success"
    return ${ret_ok}
}

main "$@"
RESULT=$?
exit ${RESULT}