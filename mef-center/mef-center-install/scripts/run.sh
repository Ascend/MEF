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
declare -i ret_ok=0  # success
declare -i ret_failed=1  # failed

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
TOOL_BINARY_PATH="${CURRENT_PATH}/bin"
export LD_LIBRARY_PATH=${CURRENT_PATH}/lib/kmc-lib:${CURRENT_PATH}/lib/lib:$LD_LIBRARY_PATH

function main()
{
    binary_file="MEF-center-controller"
    "${TOOL_BINARY_PATH}"/"${binary_file}" "$@"
}

main "$@"
RESULT=$?
exit ${RESULT}