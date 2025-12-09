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
# Description: platform-app.service exec stop nginx service script.

NGINX_NAME="${OM_WORK_DIR}/software/nginx/sbin/nginx"

if ! pgrep -f "${NGINX_NAME}" > /dev/null; then
    exit 0
fi

pstree -p "$(pgrep -f "${NGINX_NAME}")" | awk 'BEGIN{ FS="(" ; RS=")" } NF>1 { print $NF }' | xargs kill -9 &> /dev/null
