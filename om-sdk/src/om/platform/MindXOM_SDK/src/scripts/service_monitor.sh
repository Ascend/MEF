#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
# Description: Script for monitoring service processes. Check whether the process exists every 10 seconds. If the process does not exist, kill the service.

PROCESSES=(
    "${OM_WORK_DIR}/software/ibma/bin/monitor.py"
    "${OM_WORK_DIR}/software/RedfishServer/ibma_redfish_main.py"
)

echo -1000 > /proc/$$/oom_score_adj

while true
do
    sleep 10

    for process in "${PROCESSES[@]}"
    do
        if ! pgrep -f "${process}" > /dev/null 2>&1
        then
            systemctl kill ibma-edge-start.service
        fi
    done
done
