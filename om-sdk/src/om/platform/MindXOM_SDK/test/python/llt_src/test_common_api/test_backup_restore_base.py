# -*- coding: utf-8 -*-
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
import os
import pwd
from collections import namedtuple
from unittest.mock import mock_open

import mock
import pytest
from pytest_mock import MockerFixture

from common.backup_restore_service.backup_restore_base import BackupRestoreError, BackupRestoreBase
from common.file_utils import FileCheck
from common.file_utils import FileCopy
from common.utils.app_common_method import AppCommonMethod
from common.utils.result_base import Result
from redfish_config import RedfishBackupRestoreCfg as BackupRestoreCfg

CheckSourceFileCase = namedtuple("CheckSourceFileCase", "check_path_is_exist_and_valid, check_path_mode_owner_group")
EntryCase = namedtuple("EntryCase", "backup_process")
CopyFile = namedtuple("CopyFile", ["user", "group", "backup_dir", "BACKUP_DIR_MODE", "backup_file",
                                   "backup_sha256_file", "source_dir", "BACKUP_FILE_MAX_SIZE_BYTES", "get_file_sha256"])
BackupRestoreBaseOne = CopyFile(*(pwd.getpwuid(os.getuid()).pw_name, grp.getgrgid(os.getgid()).gr_name,
                                  BackupRestoreCfg.BACKUP_FILES[0], "700", BackupRestoreCfg.BACKUP_FILES[0],
                                  BackupRestoreCfg.BACKUP_FILES[0], "/", 300 * 1024 * 1024, "test"))
CheckBackupDirCase = namedtuple("CheckBackupDirCase", "check_path_is_exist_and_valid, check_path_mode_owner_group")
CheckBackupFileCase = namedtuple("CheckBackupFileCase",
                                 "expected, check_path_is_exist_and_valid, "
                                 "check_path_mode_owner_group, is_db, check_sha256_consistent")
CheckSourceDirCase = namedtuple("CheckSourceDirCase", "check_is_link, check_path_mode_owner_group")
CheckSha256ConsistentCase = namedtuple("CheckSha256ConsistentCase", "expected, check_path_is_exist_and_valid, getsize")
GetFileSha256Case = namedtuple("GetFileSha256Case", "check_path_is_exist_and_valid, getsize")


class TestUtils:
    user = pwd.getpwuid(os.getuid()).pw_name
    group = grp.getgrgid(os.getgid()).gr_name
    backup_dir = BackupRestoreCfg.BACKUP_FILES[0]
    BACKUP_DIR_MODE = "700"
    backup_file = BackupRestoreCfg.BACKUP_FILES[0]
    backup_sha256_file = BackupRestoreCfg.BACKUP_FILES[0]
    BACKUP_FILE_MODE = "600"

    def __init__(self, model):
        self.is_db = model.is_db
        self.check_sha256_consistent2 = model.check_sha256_consistent

    @staticmethod
    def check_database_available(backup_file):
        return Result(False)

    def check_sha256_consistent1(self, backup_file):
        return self.check_sha256_consistent2

    def check_sha256_consistent(self, backup_file):
        return self.check_sha256_consistent1

    def get_file_sha256(self):
        return self.get_file_sha2561

    @staticmethod
    def get_file_sha2561():
        return "test"


