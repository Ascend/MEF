# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from typing import Tuple

from common.constants.base_constants import CommonConstants


class RedfishBackupRestoreCfg:
    """ redfish进程备份恢复配置 """

    BACKUP_FILES: Tuple[str] = (CommonConstants.REDFISH_EDGE_DB_FILE_PATH,)
    BACKUP_DIR: str = CommonConstants.REDFISH_BACKUP_DIR
