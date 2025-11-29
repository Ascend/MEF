#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
# Description: ibma-edge-start.service exec stop redfish service script.

REDFISH_NAME="${OM_WORK_DIR}/software/RedfishServer/ibma_redfish_main.py"

if ! pgrep -f "${REDFISH_NAME}" > /dev/null 2>&1; then
    exit 0
fi

kill -9 "$(pgrep -f "${REDFISH_NAME}")"
