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
import os.path
from collections import namedtuple
from pathlib import Path

import pytest
from pytest_mock import MockerFixture

from common.constants.base_constants import CommonConstants
from common.constants.upgrade_constants import OMUpgradeConstants
from common.file_utils import FileCopy
from common.file_utils import FilePermission as Chmod
from common.file_utils import FilePermission as Chown
from common.file_utils import FileUtils
from common.utils.exec_cmd import ExecCmd
from common.utils.result_base import Result
from upgrade_om import OMUpgradeProcessor
from upgrade_om import UpgradeOMError
from upgrade_om import UpgradeResultCode


class TestUtils:
    module = 1
    processor_architecture = 1
    name = ""
    pw_uid = 1
    pw_gid = 1

    @staticmethod
    def as_posix():
        return False

    @staticmethod
    def is_symlink():
        return False

    @staticmethod
    def tasks():
        return False


class TestUtilsTrue:
    @staticmethod
    def exists():
        return True


class TestUpgradeOm:
    CheckXmlCase = namedtuple("CheckXmlCase", "xml_path")
    SetNginxIpCase = namedtuple("SetNginxIpCase", "check_path_is_exist_and_valid, main copy_file")
    CopyEncryptFileCase = namedtuple("CopyEncryptFileCase", "exists")
    CreateUdsCertCase = namedtuple("CreateUdsCertCase", "exists, copy_file, create_server_certs, exec_cmd_get_output")
    CopyCertCase = namedtuple("CopyCertCase", "exists, set_path_owner_group")
    SetSoftLinkCase = namedtuple("SetSoftLinkCase", "expected, src_soft_link")
    ChangeUpgradeFilesPermissionCase = namedtuple("ChangeUpgradeFilesPermissionCase",
                                                  "set_path_permission, set_path_owner_group")
    UpgradeCase = namedtuple("UpgradeCase", "expected")

    use_cases = {
        "test_set_valid_commands": {},
        "test_verify_upgrade_module_exception": {
            "first_raise": ("/home", ),
            "second_raise": ("/usr/local/mindx/MindXOM/version.xml", ),
        },
        "test_copy_encrypt_file": {
            "first_raise": ([False, False], ),
            "second_raise": ([True, False], ),
        },
        "test_create_uds_cert": {
            "first_raise": ([False, False], [Result(True), ], True, [[1, ], [1, ]]),
            "third_raise": ([True, False], [Result(True), ], False, [[1, ], [1, ]]),
            "fourth_raise": ([True, False], [Result(True), ], True, [[1, ], [1, ]]),
            "fifth_raise": ([True, False], [Result(True), ], True, [[0, ], [1, ]]),
            "sixth_raise": ([True, True], [Result(True), Result(False)], True, [[0, ], [1, ]]),
        },
        "test_copy_cert": {
            "first_raise": (False, Result(True)),
            "second_raise": (True, Result(False)),
            "third_raise": (True, Result(True)),
        },
        "test_change_upgrade_files_permission_process_exception": {
            "first": ([Result(False), ], [Result(False), ]),
            "second": ([True, ], [Result(False), ]),
            "fourth": ([True, True, True, True, True, True, True, True, True, True, True, Result(False)], [True, ]),
            "sixth": ([True, True, True, True, True, True, True,
                       True, True, True, True, True, Result(False)], [True, ]),
            "seventh": ([True, True, True, True, True, True, True, True, True, True, True, True,
                         True, True, True, True, True, True, True, True, True, True, True, True,
                         True, True, True, True, True, True, True, True],
                        [True, Result(False)]),
            "eighth": (
                [True, True, True, True, True, True, True, True, True, True, True, True,
                 True, True, True, True, True, True, True, True, True, True, True, True,
                 True, True, True, True, True, True, True, True], [True, True, Result(False)])
        },
        "test_upgrade_exception": {
            "Exception": (UpgradeResultCode.FAILED_OPERATION, ),
        },
        "test_upgrade": {
            "normal": (UpgradeResultCode.SUCCESS_OPERATION,),
        },
    }

    def test_set_valid_commands(self):
        assert not OMUpgradeProcessor.set_valid_commands()

    def test_verify_upgrade_module_exception(self, mocker: MockerFixture, model: CheckXmlCase):
        mocker.patch("common.utils.version_xml_file_manager.VersionXmlManager", return_value=None)
        mocker.patch.object(OMUpgradeConstants, "MODULE_NAME", return_value="1")
        with pytest.raises(Exception):
            OMUpgradeProcessor.verify_upgrade_module(model.xml_path, Exception())

    def test_verify_upgrade_module_third_exception(self, mocker: MockerFixture):
        mocker.patch("common.utils.version_xml_file_manager.VersionXmlManager", return_value=None)
        mocker.patch.object(OMUpgradeConstants, "PROCESSOR_ARCHITECTURE", return_value="1")
        with pytest.raises(Exception):
            OMUpgradeProcessor.verify_upgrade_module("/usr/local/mindx/MindXOM/version.xml", Exception())

    def test_copy_encrypt_file_final_raise(self):
        with pytest.raises(UpgradeOMError):
            OMUpgradeProcessor.copy_encrypt_file()

    def test_copy_encrypt_file(self, mocker: MockerFixture, model: CopyEncryptFileCase):
        mocker.patch("os.path.exists", side_effect=model.exists)
        mocker.patch.object(Chown, "set_path_owner_group", return_value=Result(True))
        with pytest.raises(StopIteration):
            OMUpgradeProcessor.copy_encrypt_file()

    def test_create_uds_cert_second_raise(self):
        with pytest.raises(UpgradeOMError):
            OMUpgradeProcessor._create_uds_cert()

    def test_create_uds_cert(self, mocker: MockerFixture, model: CreateUdsCertCase):
        mocker.patch("os.path.exists", side_effect=model.exists)
        mocker.patch.object(FileCopy, "copy_file", side_effect=model.copy_file)
        mocker.patch("create_server_certs.create_server_certs", return_value=model.create_server_certs)
        mocker.patch.object(ExecCmd, "exec_cmd_get_output", side_effect=model.exec_cmd_get_output)
        with pytest.raises(StopIteration):
            OMUpgradeProcessor._create_uds_cert()

    def test_copy_cert(self, mocker: MockerFixture, model: CopyCertCase):
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch.object(Chown, "set_path_owner_group", return_value=model.set_path_owner_group)
        mocker.patch.object(FileCopy, "copy_file", return_value=Result(False))
        with pytest.raises(UpgradeOMError):
            OMUpgradeProcessor._copy_cert(OMUpgradeConstants.NGINX_CONFIG_DIR,
                                          os.path.join(OMUpgradeConstants.CONFIG_HOME_PATH, "redfish"),
                                          CommonConstants.MINDXOM_USER)

    def test_whitebox_process(self):
        assert not OMUpgradeProcessor.whitebox_process()

    def test_whitebox_process_exception(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch.object(FileUtils, "delete_full_dir", return_value=None)
        with pytest.raises(UpgradeOMError):
            OMUpgradeProcessor.whitebox_process()

    def test_change_upgrade_files_permission_process_third_exception(self):
        with pytest.raises(UpgradeOMError):
            OMUpgradeProcessor.change_upgrade_files_permission(OMUpgradeProcessor())

    def test_change_upgrade_files_permission_process_exception(self, mocker: MockerFixture,
                                                               model: ChangeUpgradeFilesPermissionCase):
        mocker.patch.object(Chmod, "set_path_permission", side_effect=model.set_path_permission)
        mocker.patch.object(Chown, "set_path_owner_group", side_effect=model.set_path_owner_group)
        mocker.patch.object(OMUpgradeProcessor, "_change_software_mode_recursive", side_effect=[None, None, None, None])
        mocker.patch.object(Path, "rglob", return_value=[TestUtils, ])
        mocker.patch.object(FileUtils, "delete_file_or_link")
        with pytest.raises(UpgradeOMError):
            OMUpgradeProcessor.change_upgrade_files_permission(OMUpgradeProcessor())

    def test_update_cert(self, mocker: MockerFixture):
        mocker.patch.object(OMUpgradeProcessor, "_create_nginx_cert", side_effect=None)
        mocker.patch.object(OMUpgradeProcessor, "_copy_cert", side_effect=None)
        mocker.patch.object(OMUpgradeProcessor, "_create_uds_cert", side_effect=None)
        assert not OMUpgradeProcessor.update_cert(OMUpgradeProcessor())

    def test_tasks(self):
        installer = OMUpgradeProcessor()
        tasks = installer.tasks()
        assert len(list(tasks)) == 6

    def test_upgrade_exception(self, model: UpgradeCase):
        ret = OMUpgradeProcessor.upgrade(OMUpgradeProcessor())
        assert ret == model.expected

    def test_upgrade(self, mocker: MockerFixture, model: UpgradeCase):
        mocker.patch.object(OMUpgradeProcessor, "tasks", return_value=[TestUtils, ])
        ret = OMUpgradeProcessor.upgrade(OMUpgradeProcessor())
        assert ret == model.expected

    def test_create_nginx_cert_raise(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", return_value=False)
        mocker.patch("pwd.getpwnam", return_value=TestUtils)
        mocker.patch("create_server_certs.create_server_certs", return_value=False)
        with pytest.raises(UpgradeOMError):
            OMUpgradeProcessor._create_nginx_cert()

    def test_create_nginx_cert(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", return_value=True)
        assert not OMUpgradeProcessor._create_nginx_cert()

    def test_change_software_mode_recursive(self, mocker: MockerFixture):
        mocker.patch.object(Chmod, "set_path_permission", return_value=Result(False))
        with pytest.raises(UpgradeOMError):
            OMUpgradeProcessor._change_software_mode_recursive(
                OMUpgradeProcessor(), Path(CommonConstants.OM_UPGRADE_DIR_PATH).joinpath("software", "ibma"), 0o750)
