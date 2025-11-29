# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os

from common.log.logger import run_log
from common.backup_restore_service.backup_restore_base import BackupRestoreBase, BackupRestoreError


class Restore(BackupRestoreBase):

    @property
    def mode(self):
        return {".db": 0o600, ".json": 0o644}.get(self.file_ext, 0o640)

    def entry(self):
        try:
            self.restore_process()
        except BackupRestoreError as err:
            run_log.error("Execute restore failed, %s", err)
        except Exception as err:
            run_log.error("Execute restore failed, %s", err)

    def restore_process(self):
        self.check_backup_dir()
        if not self.check_backup_file():
            # 备份文件检查无效，删除备份文件，无需执行恢复操作
            self.delete_backup_file()
            run_log.info("Check Backup file is invalid. don't need to be restored")
            return

        self.check_source_dir()
        run_log.info("Start check whether %s needs to be restored.", self.source_file)
        if self.is_db:
            # 检查工作数据库文件是否可用，可用则不进行恢复操作
            check_ret = self.check_database_available(self.source_file)
            if check_ret:
                run_log.info("%s. don't need to be restored.", check_ret.error)
                return

            run_log.error("%s. need to be restored.", check_ret.error)

        # 检查工作文件和备份sha256值是否一致，一致则不进行恢复操作
        if self.check_sha256_consistent(self.source_file):
            run_log.info("%s sha256 values are consistent, don't need to be restored.", self.source_file)
            return

        self.copy_file(self.backup_file, self.source_file, self.mode)
        run_log.info("Restore %s to %s success.", self.backup_file, self.source_file)
