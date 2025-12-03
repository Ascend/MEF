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
import pytest
from pytest_mock import MockerFixture

from common.common_methods import CommonMethods
from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.utils.result_base import Result


class TestCommonMethods:
    @staticmethod
    def test_constant():
        assert CommonMethods.OK == 200 and CommonMethods.ERROR == 400 and \
               CommonMethods.PARTIAL_OK == 206 and CommonMethods.INTERNAL_ERROR == 500 and \
               CommonMethods.IGNORE_ERROR == 800 and CommonMethods.NFS_MAX_CFG_NUM == 32

    @staticmethod
    def test_object_to_json():
        status, message = CommonMethods.NOT_EXIST, "Model name is not found."
        ret = CommonMethods.object_to_json(status, message)
        assert ret["status"] == status and ret["message"] == message

    @staticmethod
    def test_get_value_by_key_with_empty_value():
        value, key1 = "", "key"
        assert not CommonMethods.get_value_by_key(value, key1)

    @staticmethod
    def test_get_value_by_key():
        file_content = """
            MemTotal:       11856388 kB
            MemFree:         9224656 kB
            MemAvailable:   10502136 kB
        """
        assert CommonMethods.get_value_by_key(file_content, "MemTotal", ":", last_match=True) == "11856388 kB"

    @staticmethod
    def test_check_json_data_with_none_json_data():
        json_data = None
        ret = CommonMethods.check_json_data(json_data)
        assert ret[0] == 0 and ret[1] == json_data

    @staticmethod
    def test_check_json_data_with_dict_json_data():
        json_data = {"json": "json"}
        ret = CommonMethods.check_json_data(json_data)
        assert ret[0] == 0 and ret[1] == json_data

    @staticmethod
    def test_check_json_data_with_wrong_type():
        json_data = "json"
        ret = CommonMethods.check_json_data(json_data)
        assert ret[0] == 1 and ret[1] == "Request data is not json."

    @staticmethod
    def test_check_json_data_with_wrong_dict():
        json_data = "{'json': 'json'}"
        ret = CommonMethods.check_json_data(json_data)
        assert ret[0] == 1 and ret[1] == "Request data is not json."

    @staticmethod
    def test_check_duplicate_attributes_with_duplicate_key():
        json_data = [("json", "test"), ("json", "double json")]
        with pytest.raises(Exception) as exception_info:
            CommonMethods.check_duplicate_attributes(json_data)
            assert "Duplicate attributes" in str(exception_info.value)

    @staticmethod
    def test_check_duplicate_attributes():
        json_data = [("json", "test"), ("second json", "double json")]
        assert CommonMethods.check_duplicate_attributes(json_data) == {"json": "test", "second json": "double json"}

    @staticmethod
    def test_get_config_value(mocker: MockerFixture):
        section_name, key = "iBMA_System", "iBMA_socket_path"
        mocker.patch.object(os.path, "join", return_value="")
        assert not CommonMethods.get_config_value(section_name, key)

    @staticmethod
    def test_load_net_tag_ini_with_check_path_failed(mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(result=False))
        assert not CommonMethods.load_net_tag_ini()

    @staticmethod
    def test_load_net_tag_ini_with_check_path_is_root_failed(mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(result=True))
        mocker.patch.object(FileCheck, "check_path_is_root", return_value=Result(result=False))
        assert not CommonMethods.load_net_tag_ini()

    @staticmethod
    def test_load_net_tag_ini_with_file_size_is_too_large(mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(result=True))
        mocker.patch.object(FileCheck, "check_path_is_root", return_value=Result(result=True))
        mocker.patch("os.path.getsize", return_value=CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES+1)
        assert not CommonMethods.load_net_tag_ini()

    @staticmethod
    def test_write_net_tag_ini_with_check_failed(mocker: MockerFixture):
        tag_ini_data = str({"eth0": ["default"]})
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=Result(result=False))
        assert not CommonMethods.write_net_tag_ini(tag_ini_data)

    @staticmethod
    def test_write_net_tag_ini(mocker: MockerFixture):
        tag_ini_data = str({"eth0": ["default"]})
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=Result(result=True))
        mocker.patch("os.fdopen", ).return_value.__enter__.return_value.write.side_effect = tag_ini_data
        assert CommonMethods.write_net_tag_ini(tag_ini_data)
