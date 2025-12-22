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
from collections import namedtuple

import mock
import pytest
from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.file_utils import FileCreate
from common.file_utils import FilePermission as Chmod
from common.file_utils import FileUtils
from common.utils.result_base import Result
from log_tar_save import DumpLogError
from log_tar_save import check_path
from log_tar_save import del_old_log_storage
from log_tar_save import main
from log_tar_save import tar_save_log

CheckPathCase = namedtuple("CheckPathCase", "expected, check_input_path_valid")
DelOldLogStorageCase = namedtuple("DelOldLogStorageCase", "expected, exists, create_dir, delete_file_or_link")
TarSaveLogCase = namedtuple("TarSaveLogCase", "open, set_path_permission")
MainCase = namedtuple("MainCase", "")


class TestUtils:
    log_base_name = "name"
    max_num = 1
    dest_dir = "/home/log/plog/manager"
    src_dir = "/home/failedcase_log"


class TestInstall:
    use_cases = {
        "test_check_path": {
            "normal": ("/home", Result(result=True)),
        },
        "test_check_path_raise": {
            "normal": (None, Result(result=False)),
        },
        "test_del_old_log_storage": {
            "normal": (None, True, Result(result=True), Result(result=True)),
            "return_none": (None, False, Result(result=True), Result(result=True)),
        },
        "test_del_old_log_storage_raise": {
            "raise0": (None, False, Result(result=False), Exception()),
        },
        "test_tar_save_log": {
            "normal": (None, Result(result=True)),
            "return_none": (None, Result(result=True)),
        },
        "test_tar_save_log_raise": {
            "raise0": (Exception(), Result(result=True)),
            "raise1": (None, Result(result=False)),
        },
        "test_main": {
        },

    }

    def test_check_path(self, mocker: MockerFixture, model: CheckPathCase):
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=model.check_input_path_valid)
        ret = check_path("/home")
        assert model.expected == ret

    def test_check_path_raise(self, mocker: MockerFixture, model: CheckPathCase):
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=model.check_input_path_valid)
        with pytest.raises(DumpLogError):
            check_path("/home")

    def test_del_old_log_storage(self, mocker: MockerFixture, model: DelOldLogStorageCase):
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        ret = del_old_log_storage("/home/data/config/redfish/", 0, "redfish_run")
        assert model.expected == ret

    @mock.patch("os.path.getmtime", mock.Mock(return_value=1686313081.14))
    def test_del_old_log_storage_raise(self, mocker: MockerFixture, model: DelOldLogStorageCase):
        mocker.patch("common.log.logger.run_log.info", return_value=None)
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        mocker.patch.object(FileUtils, "delete_file_or_link").side_effect = model.delete_file_or_link
        with pytest.raises(DumpLogError):
            del_old_log_storage("/home/data/config/redfish/", 0, "redfish_run")

    def test_tar_save_log(self, mocker: MockerFixture, model: TarSaveLogCase):
        mocker.patch("tarfile.open", read_data=model.open)
        mocker.patch.object(Chmod, "set_path_permission", return_value=model.set_path_permission)
        assert not tar_save_log(TestUtils.dest_dir, TestUtils.src_dir, "manager_run")

    def test_tar_save_log_raise(self, mocker: MockerFixture, model: TarSaveLogCase):
        mocker.patch.object(Chmod, "set_path_permission", return_value=model.set_path_permission)
        mocker.patch("tarfile.open").side_effect = model.open
        with pytest.raises(DumpLogError):
            tar_save_log(TestUtils.dest_dir, TestUtils.src_dir, "manager_run")

    def test_main(self, mocker: MockerFixture):
        mocker.patch("log_tar_save.parse_input_param", return_value=TestUtils)
        mocker.patch("common.log.logger.run_log.info", side_effect=[None, None])
        mocker.patch("log_tar_save.del_old_log_storage", return_value=None)
        mocker.patch("log_tar_save.tar_save_log", return_value=None)
        assert not main()
