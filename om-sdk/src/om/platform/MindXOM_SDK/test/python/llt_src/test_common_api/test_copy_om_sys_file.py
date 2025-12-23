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
import os
from collections import namedtuple

import pytest
from pytest_mock import MockerFixture

from bin.environ import Env
from common.constants.base_constants import CommonConstants
from common.constants.upgrade_constants import UpgradeConstants
from common.file_utils import FileUtils, FileCopy
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils
from common.utils.result_base import Result
from copy_om_sys_file import CopySysFileOperator, ParseParamError, CopyFileError

CheckDestPathCase = namedtuple("CheckDestPathCase", "expected, path")
CheckCopyTypeCase = namedtuple("CheckCopyTypeCase", "expected, path")
CheckOmSrcPathCase = namedtuple("CheckOmSrcPathCase", "expected, path")
CopyFilesCase = namedtuple("CopyFilesCase", "expected, files, src_dir, dest_dir, mode")
CopyMefSysCase = namedtuple("CopyMefSysCase", "get_model_by_npu_smi, MefInfo, exec_cmd")
ExecuteCase = namedtuple("ExecuteCase", "expected, parse_param, copy_sys_file_to_system_dir")
CopyToolKitCase = namedtuple("CopyToolKitCase", "delete_full_dir, exists")
SetSoftLinkCase = namedtuple("SetSoftLinkCase", "expected")


class TestUtilsTrue:
    allow_upgrade = True
    cp_sys_sh = True

    @staticmethod
    def check_script_file_valid(x):
        return True if x else Result(result=False, err_msg="path is not exists")


class TestUtilsFalse:
    allow_upgrade = False


class TestUtilsTrueFalse:
    allow_upgrade = True
    cp_sys_sh = ""

    @staticmethod
    def check_script_file_valid(x):
        return True if x else Result(result=False, err_msg="path is not exists")


class TestUtils:
    dest_path = "/opt/middleware/MindXOMBackup"
    copy_type = "backup_area"
    om_src_path = "/opt/middleware/MindXOM"


