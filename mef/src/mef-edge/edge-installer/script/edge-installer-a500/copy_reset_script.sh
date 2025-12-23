#!/bin/bash
# Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
EDGE_INSTALLER_BIN_PATH="${CURRENT_PATH}/../bin"

function main()
{
  "${EDGE_INSTALLER_BIN_PATH}"/innerctl copy_reset_script "$@"
}

main "$@"
RESULT=$?
exit "${RESULT}"