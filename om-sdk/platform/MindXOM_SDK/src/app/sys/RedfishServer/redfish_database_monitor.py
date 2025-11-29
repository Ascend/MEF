# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from datetime import timezone, datetime

from common.backup_restore_service.restore import Restore
from common.constants.base_constants import CommonConstants
from common.database_monitor import DatabaseMonitorBase
from common.log.logger import run_log
from fd_msg_process.midware_proc import MidwareProc


class RedfishDatabaseMonitor(DatabaseMonitorBase):
    """ redfish进程数据库监控 """

    def __init__(self, db_path: str = CommonConstants.REDFISH_EDGE_DB_FILE_PATH,
                 backup_dir: str = CommonConstants.REDFISH_BACKUP_DIR):
        super().__init__(db_path, backup_dir)
        # 是否发送清除告警到FD
        self.should_clear = False

    def report_database_alarm_to_fd(self, notify_type: str):
        """
        上报数据库异常告警到fd
        :param notify_type:是告警还是清除
        :return:
        """
        run_log.info("Report %s database alarm, notification type is %s.", self.DB_USER, notify_type)
        # 获取事件产生的时间
        alarm_time = datetime.now(tz=timezone.utc).replace(microsecond=0, tzinfo=timezone.utc).isoformat()
        payload_publish = {
            "alarm": [
                {
                    "type": "alarm",
                    "alarmId": "0x0013100E",
                    "alarmName": "MindXOM database abnormal",
                    "resource": "database",
                    "perceivedSeverity": "MAJOR",
                    "timestamp": alarm_time,
                    "notificationType": notify_type,
                    "detailedInformation": "This alarm is generated when the MindXOM database is malformed properly.",
                    "suggestion": "Log in to the device and check whether the database is malformed properly",
                    "reason": "The MindXOM database malformed properly",
                    "impact": "The MindXOM database may be unavailable",
                }
            ]
        }
        MidwareProc.report_event_to_fd(payload_publish)
        run_log.info("Report database alarm to FD done.")

    def monitor_database_status(self):
        check_ret = self.check_database_is_valid()
        if not check_ret:
            run_log.error(check_ret.error)
            self.report_database_alarm_to_fd("alarm")
            self.should_clear = True
            # 恢复数据库
            Restore(self.backup_dir, self.db_path).entry()
            return

        if self.should_clear:
            run_log.info("%s, should report clear database alarm to FD.", check_ret.error)
            self.report_database_alarm_to_fd("clear")
            self.should_clear = False
