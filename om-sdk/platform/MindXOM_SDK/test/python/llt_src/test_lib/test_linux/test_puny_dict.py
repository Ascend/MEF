# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os

from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.file_utils import FilePermission
from common.utils.exec_cmd import ExecCmd
from common.utils.result_base import Result
from lib.Linux.systems.security_service.puny_dict import PunyDict
from common.common_methods import CommonMethods
from ut_utils.mock_utils import mock_check_input_path_valid
from ut_utils.mock_utils import mock_check_path


class TestPunyDict:
    CURRENT_PATH = os.path.dirname(os.path.realpath(__file__))
    IMPORT_FILE = os.path.join(CURRENT_PATH, "import_weak_dict.conf")
    EXPORT_FILE = os.path.join(CURRENT_PATH, "export_weak_dict.conf")
    DELETE_FILE = os.path.join(CURRENT_PATH, "delete_weak_dict.conf")

    def setup_class(self):
        with open(self.IMPORT_FILE, "w+") as f_out:
            f_out.write("Test@12#$\nTest@6666\n123456\n")

    def teardown_class(self):
        for path in self.IMPORT_FILE, self.EXPORT_FILE, self.DELETE_FILE:
            if not os.path.exists(path):
                continue
            os.remove(path)

    def test_update_cracklib(self, mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol", return_value=[0, "success"])
        mocker.patch.object(FilePermission, "set_path_permission")
        PunyDict.update_cracklib(self.IMPORT_FILE)

    def test_delete_weak_dict(self, mocker: MockerFixture):
        mock_check_input_path_valid(mocker, True)
        mocker.patch("os.open")
        mocker.patch("os.fdopen").return_value.__enter__.return_value.write.side_effect = "abc"
        mocker.patch.object(PunyDict, "update_cracklib")
        mocker.patch("lib.Linux.systems.security_service.puny_dict.set_puny_dict_sign", side_effect="delete")
        assert PunyDict().delete_weak_dict() == [CommonMethods.OK, "Delete weak dictionary success."]

    def test_export_weak_dict(self, mocker: MockerFixture):
        mock_check_path(mocker, True)
        mocker.patch("lib.Linux.systems.security_service.puny_dict.get_puny_dict_sign", return_value="import")
        mocker.patch("os.open")
        mocker.patch("os.chmod")
        mocker.patch("os.fdopen", ).return_value.__enter__.return_value.write.side_effect = "abc"
        mocker.patch.object(ExecCmd, "exec_cmd_get_output", return_value=[0, "success"])
        assert PunyDict().export_weak_dict() == [CommonMethods.OK, "Export weak dictionary success."]

    def test_import_weak_dict(self, mocker: MockerFixture):
        mocker.patch.object(PunyDict, "process_import_file")
        mocker.patch.object(PunyDict, "update_cracklib")
        mocker.patch("lib.Linux.systems.security_service.puny_dict.set_puny_dict_sign", side_effect="import")
        assert PunyDict().import_weak_dict() == [CommonMethods.OK, "Import weak dictionary success."]

    def test_process_import_file(self, mocker: MockerFixture):
        mock_check_path(mocker, True)
        mocker.patch("os.path.getsize", return_value=19.8)
        mocker.patch.object(PunyDict, "read_import_file", return_value=["11", "22"])
        mocker.patch("os.open")
        mocker.patch("os.fdopen").return_value.__enter__.return_value.write.side_effect = "abc"
        PunyDict().process_import_file()

    def test_read_import_file(self, mocker: MockerFixture):
        mocker.patch("os.open")
        mocker.patch("os.fdopen").return_value.__enter__.return_value.read.side_effect = "abc"
        PunyDict().read_import_file(self.IMPORT_FILE)

    def test_post_request(self, mocker: MockerFixture):
        mocker.patch.object(PunyDict, "PUNY_DICT_LOCK").locked.return_value = False
        mocker.patch.object(
            PunyDict,
            "import_weak_dict",
            return_value=[CommonMethods.OK, "Import weak dictionary success."]
        )
        mocker.patch.object(
            PunyDict,
            "export_weak_dict",
            return_value=[CommonMethods.OK, "Export weak dictionary success."]
        )
        mocker.patch.object(
            PunyDict,
            "delete_weak_dict",
            return_value=[CommonMethods.OK, "Delete weak dictionary success."]
        )
        for opera_type in "import", "export", "delete":
            assert PunyDict().post_request({"OperationType": opera_type}) == \
                   [CommonMethods.OK, f"{opera_type[0].upper()}{opera_type[1:]} weak dictionary success."]
