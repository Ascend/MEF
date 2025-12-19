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
import grp
import inspect
import os
import pwd
from abc import ABC
from abc import abstractmethod
from hashlib import sha256
from pathlib import Path
from typing import NoReturn

from common.file_utils import FileCheck
from common.file_utils import FileCopy
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.result_base import Result


class BackupRestoreError(Exception):
    def __init__(self, err_msg: str):
        self.err_msg = err_msg
        # 将触发异常的调用点添加到错误信息中，统一打印错误时，明确触发错误的信息，便于定位。
        info: inspect.FrameInfo = inspect.stack()[1]
        msg = f"{info.function}({Path(info.filename).name}:{info.lineno})-{err_msg}"
        super().__init__(msg)


class BackupRestoreBase(ABC):
    """ 备份和恢复基类 """

    BACKUP_DIR_MODE = "700"
    BACKUP_FILE_MODE = "600"
    BACKUP_FILE_MAX_SIZE_BYTES = 300 * 1024 * 1024

    def __init__(self, backup_dir: str, file_path: str):
        self.backup_dir = backup_dir
        self.source_file = file_path
        self.backup_file = os.path.join(self.backup_dir, os.path.basename(self.source_file))
        self.source_dir = os.path.dirname(self.source_file)
        self.backup_sha256_file = f"{self.backup_file}.sha256"
        self.file_ext = os.path.splitext(self.source_file)[-1]
        self.user = pwd.getpwuid(os.getuid()).pw_name
        self.group = grp.getgrgid(os.getgid()).gr_name
        self.is_db = self.file_ext == ".db"

    @staticmethod
    def check_database_available(db_path: str) -> Result:
        return AppCommonMethod.check_database_available(db_path)

    @abstractmethod
    def entry(self):
        pass

    def copy_file(self, src_file: str, dst_file: str, mode=0o600) -> NoReturn:
        # 针对所有备份文件都只给600权限, 防止被篡改; 恢复文件权限与原文件保持一致
        ret = FileCopy.copy_file(src_file, dst_file, mode, self.user, self.group)
        if not ret:
            raise BackupRestoreError(f"Copy file failed reason is {ret.error}.")

    def check_backup_dir(self) -> NoReturn:
        check_ret = FileCheck.check_path_is_exist_and_valid(self.backup_dir)
        if not check_ret:
            raise BackupRestoreError(f"Check {self.backup_dir} is invalid, {check_ret.error}.")

        check_ret = FileCheck.check_path_mode_owner_group(self.backup_dir, self.BACKUP_DIR_MODE, self.user, self.group)
        if not check_ret:
            raise BackupRestoreError(f"Check {self.backup_dir} is invalid, {check_ret.error}.")

    def check_backup_file(self) -> bool:
        for file in self.backup_file, self.backup_sha256_file:
            # 备份文件或者sha256值文件不存在时直接返回，认为备份文件被破坏，此处拆分是防止日志刷屏
            if not os.path.exists(file):
                return False

            check_ret = FileCheck.check_input_path_valid(file)
            if not check_ret:
                run_log.error("Check %s %s.", file, check_ret.error)
                return False

            check_ret = FileCheck.check_path_mode_owner_group(file, self.BACKUP_FILE_MODE, self.user, self.group)
            if not check_ret:
                run_log.error("Check %s %s.", file, check_ret.error)
                return False

        if self.is_db:
            # 检查备份数据库文件是否可用
            check_ret = self.check_database_available(self.backup_file)
            if not check_ret:
                run_log.error(check_ret.error)
                return False

        # 检查备份文件sha256和备份sha256值是否一致
        if not self.check_sha256_consistent(self.backup_file):
            run_log.error("%s sha256 value is different from backup sha256 value.", self.backup_file)
            return False

        return True

    def check_source_dir(self) -> NoReturn:
        check_ret = FileCheck.check_is_link(self.source_dir)
        if not check_ret:
            raise BackupRestoreError(f"Check {self.source_dir} is invalid, {check_ret.error}.")

        check_ret = FileCheck.check_path_mode_owner_group(self.source_dir, user=self.user, group=self.group)
        if not check_ret:
            raise BackupRestoreError(f"Check {self.source_dir} is invalid, {check_ret.error}.")

    def check_sha256_consistent(self, file_path: str) -> bool:
        check_ret = FileCheck.check_path_is_exist_and_valid(file_path)
        if not check_ret:
            run_log.error("Check %s %s.", file_path, check_ret.error)
            return False

        if os.path.getsize(self.backup_sha256_file) > self.BACKUP_FILE_MAX_SIZE_BYTES:
            run_log.error("Check %s is too large.", self.backup_sha256_file)
            return False

        with open(self.backup_sha256_file, "r") as fd_read:
            return self.get_file_sha256(file_path) == fd_read.read()

    def delete_backup_file(self) -> NoReturn:
        for file in self.backup_file, self.backup_sha256_file:
            if os.path.exists(file):
                run_log.info("remove %s.", file)
                os.remove(file)

    def get_file_sha256(self, file_path: str) -> str:
        check_ret = FileCheck.check_path_is_exist_and_valid(file_path)
        if not check_ret:
            raise BackupRestoreError(f"Check {file_path} {check_ret.error}.")

        if os.path.getsize(file_path) > self.BACKUP_FILE_MAX_SIZE_BYTES:
            raise BackupRestoreError(f"{file_path} is too large.")

        with open(file_path, "rb") as fd_read:
            return sha256(fd_read.read()).hexdigest()

