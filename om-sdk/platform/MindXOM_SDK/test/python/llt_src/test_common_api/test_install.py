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
import builtins
import os
from collections import namedtuple
from unittest.mock import mock_open

import pytest
from pytest_mock import MockerFixture

from common.constants.upgrade_constants import OMUpgradeConstants
from common.file_utils import FileCopy
from common.file_utils import FileCreate
from common.file_utils import FilePermission
from common.file_utils import FilePermission as Chown
from common.utils.exec_cmd import ExecCmd
from copy_om_sys_file import CopySysFileOperator
from common.utils.result_base import Result
from install import OMInstaller, InstallError

GetOsNameAndVersionIdCase = namedtuple("GetOsNameAndVersionIdCase", "expected, ")
InitDatabaseCase = namedtuple("InitDatabaseCase", "exists, create_dir, set_path_owner_group")
CopyOmConfigFileCase = namedtuple("CopyOmConfigFileCase", "create_dir, set_path_permission, copy_file, exists")
EffectOmCase = namedtuple("EffectOmCase", "copy_sys_file_to_system_dir, exec_cmd")
CopyFilesToUpgradeDirCase = namedtuple("CopyFilesToUpgradeDirCase", "create_dir, isdir, copytree, copyfile")
ReplaceOsCmdConfCase = namedtuple("ReplaceOsCmdConfCase",
                                  "exists, create_dir, judge_os_cmd_file, copy_file, set_path_permission")
InstallCase = namedtuple("InstallCase", "expected, tasks")


