# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os

from common.backup_restore_service.backup import Backup
from common.constants.base_constants import CommonConstants
from common.file_utils import FileCreate
from common.utils.timer import RepeatingTimer


class DatabaseBackup(Backup):
    LOOP_INTERVAL: int = 30

    def __init__(self):
        if not os.path.exists(CommonConstants.MONITOR_BACKUP_DIR):
            FileCreate.create_dir(CommonConstants.MONITOR_BACKUP_DIR, 0o700)
        super().__init__(CommonConstants.MONITOR_BACKUP_DIR, CommonConstants.MONITOR_EDGE_DB_FILE_PATH)

    def entry(self):
        RepeatingTimer(self.LOOP_INTERVAL, super().entry).start()
