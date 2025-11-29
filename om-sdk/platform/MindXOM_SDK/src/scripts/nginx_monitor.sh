#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
# Description: Script for monitoring nginx processes. Check whether the process exists every 10 seconds. If the process does not exist, restart the service.

NGINX_NAME="${OM_WORK_DIR}/software/nginx/sbin/nginx"

echo -1000 > /proc/$$/oom_score_adj

while true
do
    sleep 10

    if ! pgrep -f "${NGINX_NAME}" > /dev/null 2>&1
    then
        systemctl kill start-nginx.service
    fi
done