class TestInstall:
    func_name = "print"
    func = getattr(builtins, func_name)
    use_cases = {
        "test_get_os_name_and_version_id": {
            "normal": (['Ubuntu', '20.04'], ),
        },
        "test_judge_os_cmd_file_not_ret": {
            "normal": (),
        },
        "test_judge_os_cmd_file_exception": {
            "normal": (),
        },
        "test_create_flag_file": {
            "normal": (),
        },
        "test_init_database_exception": {
            "first": (Result(result=False), Result(result=False), Result(result=True)),
            "second": (Result(result=False), Result(result=True), Result(result=False)),
            "last": (Result(result=True), Result(result=True), Result(result=True)),
        },
        "test_copy_om_config_file_exception": {
            "first": ([Result(result=False), Result(result=True), ], Result(result=True), [Result(result=True),
                                                                                           Result(result=False)], True),
            "second": ([Result(result=True), Result(result=True), ], Result(result=False),
                       [Result(result=True), Result(result=False)], True),
            "third": ([Result(result=True), Result(result=False), ], Result(result=True),
                      [Result(result=True), Result(result=False)], True),
            "fourth": ([Result(result=True), Result(result=True), ], Result(result=True),
                       [Result(result=False), Result(result=False)], True),
            "fifth": ([Result(result=True), Result(result=True), ], Result(result=True),
                      [Result(result=True), Result(result=False)], True),
            "last": ([Result(result=True), Result(result=True), ], Result(result=True),
                     [Result(result=True), Result(result=False)], False),
        },
        "test_effect_om_exception": {
            "first": (Exception(), [0, 1]),
            "second": (None, [1, 1]),
            "last": (None, [0, 1]),
        },
        "test_copy_files_to_upgrade_dir": {
            "normal": (Result(result=True), False, None, None),
            "normal1": (Result(result=True), True, None, None),
        },
        "test_copy_files_to_upgrade_dir_exception": {
            "first": (Result(result=False), Exception(), None, None),
            "second": (Result(result=True), Exception(), None, None),
        },
        "test_replace_os_cmd_conf": {
            "normal": (True, Result(result=True), True, Result(result=True), Result(result=True), ),
        },
        "test_replace_os_cmd_conf_exception": {
            "first": (False, Result(result=False), True, Result(result=True), Result(result=True),),
            "second": (True, Result(result=False), True, Result(result=False), Result(result=True),),
            "third": (True, Result(result=False), True, Result(result=True), Result(result=False),),
            "last": (True, Result(result=False), False, Result(result=True), Result(result=True),),
        },
        "test_install": {
            "normal": (0, [func, ]),
            "exception": (1, None),
        },

    }

    @staticmethod
    def replace_os_cmd_conf_basic(mocker, model):
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        mocker.patch.object(OMInstaller, "judge_os_cmd_file", return_value=model.judge_os_cmd_file)
        mocker.patch.object(FileCopy, "copy_file", return_value=model.copy_file)
        mocker.patch.object(FilePermission, "set_path_permission", return_value=model.set_path_permission)

    def test_judge_os_cmd_file_not_ret(self):
        with pytest.raises(InstallError):
            OMInstaller.judge_os_cmd_file(['EulerOS', '2.0'], "a")

    def test_judge_os_cmd_file_exception(self, mocker: MockerFixture):
        mocker.patch("builtins.open", mock_open(read_data="test"))
        with pytest.raises(Exception):
            OMInstaller.judge_os_cmd_file(
                ['EulerOS', '2.0'], os.path.join(os.path.join(OMInstaller.ROOT_DIR_PATH, "config"),
                                                 "os_cmd_eluros2.0.conf"))

    def test_create_flag_file_exception(self):
        with pytest.raises(InstallError):
            OMInstaller.create_flag_file(1)

    def test_create_flag_file(self):
        ret = OMInstaller.create_flag_file(OMUpgradeConstants.UPGRADE_FINISHED_FLAG)
        assert not ret

    def test_init_database_exception(self, mocker: MockerFixture, model: InitDatabaseCase):
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        mocker.patch.object(Chown, "set_path_owner_group", return_value=model.set_path_owner_group)
        with pytest.raises(InstallError):
            OMInstaller.init_database()

    def test_copy_om_config_file_exception(self, mocker: MockerFixture, model: CopyOmConfigFileCase):
        mocker.patch.object(FileCreate, "create_dir", side_effect=model.create_dir)
        mocker.patch.object(FilePermission, "set_path_permission", return_value=model.set_path_permission)
        mocker.patch.object(FileCopy, "copy_file", side_effect=model.copy_file)
        mocker.patch("os.path.exists", return_value=model.exists)
        with pytest.raises(InstallError):
            OMInstaller.copy_om_config_file()

    def test_effect_om_exception(self, mocker: MockerFixture, model: EffectOmCase):
        mocker.patch.object(CopySysFileOperator, "copy_sys_file_to_system_dir").side_effect = \
            model.copy_sys_file_to_system_dir
        mocker.patch.object(ExecCmd, "exec_cmd", side_effect=model.exec_cmd)
        with pytest.raises(InstallError):
            OMInstaller.effect_om(OMInstaller())

    def test_copy_files_to_upgrade_dir(self, mocker: MockerFixture, model: CopyFilesToUpgradeDirCase):
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        mocker.patch("os.path.isdir", return_value=model.isdir)
        mocker.patch("shutil.copytree", return_value=model.copytree)
        mocker.patch("shutil.copyfile", return_value=model.copyfile)
        assert not OMInstaller.copy_files_to_upgrade_dir(OMInstaller())

    def test_copy_files_to_upgrade_dir_exception(self, mocker: MockerFixture, model: CopyFilesToUpgradeDirCase):
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        mocker.patch("os.path.isdir").side_effect = model.isdir
        with pytest.raises(InstallError):
            OMInstaller.copy_files_to_upgrade_dir(OMInstaller())

    def test_replace_os_cmd_conf(self, mocker: MockerFixture, model: ReplaceOsCmdConfCase):
        self.replace_os_cmd_conf_basic(mocker, model)
        assert not OMInstaller.replace_os_cmd_conf(OMInstaller())

    def test_replace_os_cmd_conf_exception(self, mocker: MockerFixture, model: ReplaceOsCmdConfCase):
        self.replace_os_cmd_conf_basic(mocker, model)
        with pytest.raises(InstallError):
            OMInstaller.replace_os_cmd_conf(OMInstaller())

    def test_tasks(self):
        installer = OMInstaller()
        tasks = installer.tasks()
        assert len(list(tasks)) == 13

    def test_install(self, mocker: MockerFixture, model: InstallCase):
        mocker.patch.object(OMInstaller, "tasks", return_value=model.tasks)
        ret = OMInstaller.install(OMInstaller())
        assert model.expected == ret
