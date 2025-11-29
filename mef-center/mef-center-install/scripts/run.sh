#!/bin/bash
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
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