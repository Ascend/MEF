# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

import os
from typing import NoReturn

from common.backup_restore_service.backup_restore_base import BackupRestoreBase
from common.backup_restore_service.backup_restore_base import BackupRestoreError
from common.file_utils import FileCheck
from common.log.logger import run_log


class Backup(BackupRestoreBase):

    @property
    def mode(self):
        return {".db": "600", ".json": "644"}.get(self.file_ext, "640")

    def backup_process(self):
        # 源文件不存在时直接返回，结束此次备份，此处不记日志，防止刷屏
        if not os.path.exists(self.source_file):
            return

        self.check_source_dir()
        self.check_source_file()

        # 备份文件有效性和源文件sha256一致性检查
        if not self.check_backup_file() or not self.check_sha256_consistent(self.source_file):
            self.check_backup_dir()
            self.delete_backup_file()
            self.copy_file(self.source_file, self.backup_file)
            self.save_source_file_sha256()
            run_log.info("Backup %s to %s success.", self.source_file, self.backup_file)

    def check_source_file(self) -> NoReturn:
        check_ret = FileCheck.check_input_path_valid(self.source_file)
        if not check_ret:
            raise BackupRestoreError(f"Check {self.source_file} is invalid, {check_ret.error}.")

        check_ret = FileCheck.check_path_mode_owner_group(self.source_file, self.mode, self.user, self.group)
        if not check_ret:
            raise BackupRestoreError(f"Check {self.source_file} is invalid, {check_ret.error}.")

        if self.is_db:
            # 检查工作数据库文件是否可用
            check_ret = self.check_database_available(self.source_file)
            if not check_ret:
                raise BackupRestoreError(check_ret.error)

    def entry(self):
        try:
            self.backup_process()
        except BackupRestoreError as err:
            run_log.error("Execute backup failed, %s", err)
        except Exception as err:
            run_log.error("Execute backup failed, %s", err)

    def save_source_file_sha256(self) -> NoReturn:
        flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
        source_file_sha256 = self.get_file_sha256(self.source_file)
        with os.fdopen(os.open(self.backup_sha256_file, flags, 0o600), "w") as fd_write:
            fd_write.write(source_file_sha256)
