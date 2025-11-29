#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
# Description: platform-app.service exec stop nginx service script.

NGINX_NAME="${OM_WORK_DIR}/software/nginx/sbin/nginx"

if ! pgrep -f "${NGINX_NAME}" > /dev/null; then
    exit 0
fi

pstree -p "$(pgrep -f "${NGINX_NAME}")" | awk 'BEGIN{ FS="(" ; RS=")" } NF>1 { print $NF }' | xargs kill -9 &> /dev/null
