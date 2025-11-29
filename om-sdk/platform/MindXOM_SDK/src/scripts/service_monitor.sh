#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
