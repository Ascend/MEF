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
import configparser
from collections import namedtuple
from typing import Any

from pytest_mock import MockerFixture

from common.constants.error_codes import CommonErrorCodes
from common.exception.biz_exception import BizException
from common.file_utils import FileCheck, FileUtils
from common.utils.exec_cmd import ExecCmd
from common.utils.policy_based_route import PolicyBasedRouting
from common.utils.result_base import Result
from conftest import TestBase
from devm.device_mgr import DEVM
from devm.exception import DeviceManagerError, DeviceNotExistError
from lib.Linux.systems.lte_info import set_lte_dial_up_authtype, LteInfo
from test_mqtt_api.get_log_info import GetLogInfo

getLog = GetLogInfo()


# 用于Mock ConfigParser类
class MockConfigParser(configparser.RawConfigParser):

    def __init__(self, sec1: str, sec2: str):
        super().__init__()
        self.sec1 = sec1
        self.sec2 = sec2

    def read(self, filenames, encoding=None):
        return [self.sec1, self.sec2]

    def sections(self):
        return [self.sec1, self.sec2]

    def get(self, section: str, option: str, *args: Any, **kwargs: Any):
        x = "x"
        if section == self.sec1 and option == self.sec2:
            x = "1"
        return x

    def set(self, section: str, option: str, value: str = "test", *args: Any, **kwargs: Any):
        return ""

    def getboolean(self, section: str, option: str, *args: Any, **kwargs: Any):
        if section and option:
            return True
        else:
            return False


class TestSetLteDialUpAuthType(TestBase):

    @staticmethod
    def test_set_lte_dial_up_authtype_invalid_path(mocker: MockerFixture):
        getLog.clear_log()
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid",
                            return_value=Result(result=False, err_msg="path is not exists"))
        set_lte_dial_up_authtype()
        assert "path is not exists" in getLog.get_log()

    @staticmethod
    def test_set_lte_dial_up_authtype_invalid_content(mocker: MockerFixture):
        getLog.clear_log()
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(result=True))
        mocker.patch("configparser.RawConfigParser", return_value=MockConfigParser("1", "2"))
        set_lte_dial_up_authtype()
        assert "" in getLog.get_log()

    @staticmethod
    def test_set_lte_dial_up_authtype_invalid_authtype(mocker: MockerFixture):
        getLog.clear_log()
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(result=True))
        mocker.patch("configparser.RawConfigParser", return_value=MockConfigParser("lte_apn", "2"))
        set_lte_dial_up_authtype()
        assert "set lte dial up auth type failed !x" in getLog.get_log()

    @staticmethod
    def test_set_lte_dial_up_authtype_succeed(mocker: MockerFixture):
        getLog.clear_log()
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(result=True))
        mocker.patch("configparser.RawConfigParser", return_value=MockConfigParser("lte_apn", "auth_type"))
        mocker.patch.object(DEVM, "get_device")
        set_lte_dial_up_authtype()
        assert "set lte dial up auth type info!" in getLog.get_log()


