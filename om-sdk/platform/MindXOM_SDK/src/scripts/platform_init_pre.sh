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
# Description: platform-app.service start prepare script.

# 解决eth1网口配置默认网关之后，重启系统丢失默认网关问题
if systemctl status network > /dev/null 2>&1; then
	systemctl restart network > /dev/null 2>&1
fi

export LD_LIBRARY_PATH="${OM_WORK_DIR}"/lib:"${LD_LIBRARY_PATH}"
export PYTHONPATH="${OM_WORK_DIR}"/software/ibma:"${OM_WORK_DIR}"/software/ibma/opensource/python:"${OM_WORK_DIR}"/scripts/python
python3 -u "${OM_WORK_DIR}"/scripts/python/backup_restore_service.py &
