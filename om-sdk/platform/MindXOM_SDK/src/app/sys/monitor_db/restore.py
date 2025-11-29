# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from common.backup_restore_service.restore import Restore
from common.constants.base_constants import CommonConstants
from common.database_monitor import DatabaseMonitorBase
from common.log.logger import run_log
from common.utils.timer import RepeatingTimer


class DatabaseRestore(DatabaseMonitorBase):
    """ monitor进程数据库监控 """
    INTERVAL: int = 60

    def __init__(self):
        super().__init__(CommonConstants.MONITOR_EDGE_DB_FILE_PATH, CommonConstants.MONITOR_BACKUP_DIR)

    def monitor_database_status(self):
        """ 监控monitor进程数据库状态，异常则恢复 """
        check_ret = self.check_database_is_valid()
        if not check_ret:
            run_log.error(check_ret.error)
            # 恢复数据库
            Restore(self.backup_dir, self.db_path).entry()

    def monitor(self):
        RepeatingTimer(self.INTERVAL, self.monitor_database_status).start()
