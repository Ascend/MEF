#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
# Description: platform-app.service start post-process script.
source /usr/local/mindx/MindXOM/scripts/safe_common.sh
if is_web_access_enable; then
  "${OM_WORK_DIR}/scripts/nginx_monitor.sh" &
fi