#!/bin/bash
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
umask 077
declare -i ret_ok=0  # success

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
UPGRADE_BIN_PATH="${CURRENT_PATH}/../bin"
UPGRADE_BIN="MEF-center-upgrade"

function main()
{
    "${UPGRADE_BIN_PATH}"/"${UPGRADE_BIN}"
    ret=$?
    return ${ret}
}

main "$@"
RESULT=$?
exit ${RESULT}