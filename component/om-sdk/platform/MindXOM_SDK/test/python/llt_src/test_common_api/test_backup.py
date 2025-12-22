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
import os
from collections import namedtuple

import pytest
from pytest_mock import MockerFixture

from common.backup_restore_service.backup import Backup
from common.backup_restore_service.backup_restore_base import BackupRestoreError
from common.file_utils import FileCheck
from common.utils.result_base import Result
from redfish_config import RedfishBackupRestoreCfg as BackupRestoreCfg

CheckSourceFileCase = namedtuple("CheckSourceFileCase", "exists, check_input_path_valid, check_path_mode_owner_group")
EntryCase = namedtuple("EntryCase", "backup_process")


class TestBackup:
    use_cases = {
        "test_check_source_file": {
            "first": (False, Result(False), Result(False), ),
            "second": (False, Result(True), Result(False),),
            "third": (False, Result(True), Result(True),),
        },
        "test_entry": {
            "BackupRestoreError": (BackupRestoreError(""), ),
            "Exception": (Exception(), ),
        },
    }

    def test_backup_process(self, mocker: MockerFixture):
        mocker.patch.object(Backup, "check_source_dir").return_value = True
        mocker.patch.object(Backup, "check_source_file").return_value = True
        mocker.patch.object(Backup, "check_backup_dir").return_value = True
        mocker.patch.object(Backup, "delete_backup_file").return_value = True
        mocker.patch.object(Backup, "copy_file").return_value = True
        mocker.patch.object(Backup, "save_source_file_sha256").return_value = True
        assert not Backup.backup_process(Backup(BackupRestoreCfg.BACKUP_DIR, BackupRestoreCfg.BACKUP_FILES[0]))

    def test_check_source_file(self, mocker: MockerFixture, model: CheckSourceFileCase):
        mocker.patch.object(os.path, "exists", return_value=model.exists)
        mocker.patch.object(FileCheck, "check_input_path_valid",
                            return_value=model.check_input_path_valid)
        mocker.patch.object(FileCheck, "check_path_mode_owner_group", return_value=model.check_path_mode_owner_group)
        with pytest.raises(BackupRestoreError):
            Backup.check_source_file(Backup(BackupRestoreCfg.BACKUP_DIR, BackupRestoreCfg.BACKUP_FILES[0]))

    def test_entry(self, mocker: MockerFixture, model: EntryCase):
        mocker.patch.object(Backup, "backup_process").side_effect = model.backup_process
        assert not Backup.entry(Backup(BackupRestoreCfg.BACKUP_DIR, BackupRestoreCfg.BACKUP_FILES[0]))

    def test_save_source_file_sha256(self, mocker: MockerFixture):
        mocker.patch.object(Backup, "get_file_sha256").return_value = ""
        mocker.patch("os.open")
        mocker.patch("os.fdopen").return_value.__enter__.return_value.write.side_effect = "abc"
        assert not Backup.save_source_file_sha256(Backup(BackupRestoreCfg.BACKUP_DIR, BackupRestoreCfg.BACKUP_FILES[0]))
