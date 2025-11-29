# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os
import sys
from subprocess import Popen, PIPE

from bin.monitor_config import MonitorBackupRestoreCfg
from common.backup_restore_service.backup import Backup
from common.backup_restore_service.restore import Restore
from common.file_utils import FileUtils
from common.log.logger import run_log


class BackupRestoreService:
    """ 备份恢复服务入口 """

    BACKUP_FILES = MonitorBackupRestoreCfg.BACKUP_FILES
    SOURCE_DIR = ("/home/data/config/monitor", "/home/data/ies")
    BACKUP_DIR = MonitorBackupRestoreCfg.BACKUP_DIR

    def __init__(self):
        self.file_names = [os.path.basename(file_path) for file_path in self.BACKUP_FILES]
        self.inotify_cmd = ("inotifywait", "-mq", "-e", "create,modify,delete") + self.SOURCE_DIR

    def backup(self):
        # 起一个inotifywait进程监控需要实时备份的文件
        with Popen(self.inotify_cmd, stdout=PIPE, stderr=PIPE, shell=False) as proc:
            while True:
                try:
                    self._inotify_result_process(proc)
                except Exception:
                    run_log.error("Inotify result process failed.")

    def restore(self):
        for file_path in self.BACKUP_FILES:
            Restore(self.BACKUP_DIR, file_path).entry()

    def _inotify_result_process(self, proc):
        # 事件格式: b'/home/data/ies/' b'CREATE' b'access_control.ini'
        result = proc.stdout.readline().decode("utf-8")
        if not result:
            return

        src_path, event_type, file_name, *_ = result.split()
        if file_name not in self.file_names:
            return

        run_log.info("Check file %s %s event.", file_name, event_type)
        # 创建和修改原文件之后需要执行备份
        source_file = os.path.join(src_path, file_name)
        if event_type in ("MODIFY", "CREATE"):
            Backup(self.BACKUP_DIR, source_file).entry()
        # 删除原文件之后需要删除备份的文件及其sha256值文件
        elif event_type == "DELETE":
            if source_file in MonitorBackupRestoreCfg.DELETE_BLACKLIST:
                return

            backup_file = os.path.join(self.BACKUP_DIR, file_name)
            for del_file in backup_file, f"{backup_file}.sha256":
                FileUtils.delete_file_or_link(del_file)
        else:
            run_log.warning("Invalid event type.")


def main():
    instance = BackupRestoreService()
    # 执行恢复检查
    instance.restore()
    # 执行备份检查，并启动inotifywait进程持续监听配置文件变化
    instance.backup()


if __name__ == "__main__":
    try:
        main()
    except Exception as err:
        run_log.error("Execute backup restore failed, %s", err)
        sys.exit(1)

    sys.exit(0)
