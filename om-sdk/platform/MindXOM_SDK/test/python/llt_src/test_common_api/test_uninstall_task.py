# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from collections import namedtuple
from unittest.mock import patch

import mock
import pytest
from pytest_mock import MockerFixture

from common.file_utils import FileAttribute
from common.file_utils import FileCheck
from common.file_utils import FileUtils
from common.utils.exec_cmd import ExecCmd
from common.utils.result_base import Result
from uninstall.uninstall_task import UninstallError
from uninstall.uninstall_task import UnsetImmutableAttr, RemoveService, ClearConfig, uninstall
from utils import OperationRetCode


UninstallCase = namedtuple("UninstallCase", "expect run")


class TestUtils:
    name = ""


class TestunInstallTask:
    use_cases = {
        "test_uninstall": {
            "normal": (OperationRetCode.SUCCESS_OPERATION, TestUtils),
            "UninstallError": (OperationRetCode.FAILED_OPERATION, UninstallError("")),
        },

    }

    def test_check_work_dir(self, mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(False))
        with pytest.raises(UninstallError):
            UnsetImmutableAttr._check_work_dir()

    def test_unset_immutable_attr(self, mocker: MockerFixture):
        mocker.patch.object(FileAttribute, "set_path_immutable_attr", return_value=Result(False))
        with pytest.raises(UninstallError):
            UnsetImmutableAttr._unset_immutable_attr()

    def test_steps(self):
        installer = UnsetImmutableAttr()
        tasks = installer.steps()
        assert len(list(tasks)) == 2

    def test_rf_session_id(self, mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=Result(False))
        assert not RemoveService._daemon_reload()

    def test_steps1(self):
        tasks = RemoveService.steps(RemoveService())
        assert len(list(tasks)) == 4

    def test_stop_service(self, mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=Result(False))
        with pytest.raises(UninstallError):
            RemoveService._stop_service(RemoveService())

    @patch("os.path.join", mock.Mock(return_value=1))
    def test_remove_service_link(self, mocker: MockerFixture):
        mocker.patch.object(FileUtils, "delete_file_or_link").side_effect = Exception()
        assert not RemoveService._remove_service_link(RemoveService())

    @patch("os.path.join", mock.Mock(return_value=[1, ]))
    def test_remove_services_and_scripts(self, mocker: MockerFixture):
        mocker.patch.object(FileUtils, "delete_file_or_link").side_effect = Exception()
        assert not RemoveService._remove_services_and_scripts(RemoveService())

    def test_steps2(self):
        tasks = ClearConfig.steps(ClearConfig())
        assert len(list(tasks)) == 3

    def test_ies_config_files(self):
        tasks = ClearConfig._ies_config_files(ClearConfig())
        assert len(list(tasks)) == 8

    def test_remove_config_dir(self, mocker: MockerFixture):
        mocker.patch.object(FileUtils, "delete_dir_content", return_value=Result(False))
        assert not ClearConfig._remove_config_dir(ClearConfig())

    def test_remove_config_files(self, mocker: MockerFixture):
        mocker.patch.object(FileUtils, "delete_file_or_link").side_effect = Exception()
        assert not ClearConfig._remove_config_files(ClearConfig())

    def test_uninstall(self, mocker: MockerFixture, model: UninstallCase):
        mocker.patch("utils.get_login_user", return_value=("root1", True))
        mocker.patch.object(UnsetImmutableAttr, "run", side_effect=model.run)
        mocker.patch.object(RemoveService, "run", side_effect=model.run)
        mocker.patch.object(ClearConfig, "run", side_effect=model.run)
        ret = uninstall()
        assert ret == model.expect
