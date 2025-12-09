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
