#!/bin/bash
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

umask 0022

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
EDGE_INSTALLER_BIN_PATH="${CURRENT_PATH}/../bin"

function main()
{
  "${EDGE_INSTALLER_BIN_PATH}"/innerctl recovery "$@"
}

main "$@"
RESULT=$?
exit "${RESULT}"