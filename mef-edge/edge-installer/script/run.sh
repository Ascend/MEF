#!/bin/bash
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

umask 0022

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
EDGE_CTL_PATH="${CURRENT_PATH}/edge_installer/bin"
export LD_LIBRARY_PATH=${CURRENT_PATH}/lib:$LD_LIBRARY_PATH

function main()
{
  "${EDGE_CTL_PATH}"/edgectl "$@"
}

main "$@"
RESULT=$?
exit ${RESULT}