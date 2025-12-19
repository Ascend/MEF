# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from typing import Tuple

from common.constants.base_constants import CommonConstants


class RedfishBackupRestoreCfg:
    """ redfish进程备份恢复配置 """

    BACKUP_FILES: Tuple[str] = (CommonConstants.REDFISH_EDGE_DB_FILE_PATH,)
    BACKUP_DIR: str = CommonConstants.REDFISH_BACKUP_DIR