class TestBbackupRestoreBase:
    use_cases = {
        "test_check_backup_dir": {
            "first": (Result(False), Result(False), ),
            "second": (Result(True), Result(False),),
        },
        "test_check_backup_file": {
            "first": (False, Result(False), Result(False), True, False),
            "second": (False, Result(True), Result(False), True, False),
            "third": (False, Result(True), Result(True), True, False),
            "fifth": (True, Result(True), Result(True), False, True),
        },
        "test_check_backup_file1": {
            "forth": (False, Result(True), Result(True), False, False),
        },
        "test_check_source_dir": {
            "first": (Result(False), Result(True)),
            "second": (Result(True), Result(False)),
        },
        "test_check_sha256_consistent": {
            "first": (False, Result(False), 300 * 1024 * 1024),
            "second": (False, Result(True), 300 * 1024 * 1024 + 1),
            "third": (False, Result(True), 300 * 1024 * 1024),
        },
        "test_get_file_sha256": {
            "first": (Result(False), 300 * 1024 * 1024 + 1),
            "second": (Result(True), 300 * 1024 * 1024 + 1),
        },
    }

    def test_check_database_available(self, mocker: MockerFixture):
        mocker.patch.object(AppCommonMethod, "check_database_available").return_value = True
        assert BackupRestoreBase.check_database_available(BackupRestoreCfg.BACKUP_FILES[0]) is True

    def test_entry(self):
        BackupRestoreBase.entry(BackupRestoreBase)

    def test_copy_file(self, mocker: MockerFixture):
        mocker.patch.object(FileCopy, "copy_file", return_value=Result(False))
        with pytest.raises(BackupRestoreError):
            assert not BackupRestoreBase.copy_file(BackupRestoreBaseOne, BackupRestoreCfg.BACKUP_FILES[0],
                                                   BackupRestoreCfg.BACKUP_FILES[0])

    def test_check_backup_dir(self, mocker: MockerFixture, model: CheckBackupDirCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid",
                            return_value=model.check_path_is_exist_and_valid)
        mocker.patch.object(FileCheck, "check_path_mode_owner_group", return_value=model.check_path_mode_owner_group)
        with pytest.raises(BackupRestoreError):
            assert not BackupRestoreBase.check_backup_dir(BackupRestoreBaseOne)

    def test_check_backup_file(self, mocker: MockerFixture, model: CheckBackupFileCase):
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid",
                            return_value=model.check_path_is_exist_and_valid)
        mocker.patch.object(FileCheck, "check_path_mode_owner_group", return_value=model.check_path_mode_owner_group)
        ret = BackupRestoreBase.check_backup_file(TestUtils(model))
        assert model.expected == ret

    def test_check_backup_file1(self, mocker: MockerFixture, model: CheckBackupFileCase):
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(True))
        mocker.patch.object(FileCheck, "check_path_mode_owner_group", return_value=Result(True))
        CopyFile1 = namedtuple("CopyFile1",
                              ["user", "group", "backup_dir", "BACKUP_DIR_MODE", "backup_file",
                               "backup_sha256_file", "is_db", "BACKUP_FILE_MODE", "check_sha256_consistent"])
        backup_restore_base_two = CopyFile1(*(pwd.getpwuid(os.getuid()).pw_name, grp.getgrgid(os.getgid()).gr_name,
                                             BackupRestoreCfg.BACKUP_FILES[0], "700", BackupRestoreCfg.BACKUP_FILES[0],
                                             BackupRestoreCfg.BACKUP_FILES[0], False, "600",
                                             TestUtils.check_database_available))
        ret = BackupRestoreBase.check_backup_file(backup_restore_base_two)
        assert model.expected == ret

    def test_check_source_dir(self, mocker: MockerFixture, model: CheckSourceDirCase):
        mocker.patch.object(FileCheck, "check_is_link", return_value=model.check_is_link)
        mocker.patch.object(FileCheck, "check_path_mode_owner_group", return_value=model.check_path_mode_owner_group)
        with pytest.raises(BackupRestoreError):
            assert not BackupRestoreBase.check_source_dir(BackupRestoreBaseOne)

    def test_check_sha256_consistent(self, mocker: MockerFixture, model: CheckSha256ConsistentCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid",
                            return_value=model.check_path_is_exist_and_valid)
        CopyFile1 = namedtuple("CopyFile1",
                              ["user", "group", "backup_dir", "BACKUP_DIR_MODE", "backup_file", "backup_sha256_file",
                               "is_db", "BACKUP_FILE_MODE", "check_sha256_consistent", "BACKUP_FILE_MAX_SIZE_BYTES",
                               "get_file_sha256"])
        backup_restore_base_two = CopyFile1(*(pwd.getpwuid(os.getuid()).pw_name, grp.getgrgid(os.getgid()).gr_name,
                                             BackupRestoreCfg.BACKUP_FILES[0], "700", BackupRestoreCfg.BACKUP_FILES[0],
                                             BackupRestoreCfg.BACKUP_FILES[0], False, "600",
                                             TestUtils.check_database_available, 300 * 1024 * 1024,
                                             TestUtils.check_database_available))
        mocker.patch("os.path.getsize", return_value=model.getsize)
        mocker.patch("builtins.open", mock_open(read_data="test"))
        ret = BackupRestoreBase.check_sha256_consistent(backup_restore_base_two, "/")
        assert model.expected == ret

    def test_delete_backup_file(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch("os.remove")
        assert not BackupRestoreBase.delete_backup_file(BackupRestoreBaseOne)

    def test_get_file_sha256(self, mocker: MockerFixture, model: GetFileSha256Case):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid",
                            return_value=model.check_path_is_exist_and_valid)
        mocker.patch("os.path.getsize", return_value=model.getsize)
        mocker.patch("builtins.open", mock_open(read_data="test".encode('utf-8')))
        with pytest.raises(BackupRestoreError):
            assert not BackupRestoreBase.get_file_sha256(BackupRestoreBaseOne, "")

    @mock.patch.object(FileCheck, 'check_path_is_exist_and_valid', mock.Mock(return_value=True))
    def test_get_file_sha256_noraml(self, mocker: MockerFixture):
        mocker.patch("os.path.getsize", return_value=300 * 1024 * 1024)
        mocker.patch("builtins.open", mock_open(read_data="test".encode('utf-8')))
        assert BackupRestoreBase.get_file_sha256(BackupRestoreBaseOne, "")
