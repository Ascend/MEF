# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