class TestLteInfo:
    use_cases = {
        "test_set_lte_state": {
            "error1": ([400, "ERR.001,LTE open failed!"],
                       [DeviceManagerError], True, False, None, None, None),
            "error2": ([400, "ERR.002,LTE close failed!"],
                       [DeviceManagerError], False, True, None, None, None),
            "error3": ([400, "ERR.003,data open failed!"],
                       [DeviceManagerError], True, True, True, None, None),
            "error4": ([400, "ERR.004,data close failed!"],
                       [DeviceManagerError], True, True, False, None, None),
            "OK": ([200, ""],
                   [None, None, None, None, None, None], False, False, False, None, None),
        },
        "test_modify_lte": {
            "error1": ("get lte config info from file failed: ",
                       BizException(CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST), None, None, None),
            "error2": ("modify lte config file failed:",
                       MockConfigParser("lte_apn", "auth_type"), None, None, Exception),
        },
        "test_get_ip_address": {
            "normal": ("succeed", [0, "succeed"]),
            "error": ("", [1, "failed"])
        },
        "test_get_ethernet_default_gateway": {
            "normal": (["eth0", "51.38.64.1"], ["51.38.64.1", None]),
            "null": ([], [None, None])
        },
        "test_get_ethernet_gateway": {
            "null": ("", [1, ""]),
            "patched": ("51.38.64.1", [0, "default via 51.38.64.1 dev enp125s0f4"]),
            "not_patched": ("", [0, "test "])
        },
        "test_config_policy_routing": {
            "no_gateway": ([200, ""], [], None, None, None, None, None),
            "config_policy_route_failed": ([400, "config policy routes failed before open LTE"],
                                           ["eth", "192.168.1.1"], "10", None, None, None, False),
            "succeed": ([200, ""], ["eth", "192.168.1.1"], "10", None, None, None, True),
        },
        "test_cancel_policy_routing": {
            "get_option_failed": ([200, ""],
                                  BizException(CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST), None, None, None),
            "no_route_table_id": ([200, ""],
                                  [{"eth_name": "test", "eth_gw_ip": "192.168.1.1"}], None, None, None),
            "init_exception": ([400, "init PolicyBasedRouting caught exception"],
                               [{"eth_name": "test", "eth_gw_ip": "192.168.1.1.1", "route_table_id": "10"}],
                               None, None, None),
            "normal": ([200, ""], [{"eth_name": "test", "eth_gw_ip": "192.168.1.1", "route_table_id": "10"}],
                       None, None, None),
        },
        "test_check_input_params": {
            "check_operator_failed": ([400, "username or request_ip wrong format"], "admin", "192", None, None, None),
            "check_fd_server_ip_failed": ([400, "The fd_server_ip is invalid"],
                                          "admin", "192.168.1.1", "None", None, None),
            "check_state_data_failed": ([400, "LTE state_data wrong format"],
                                        "admin", "192.168.1.1", "192.168.1.1", "None", None),
            "check_state_lte_failed": ([400, "LTE state_lte wrong format"],
                                       "admin", "192.168.1.1", "192.168.1.1", True, None),
            "not state_lte and state_data": ([400, "Could not open data when lte is closed"],
                                             "admin", "192.168.1.1", "192.168.1.1", True, False),
            "OK": ([200, ""], "admin", "192.168.1.1", "192.168.1.1", True, True),
        },
        "test_get_all_info": {
            "not_has_device": ("", False, None, None, None),
            "get_present_failed": ("", True, [False, ], None, None),
            "get_present_exception": ("device Wireless_Module not exists", True, [DeviceNotExistError, ], None, None),
            "lte_base_init_false": ("", True, [True, ], False, False),
            "get_sim_state_failed": ("", True, [True, False], False, True),
            "check_path_failed": ("path is not exists", True, [True, True],
                                  Result(result=False, err_msg="path is not exists"), True),
            "get_signal_info_exception": ("get signal strength failed.", True, [True, True, DeviceManagerError],
                                          Result(result=True), True),
        },
        "test_patch_request": {
            "locked": ([400, 'Lte modify is busy'], None, True, None, None, None, None, None),
            "sim_state_false": ([400, 'ERR.005,sim card is not exist!'],
                                None, False, [False, ], None, None, None, None),
            "sim_state_exception": ([400, 'Device Wireless_Module not exists!'],
                                    None, False, [DeviceNotExistError, ], None, None, None, None),
            "invalid_param": ([400, 'username or request_ip wrong format'],
                              {"state_data": "", "state_lte": "", "fd_server_ip": "", "_User": ",./", "_Xip": ",./"},
                              False, [True, ], None, None, None, None),
            "invalid_path": ([400, 'Lte config reads failed'],
                             {"state_data": True,
                              "state_lte": True,
                              "fd_server_ip": "192.168.1.1",
                              "_User": "admin",
                              "_Xip": "192.168.1.1"},
                             False, [True, ], Result(result=False, err_msg="path is not exists"), None, None, None),
            "config_routing_failed": ([400, 'error'],
                                      {"state_data": True,
                                       "state_lte": True,
                                       "fd_server_ip": "192.168.1.1",
                                       "_User": "admin",
                                       "_Xip": "192.168.1.1"},
                                      False, [True, ], Result(result=True), [400, 'error'], None, None),
            "switch_to_lte_false": ([400, 'error'],
                                    {"state_data": False,
                                     "state_lte": False,
                                     "fd_server_ip": "192.168.1.1",
                                     "_User": "admin",
                                     "_Xip": "192.168.1.1"},
                                    False, [True, ], Result(result=True), [0, ""], [0, ""], [400, 'error']),
            "set_lte_state_failed": ([400, 'error'],
                                     {"state_data": True,
                                      "state_lte": True,
                                      "fd_server_ip": "192.168.1.1",
                                      "_User": "admin",
                                      "_Xip": "192.168.1.1"},
                                     False, [True, ], Result(result=True), [200, ""], [400, 'error'], None),
            "succeed": ([200, ],
                        {"state_data": True,
                         "state_lte": True,
                         "fd_server_ip": "192.168.1.1",
                         "_User": "admin",
                         "_Xip": "192.168.1.1"},
                        False, [True, ], Result(result=True), [200, ""], [200, ""], [200, ""]),
        }

    }
    PatchRequestCase = namedtuple("PatchRequestCase",
                                  "expect, request_dict, lock, get_device, check_path, "
                                  "config_routing, set_lte_state, cancel_routing")
    GetAllInfoCase = namedtuple("GetAllInfoCase", "expect, has_device, get_device, check_path, lte_base_init")
    CheckInputParamsCase = namedtuple("CheckInputParamsCase",
                                      "expect, username, request_ip, fd_server_ip, state_data, state_lte")
    CancelPolicyRoutingCase = namedtuple("CancelPolicyRoutingCase",
                                         "expect, get_option, cancel_policy_route, fd_server_ip, web_request_ip")
    SetLteStateCase = namedtuple("SetLteStateCase",
                                 "expect, get_device, state_lte, state_lte_file, state_data, username, request_ip")
    ModifyLteCase = namedtuple("ModifyLteCase", "expect, get_config, state_lte, state_data, check_is_link")
    GetIPAddressCase = namedtuple("GetIPAddressCase", "expect, exec_cmd")
    GetDefaultGatewayCase = namedtuple("GetDefaultGatewayCase", "expect, gateway")
    GetEthernetGatewayCase = namedtuple("GetEthernetGatewayCase", "expect, exec_cmd")
    ConfigPolicyRoutingCase = namedtuple("ConfigPolicyRoutingCase",
                                         "expect, get_gateway, get_option, routing, "
                                         "fd_server_ip, web_request_ip, config_policy")

    @staticmethod
    def test_set_lte_state(mocker: MockerFixture, model: SetLteStateCase):
        mocker.patch.object(DEVM, "get_device", side_effect=model.get_device)
        assert LteInfo().set_lte_state(model.state_lte,
                                       model.state_lte_file,
                                       model.state_data,
                                       model.username,
                                       model.request_ip) == model.expect

    @staticmethod
    def test_modify_lte(mocker: MockerFixture, model: ModifyLteCase):
        getLog.clear_log()
        if isinstance(model.get_config, Exception):
            mocker.patch.object(FileUtils, "get_config_parser", side_effect=model.get_config)
        else:
            mocker.patch.object(FileUtils, "get_config_parser", return_value=model.get_config)
        mocker.patch.object(FileCheck, "check_is_link_exception", side_effect=model.check_is_link)
        mocker.patch("os.fdopen").return_value.__enter__.return_value.read.side_effect = "abc"
        LteInfo().modify_lte(model.state_lte, model.state_data)
        assert model.expect in getLog.get_log()

    @staticmethod
    def test_get_ip_address(mocker: MockerFixture, model: GetIPAddressCase):
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol", return_value=model.exec_cmd)
        assert LteInfo().get_ip_address() == model.expect

    @staticmethod
    def test_get_ethernet_default_gateway(mocker: MockerFixture, model: GetDefaultGatewayCase):
        mocker.patch.object(LteInfo, "_get_ethernet_gateway", side_effect=model.gateway)
        assert LteInfo()._get_ethernet_default_gateway() == model.expect

    @staticmethod
    def test_get_ethernet_gateway(mocker: MockerFixture, model: GetEthernetGatewayCase):
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol", return_value=model.exec_cmd)
        assert LteInfo()._get_ethernet_gateway("") == model.expect

    @staticmethod
    def test_config_policy_routing(mocker: MockerFixture, model: ConfigPolicyRoutingCase):
        mocker.patch.object(LteInfo, "_get_ethernet_default_gateway", return_value=model.get_gateway)
        mocker.patch.object(FileUtils, "get_option", return_value=model.get_option)
        mocker.patch("common.utils.policy_based_route.PolicyBasedRouting", side_effect=model.routing)
        mocker.patch.object(PolicyBasedRouting, "config_policy_route", return_value=model.config_policy)
        assert LteInfo()._config_policy_routing(model.fd_server_ip, model.web_request_ip) == model.expect

    @staticmethod
    def test_cancel_policy_routing(mocker: MockerFixture, model: CancelPolicyRoutingCase):
        mocker.patch.object(FileUtils, "get_option_list", side_effect=model.get_option)
        mocker.patch.object(PolicyBasedRouting, "cancel_policy_route", return_value=model.cancel_policy_route)
        assert LteInfo()._cancel_policy_routing(model.fd_server_ip, model.web_request_ip) == model.expect

    @staticmethod
    def test_check_input_params(model: CheckInputParamsCase):
        assert LteInfo()._check_input_params(model.username,
                                             model.request_ip,
                                             model.fd_server_ip,
                                             model.state_data,
                                             model.state_lte) == model.expect

    @staticmethod
    def test_get_all_info(mocker: MockerFixture, model: GetAllInfoCase):
        getLog.clear_log()
        mocker.patch.object(DEVM, "has_device", return_value=model.has_device)
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.side_effect = model.get_device
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch.object(LteInfo, "lte_base_init", return_value=model.lte_base_init)
        mocker.patch("configparser.ConfigParser", return_value=MockConfigParser("lte_apn", "auth_type"))
        LteInfo().get_all_info()
        assert model.expect in getLog.get_log()

    @staticmethod
    def test_patch_request(mocker: MockerFixture, model: PatchRequestCase):
        mocker.patch.object(LteInfo, "LTE_LOCK").locked.return_value = model.lock
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.side_effect = model.get_device
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch("configparser.ConfigParser", return_value=MockConfigParser("lte_apn", "auth_type"))
        mocker.patch.object(LteInfo, "_config_policy_routing", return_value=model.config_routing)
        mocker.patch.object(LteInfo, "set_lte_state", return_value=model.set_lte_state)
        mocker.patch.object(LteInfo, "modify_lte")
        mocker.patch.object(LteInfo, "get_all_info")
        mocker.patch.object(LteInfo, "_cancel_policy_routing", return_value=model.cancel_routing)
        mocker.patch("configparser.ConfigParser", return_value=MockConfigParser("lte_apn", "auth_type"))
        assert LteInfo().patch_request(model.request_dict) == model.expect
