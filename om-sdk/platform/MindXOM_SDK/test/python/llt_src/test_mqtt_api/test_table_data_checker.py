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

from net_manager.checkers.table_data_checker import CertManagerChecker, NetManagerCfgChecker


class TestUtils1:
    node_id = "3c7a6b5d-8f1e-4a3d-9b5c-2c6d8e1f0a4b"
    success = True
    net_mgmt_type = "FusionDirector"
    server_name = "FusionDirector"
    ip = "127.0.1.1"
    port = "403"
    cloud_user = "123"
    cloud_pwd = "123"
    status = "ready"

    @staticmethod
    def to_dict():
        return {"name": "a"}

    @staticmethod
    def name():
        return ""


class TestUtils:
    node_id = ""
    net_mgmt_type = ""
    server_name = ""
    ip = ""
    port = ""
    cloud_user = ""
    cloud_pwd = ""
    status = ""


class TestGetJsonInfoObj:
    CheckDictCase = namedtuple("CheckDictCase", "tested_class, param_key, input_param, excepted_success, check")
    CheckWebCfgCase = namedtuple("CheckWebCfgCase", "net_manager, excepted_success")
    CheckFdCfgCase = namedtuple("CheckFdCfgCase", "net_manager, excepted_success")
    CheckNetCfgCase = namedtuple("CheckNetCfgCase", "net_manager, excepted_success")
    use_cases = {
        "test_check_net_cfg": {
            "False": (TestUtils, False),
            "True": (TestUtils1, True),
        },
        "test_check_fd_cfg": {
            "False": (TestUtils, False),
            "True": (TestUtils1, True),
        },
        "test_check_web_cfg": {
            "False": (TestUtils1, False),
            "True": (TestUtils, True),
        },
        "test_check_dict": {
            "not_exists_checker_is_existed": (CertManagerChecker, "1", {"1": TestUtils1}, False, TestUtils1),
            "not_exists_checker_not_existed": (CertManagerChecker, "2", {"1": "1"}, False, TestUtils1),
            "not net_manager": (CertManagerChecker, None, {}, False, TestUtils1),
        },
        "test_check_dict1": {
            "not_exists_checker_is_existed": (NetManagerCfgChecker, "1", {"1": TestUtils1}, True, TestUtils1),
            "not_exists_checker_not_existed": (NetManagerCfgChecker, "2", {"1": "1"}, False, TestUtils1),
            "not net_manager": (NetManagerCfgChecker, None, {}, False, TestUtils1),
        },
    }

    def test_check_dict(self, model: CheckDictCase):
        ret = model.tested_class(model.param_key).check_dict(model.input_param)
        assert model.excepted_success == ret.success

    def test_check_web_cfg(self, model: CheckWebCfgCase):
        ret = NetManagerCfgChecker.check_web_cfg(model.net_manager)
        assert model.excepted_success == ret.success

    def test_check_fd_cfg(self, model: CheckFdCfgCase):
        ret = NetManagerCfgChecker.check_fd_cfg(model.net_manager)
        assert model.excepted_success == ret.success

    def test_check_net_cfg(self, model: CheckNetCfgCase):
        ret = NetManagerCfgChecker.check_net_cfg(NetManagerCfgChecker, model.net_manager)
        assert model.excepted_success == ret.success

    def test_check_dict1(self, model: CheckDictCase):
        ret = model.tested_class(model.param_key).check_dict(model.input_param)
        assert model.excepted_success == ret.success
