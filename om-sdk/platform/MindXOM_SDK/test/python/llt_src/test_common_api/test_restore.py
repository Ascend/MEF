# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from collections import namedtuple

from pytest_mock import MockerFixture

from common.backup_restore_service.backup_restore_base import BackupRestoreError
from common.backup_restore_service.restore import Restore
from common.utils.result_base import Result

EntryCase = namedtuple("EntryCase", "err")


class TestUtils:
    @staticmethod
    def check_database_available():
        return Result(False)

    @staticmethod
    def check_database_available1():
        return Result(True)

    @staticmethod
    def check_database_available2(backup_file):
        return Result(True)

    @staticmethod
    def check_database_available3(backup_file):
        return Result(False)

    @staticmethod
    def check_database_available4(backup_file, backup_file1, mode):
        return Result(False)


class TestRestore:
    use_cases = {
        "test_entry": {
            "first": (Exception(), ),
            "second": (BackupRestoreError(""),),
        },
    }

    def test_entry(self, mocker: MockerFixture, model: EntryCase):
        mocker.patch.object(Restore, "restore_process").side_effect = model.err
        assert not Restore.entry(Restore)

    def test_restore_process_first(self):
        CopyFile1 = namedtuple("CopyFile1", ["exists", "backup_dir", "check_backup_dir", "check_backup_file",
                                             "delete_backup_file"])
        backup_restore_base_two = CopyFile1(*(True, "", TestUtils.check_database_available1,
                                              TestUtils.check_database_available, TestUtils.check_database_available))
        assert not Restore.restore_process(backup_restore_base_two)

    def test_restore_process_second(self):
        CopyFile1 = namedtuple("CopyFile1", ["exists", "backup_dir", "check_backup_dir", "check_backup_file",
                                             "check_source_dir", "is_db", "check_database_available", "source_file"])
        backup_restore_base_two = CopyFile1(*(True, "", TestUtils.check_database_available1,
                                              TestUtils.check_database_available1, TestUtils.check_database_available,
                                              TestUtils.check_database_available1, TestUtils.check_database_available2,
                                              TestUtils.check_database_available1))
        assert not Restore.restore_process(backup_restore_base_two)

    def test_restore_process_third(self):
        CopyFile1 = namedtuple("CopyFile1", ["exists", "backup_dir", "check_backup_dir", "check_backup_file",
                                             "check_source_dir", "is_db", "check_database_available",
                                             "check_sha256_consistent", "source_file"])
        backup_restore_base_two = CopyFile1(*(True, "", TestUtils.check_database_available1,
                                              TestUtils.check_database_available1, TestUtils.check_database_available,
                                              TestUtils.check_database_available, TestUtils.check_database_available3,
                                              TestUtils.check_database_available2, TestUtils.check_database_available1))
        assert not Restore.restore_process(backup_restore_base_two)

    def test_restore_process_forth(self):
        CopyFile1 = namedtuple("CopyFile1", ["exists", "backup_dir", "check_backup_dir", "check_backup_file",
                                             "check_source_dir", "is_db", "check_database_available",
                                             "check_sha256_consistent", "source_file", "copy_file", "backup_file",
                                             "mode"])
        backup_restore_base_two = CopyFile1(*(True, "", TestUtils.check_database_available1,
                                              TestUtils.check_database_available1, TestUtils.check_database_available,
                                              TestUtils.check_database_available, TestUtils.check_database_available3,
                                              TestUtils.check_database_available3, TestUtils.check_database_available1,
                                              TestUtils.check_database_available4, TestUtils.check_database_available1,
                                              0o600))
        assert not Restore.restore_process(backup_restore_base_two)