class TestCopyOmSysFile:
    use_cases = {
        "test_check_dest_path": {
            "normal": ("/", "/"),
        },
        "test_check_dest_path_exception": {
            "exception": ("/", "/data"),
        },
        "test_check_copy_type": {
            "normal": ("omsdk_upgrade", "omsdk_upgrade"),
        },
        "test_check_copy_type_exception": {
            "exception": ("omsdk_upgrade", "o"),
        },
        "test_check_om_src_path": {
            "normal": (CommonConstants.OM_WORK_DIR_PATH, CommonConstants.OM_WORK_DIR_PATH),
        },
        "test_check_om_src_path_exception": {
            "exception": (None, "o"),
        },
        "test_copy_files": {
            "normal": (None, (UpgradeConstants.ASCENDSIP_G2_CRL,), os.path.join("/", "etc", "hwsipcrl"),
                       os.path.join('/', "etc", "hwsipcrl"), 0o644),
        },
        "test_copy_files_exception": {
            "normal": (None, (UpgradeConstants.ASCENDSIP_G2_CRL,), os.path.join("/", "etc", "hwsipcrl"),
                       os.path.join('/', "etc", "hwsipcrl"), 0o644),
        },
        "test_copy_mef_sys": {
            "normal": ("Atlas 500 A2", TestUtilsTrue, 0),
            "false_when_get_model_by_npu_smi_is_A1": ("Atlas 500 A1", TestUtilsTrue, 0),
            "false_when_allow_upgrade_is_false": ("Atlas 500 A2", TestUtilsFalse, 0),
        },
        "test_copy_mef_sys_exception": {
            "mef copy service sh invalid": ("Atlas 500 A2", TestUtilsTrueFalse, 0),
            "copy mef service failed": ("Atlas 500 A2", TestUtilsTrue, 1),
        },
        "test_execute": {
            "false_when_without_param": (1, True, None),
            "normal": (0, TestUtils, None),
        },
        "test_copy_sys_file_to_system_dir": {},
        "test_copy_g2_crl": {
            "normal": (),
        },
        "test_copy_toolkit": {
            "first_none": (None, [False, True]),
            "normal": (None, [True, False]),
        },
        "test_copy_toolkit_delete_full_dir_exception": {
            "first": (Exception(), [True, False]),
            "second": (None, [True, True]),
        },
        "test_copy_toolkit_shutil_exception": {
            "normal": (None,),
        },
        "test_set_soft_link": {
            "normal": (True,),
        },
        "test_set_soft_link_exception": {
            "normal": (True,),
        },
        "test_copy_service": {
            "normal": (),
        },
        "test_parse_param": {
            "exception": (None,),
        },
    }

    def test_check_dest_path(self, model: CheckDestPathCase):
        ret = CopySysFileOperator.check_dest_path(model.path)
        assert model.expected == ret

    def test_check_dest_path_exception(self, model: CheckDestPathCase):
        with pytest.raises(ParseParamError):
            CopySysFileOperator().check_dest_path(model.path)

    def test_check_copy_type(self, model: CheckCopyTypeCase):
        ret = CopySysFileOperator.check_copy_type(model.path)
        assert model.expected == ret

    def test_check_copy_type_exception(self, model: CheckCopyTypeCase):
        with pytest.raises(ParseParamError):
            CopySysFileOperator().check_copy_type(model.path)

    def test_check_om_src_path(self, model: CheckOmSrcPathCase):
        ret = CopySysFileOperator.check_om_src_path(model.path)
        assert model.expected == ret

    def test_check_om_src_path_exception(self, model: CheckOmSrcPathCase):
        with pytest.raises(ParseParamError):
            CopySysFileOperator().check_om_src_path(model.path)

    def test_copy_files(self, mocker: MockerFixture, model: CopyFilesCase):
        mocker.patch.object(FileCopy, "copy_file", return_value=Result(result=True))
        ret = CopySysFileOperator.copy_files(model.files, model.src_dir, model.dest_dir, model.mode)
        assert model.expected == ret

    def test_copy_files_exception(self, mocker: MockerFixture, model: CopyFilesCase):
        mocker.patch.object(FileCopy, "copy_file", return_value=Result(result=False))
        with pytest.raises(CopyFileError):
            CopySysFileOperator.copy_files(model.files, model.src_dir, model.dest_dir, model.mode)

    def test_copy_mef_sys(self, mocker: MockerFixture, model: CopyMefSysCase):
        mocker.patch.object(SystemUtils, "get_model_by_npu_smi", return_value=model.get_model_by_npu_smi)
        mocker.patch("lib.Linux.mef.mef_info.MefInfo", return_value=model.MefInfo)
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=model.exec_cmd)
        assert not CopySysFileOperator.copy_mef_sys("/")

    def test_copy_mef_sys_exception(self, mocker: MockerFixture, model: CopyMefSysCase):
        mocker.patch.object(SystemUtils, "get_model_by_npu_smi", return_value=model.get_model_by_npu_smi)
        mocker.patch("lib.Linux.mef.mef_info.MefInfo", return_value=model.MefInfo)
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=model.exec_cmd)
        with pytest.raises(CopyFileError):
            CopySysFileOperator.copy_mef_sys("/")

    def test_execute(self, mocker: MockerFixture, model: ExecuteCase):
        mocker.patch.object(CopySysFileOperator, "_parse_param", return_value=model.parse_param)
        mocker.patch.object(CopySysFileOperator, "copy_sys_file_to_system_dir",
                            return_value=model.copy_sys_file_to_system_dir)
        ret = CopySysFileOperator.execute(CopySysFileOperator())
        assert model.expected == ret

    def test_copy_sys_file_to_system_dir(self, mocker: MockerFixture):
        mocker.patch.object(CopySysFileOperator, "_copy_service", return_value=None)
        mocker.patch.object(CopySysFileOperator, "_set_soft_link", return_value=True)
        mocker.patch.object(CopySysFileOperator, "copy_mef_sys", return_value=None)
        mocker.patch.object(CopySysFileOperator, "_copy_g2_crl", return_value=None)
        mocker.patch.object(CopySysFileOperator, "_copy_toolkit", side_effect=[None, None, ])
        assert not CopySysFileOperator.copy_sys_file_to_system_dir(
            CopySysFileOperator(), TestUtils.dest_path, TestUtils.copy_type, TestUtils.om_src_path)

    def test_copy_g2_crl(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", side_effect=[True, False])
        mocker.patch.object(CopySysFileOperator, "copy_files", return_value=None)
        assert not CopySysFileOperator._copy_g2_crl(CopySysFileOperator(), TestUtils.om_src_path, TestUtils.dest_path)

    def test_copy_toolkit(self, mocker: MockerFixture, model: CopyToolKitCase):
        mocker.patch.object(Env, "start_from_m2", return_value=True)
        mocker.patch.object(FileUtils, "delete_full_dir", return_value=None)
        mocker.patch("os.path.exists", side_effect=model.exists)
        mocker.patch("shutil.move")
        assert not CopySysFileOperator._copy_toolkit(CopySysFileOperator(), TestUtils.om_src_path, TestUtils.dest_path)

    def test_copy_toolkit_delete_full_dir_exception(self, mocker: MockerFixture, model: CopyToolKitCase):
        mocker.patch.object(Env, "start_from_m2", return_value=True)
        mocker.patch.object(FileUtils, "delete_full_dir").side_effect = model.delete_full_dir
        mocker.patch("os.path.exists", side_effect=model.exists)
        with pytest.raises(StopIteration):
            CopySysFileOperator._copy_toolkit(CopySysFileOperator(), TestUtils.om_src_path, "/home/data/1")

    def test_copy_toolkit_shutil_exception(self, mocker: MockerFixture):
        mocker.patch.object(Env, "start_from_m2", return_value=True)
        mocker.patch.object(FileUtils, "delete_full_dir").side_effect = None
        mocker.patch("os.path.exists", side_effect=[True, True])
        with pytest.raises(StopIteration):
            CopySysFileOperator._copy_toolkit(CopySysFileOperator(), TestUtils.om_src_path, TestUtils.dest_path)

    def test_set_soft_link(self, mocker: MockerFixture, model: SetSoftLinkCase):
        mocker.patch("os.symlink", return_value=True)
        ret = CopySysFileOperator. \
            _set_soft_link(CopySysFileOperator(), os.path.join("/", "usr", "lib", "systemd", "system"),
                           os.path.join(TestUtils.dest_path, "etc", "systemd", "system", "multi-user.target.wants"))
        assert model.expected == ret

    def test_set_soft_link_exception(self):
        with pytest.raises(CopyFileError):
            CopySysFileOperator. \
                _set_soft_link(CopySysFileOperator(), os.path.join("/", "usr", "lib", "systemd", "system"),
                               os.path.join(TestUtils.dest_path, "etc", "systemd", "system", "multi-user.target.wants"))

    def test_copy_service(self, mocker: MockerFixture):
        mocker.patch.object(CopySysFileOperator, "copy_files").return_value = True
        assert not CopySysFileOperator._copy_service(
            CopySysFileOperator(), os.path.join("/", "usr", "lib", "systemd", "system"),
            os.path.join(TestUtils.dest_path, "etc", "systemd", "system", "multi-user.target.wants"))

    def test_parse_param(self):
        with pytest.raises(ParseParamError):
            CopySysFileOperator._parse_param(CopySysFileOperator())
