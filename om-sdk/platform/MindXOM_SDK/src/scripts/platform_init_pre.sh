#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
# Description: platform-app.service start prepare script.

# 解决eth1网口配置默认网关之后，重启系统丢失默认网关问题
if systemctl status network > /dev/null 2>&1; then
	systemctl restart network > /dev/null 2>&1
fi

export LD_LIBRARY_PATH="${OM_WORK_DIR}"/lib:"${LD_LIBRARY_PATH}"
export PYTHONPATH="${OM_WORK_DIR}"/software/ibma:"${OM_WORK_DIR}"/software/ibma/opensource/python:"${OM_WORK_DIR}"/scripts/python
python3 -u "${OM_WORK_DIR}"/scripts/python/backup_restore_service.py &
