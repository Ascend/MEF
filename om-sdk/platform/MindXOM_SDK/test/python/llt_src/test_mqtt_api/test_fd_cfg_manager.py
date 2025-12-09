# -*- coding: utf-8 -*-
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
from unittest.mock import mock_open

import mock
import pytest
from pytest_mock import MockerFixture

from common.file_utils import FileUtils
from common.schema import AdapterResult
from common.utils.app_common_method import AppCommonMethod
from common.utils.result_base import Result
from fd_msg_process.config import Config
from lib_restful_adapter import LibRESTfulAdapter
from net_manager.checkers.fd_param_checker import FdMsgChecker
from net_manager.constants import NetManagerConstants
from net_manager.manager.fd_cfg_manager import FdCfgManager
from net_manager.manager.fd_cfg_manager import FdConfigData, FdMsgData, FdMsgException
from net_manager.manager.import_manager import ImportCrl
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.models import NetManager
from net_manager.schemas import WebNetInfo
from test_mef_msg import TestUtils as TestUtils_


class TestUtils:
    def __init__(self, a_a):
        self.status1 = a_a
    ip = "1"
    server_name = ""
    net_mgmt_type = NetManagerConstants.FUSION_DIRECTOR
    port = "403"
    status = AppCommonMethod.OK
    message = {"SerialNumber": ""}


class TestUtils4(FdConfigData):
    data = FdConfigData
    cert_contents = "123"
    status = "connecting"

    @staticmethod
    def adapter(a_a, b_b):
        return

    @staticmethod
    def get_sys_info():
        return

    class net_manager(NetManager):
        status = "ready"
        ip = "1"
        node_id = "1"
        net_mgmt_type = NetManagerConstants.FUSION_DIRECTOR
        port = "1"

        @staticmethod
        def decrypt_cloud_pwd():
            return


class TestUtils0(WebNetInfo, TestUtils4):
    @staticmethod
    def decrypt_cloud_pwd():
        return


class TestUtils1(TestUtils):
    cert_contents = "123"

    def __init__(self, a):
        super().__init__(a)
        self.message = {"SerialNumber": "", "AssetTag": "a"}


class TestUtils2(TestUtils4):
    class net_manager(FdConfigData):
        net_mgmt_type = 0

    class cert_manager:
        @staticmethod
        def get_all():
            return []

        @staticmethod
        def get_all_contain_expired():
            return []

    @staticmethod
    def modify_alarm(a):
        return []


class TestUtils3(TestUtils4):
    @staticmethod
    def get_sys_info():
        return TestUtils4

    class cert_manager:
        @staticmethod
        def get_all():
            return [TestUtils1(""), ]

        @staticmethod
        def get_all_contain_expired():
            return [TestUtils1(""), ]


