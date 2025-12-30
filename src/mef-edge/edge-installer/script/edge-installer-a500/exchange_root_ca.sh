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
EDGE_INSTALLER_BIN_PATH="${CURRENT_PATH}/../bin"
SOFTWARE_PATH=$(realpath "${CURRENT_PATH}"/../..)
export LD_LIBRARY_PATH="${SOFTWARE_PATH}"/lib:$LD_LIBRARY_PATH

function main()
{
  "${EDGE_INSTALLER_BIN_PATH}"/innerctl exchange_certs "$@"
}

main "$@"
RESULT=$?
exit "${RESULT}"