class TestGetJsonInfoObj:
    CheckGetCurFdIpCase = namedtuple("CheckGetCurFdIpCase", "expected, get_net_cfg_info")
    CheckUpDateEtcHostsCase = namedtuple("CheckUpDateEtcHostsCase", "expected, check_status_is_ok")
    CheckRestoreMiniOsConfigCase = namedtuple("CheckRestoreMiniOsConfigCase", "exists, import_deal")
    CheckGetSysInfoCase = namedtuple("CheckGetSysInfoCase", "expected, class1")
    CheckGetWsConfigCase = namedtuple("CheckGetWsConfigCase", "expected, class1")
    CheckFromStrCase = namedtuple("CheckFromStrCase", "msg, loads")
    use_cases = {
        "test_get_cur_fd_ip": {
            "first": ('1', [TestUtils(1), ]),
            "all": ("", Exception()),
        },
        "test_get_cur_fd_server_name": {
            "first": (NetManagerConstants.SERVER_NAME, TestUtils),
            "all": (NetManagerConstants.SERVER_NAME, Exception()),
        },
        "test_check_fd_mode_and_status_ready": {
            "first": (False, [TestUtils(1)]),
            "all": (False, Exception()),
        },
        "test_check_manager_type_is_web": {
            "first": (False, [TestUtils(1)]),
            "all": (True, Exception()),
        },
        "test_update_etc_hosts": {
            "first": (False, [False, False]),
            "second": (False, [True, False]),
            "third": (True, [True, True]),
        },
        "test_restore_mini_os_config": {
            "first": (False, [True, ]),
            "second": (True, [True, ]),
            "third": (True, True),
        },
        "test_get_sys_info_first": {
            "first": (False, TestUtils4),
            "second": (False, FdCfgManager),
        },
        "test_get_sys_info_second": {
            "first": (True, TestUtils1, ),
            "second": (False, TestUtils),
        },
        "test_get_ws_config": {
            "first": (True, TestUtils2, ),
            "second": (False, TestUtils4),
            "third": (True, TestUtils3,),
        },
        "test_from_str_raise": {
            "first": ("a", Result(True)),
            "second": ("a", Exception()),
            "third": ("a" * Config.mqtt_max_msg_payload_size, Result(True)),
        },
    }

    def test_get_cur_fd_ip(self, mocker: MockerFixture, model: CheckGetCurFdIpCase):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info").side_effect = model.get_net_cfg_info
        ret = FdCfgManager.get_cur_fd_ip()
        assert ret == model.expected

    def test_get_cur_fd_server_name(self, mocker: MockerFixture, model: CheckGetCurFdIpCase):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info").side_effect = model.get_net_cfg_info
        ret = FdCfgManager.get_cur_fd_server_name()
        assert ret == model.expected

    def test_check_fd_mode_and_status_ready(self, mocker: MockerFixture, model: CheckGetCurFdIpCase):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info").side_effect = model.get_net_cfg_info
        ret = FdCfgManager.check_fd_mode_and_status_ready()
        assert ret == model.expected

    def test_check_manager_type_is_web(self, mocker: MockerFixture, model: CheckGetCurFdIpCase):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info").side_effect = model.get_net_cfg_info
        ret = FdCfgManager.check_manager_type_is_web()
        assert ret == model.expected

    def test_update_etc_hosts(self, mocker: MockerFixture, model: CheckUpDateEtcHostsCase):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface").side_effect = [True, True]
        mocker.patch.object(AppCommonMethod, "check_status_is_ok").side_effect = model.check_status_is_ok
        ret = FdCfgManager.update_etc_hosts("", "", "")
        assert bool(ret) == model.expected

    @mock.patch("json.loads",
                mock.Mock(return_value={"cert": [{"cert_contents": "", "source": "", "name": ""}], "config": ""}))
    def test_restore_mini_os_config(self, mocker: MockerFixture, model: CheckRestoreMiniOsConfigCase):
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch("builtins.open", mock_open(read_data="model.read_data"))
        mocker.patch.object(ImportCrl, "import_deal").side_effect = model.import_deal
        mocker.patch.object(NetCfgManager, "update_net_cfg_info").side_effect = [True, ]
        mocker.patch.object(FileUtils, "delete_dir_content").side_effect = [True, ]
        assert not FdCfgManager.restore_mini_os_config()

    def test_restore_mini_os_config_ex(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", return_value=True)
        assert not FdCfgManager.restore_mini_os_config()

    def test_check_and_reset_status_first(self):
        assert not FdCfgManager.check_and_reset_status()

    def test_check_and_reset_status(self, mocker: MockerFixture):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info").side_effect = TestUtils4
        mocker.patch.object(NetCfgManager, "update_net_cfg_info").side_effect = [True, ]
        assert not FdCfgManager.check_and_reset_status()

    def test_modify_alarm(self, mocker: MockerFixture):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface").side_effect = [True, ]
        mocker.patch.object(LibRESTfulAdapter, "check_status_is_ok").side_effect = [True, ]
        assert not FdCfgManager.modify_alarm([])

    def test_check_status_is_ready(self):
        assert FdCfgManager.check_status_is_ready(TestUtils4)

    def test_check_is_switch_fd(self):
        assert FdCfgManager.check_is_switch_fd(TestUtils4, "")

    def test_get_sys_info_first(self, model: CheckGetSysInfoCase):
        ret = FdCfgManager.get_sys_info(model.class1)
        assert bool(ret) == model.expected

    def test_get_sys_info_second(self, mocker: MockerFixture, model: CheckGetSysInfoCase):
        mocker.patch.object(AdapterResult, "from_dict").side_effect = model.class1
        ret = FdCfgManager.get_sys_info(TestUtils4)
        assert bool(ret) == model.expected

    def test_get_ws_config(self, model: CheckGetWsConfigCase):
        ret = FdCfgManager.get_ws_config(model.class1)
        assert bool(ret) == model.expected

    def test_check_cert_status(self):
        assert not FdCfgManager.check_cert_status(FdCfgManager)

    def test_start_cert_status_monitor(self):
        assert not FdCfgManager.start_cert_status_monitor(FdCfgManager)

    def test_post_init(self):
        assert not FdConfigData.__post_init__(FdConfigData)

    def test_from_dict(self):
        assert FdConfigData.from_dict(**{"net": TestUtils0, "sys": FdConfigData})

    def test_gen_extra_headers(self):
        assert FdConfigData(cloud_pwd="abc", cloud_user="abc").gen_extra_headers()

    def test_from_str_raise(self, mocker: MockerFixture, model: CheckFromStrCase):
        mocker.patch("json.loads", side_effect=[model.loads, ])
        with pytest.raises(FdMsgException):
            assert FdMsgData.from_str(model.msg)

    def test_from_str(self, mocker: MockerFixture):
        mocker.patch("json.loads", side_effect=["", ])
        mocker.patch.object(FdMsgChecker, "check", return_value=Result(True))
        mocker.patch.object(FdMsgData, "from_dict", return_value=Result(True))
        assert FdMsgData.from_str("a")

    def test_gen_ws_msg_obj(self):
        assert FdMsgData.gen_ws_msg_obj("", "alarm")

    def test_to_ws_msg_str(self):
        CopyFile = namedtuple("CopyFile1", ["header", "route", "content"])
        backup_restore_base_two = CopyFile(*(TestUtils_, TestUtils_, {"header": "1"}))
        assert FdMsgData.to_ws_msg_str(backup_restore_base_two)
