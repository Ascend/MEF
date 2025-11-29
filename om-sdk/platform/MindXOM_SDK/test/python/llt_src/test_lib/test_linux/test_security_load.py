#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

import configparser
from collections import namedtuple
from unittest.mock import mock_open, patch

from common.file_utils import FileCheck
from common.utils.result_base import Result
from pytest_mock import MockerFixture

from lib.Linux.systems.security_service.login_rule_mgr import LoginRuleManager
from lib.Linux.systems.security_service.models import LoginRules
from ut_utils.mock_utils import mock_path_exists, mock_write_file_with_os_open, mock_os_sync, mock_time_sleep

from common.file_utils import FileCopy
from common.file_utils import FileUtils
from common.net_check import NetCheck
from common.common_methods import CommonMethods
from lib.Linux.systems.security_service.security_load import SecurityLoad

GetAllArpItem = namedtuple("GetAllArpItem", "expect, read_data")
CheckRemoteRequestIp = namedtuple("CheckRemoteRequestIp", "expect, remote_ip, expect_ip, seg")
CheckRemoteRequestMac = namedtuple("CheckRemoteRequestMac", "expect, arp, remote_ip, expect_mac")
CheckUsrLoadTime = namedtuple("CheckUsrLoadTime", "expect, cfg, curr")
CheckOneSessionSecCfg = namedtuple("CheckOneSessionSecCfg", "expect, load, ip, mac")
CheckSessionSecCfgFile = namedtuple("CheckSessionSecCfgFile", "expect, read")
GetUsrSessionSecCfgFile = namedtuple("GetUsrSessionSecCfgFile", "expect, sections, value")
SetAllSessionSecCfg = namedtuple("SetAllSessionSecCfg", "expect, link, add, set, cfg")
CheckSessionRequestUsrData = namedtuple("CheckSessionRequestUsrData", "expect, cfg, check")
CheckTimeHourTime = namedtuple("CHeckTimeHourTime", "expect, hh_mm")
JudgeListRepeated = namedtuple("JudgeListRepeated", "expect, data")
CheckPatchRequestData = namedtuple("CheckPatchRequestData", "expect, request, hh_mm, addr, cast")
PatchRequest = namedtuple("PatchRequest", "expect, request, locked, check, get_all, overwrite")
PostRequest = namedtuple("PostRequest", "expect, request, locked, check, import_cfg, export")
ImportSessionCfgFile = namedtuple("ImportSessionCfgFile",
                                  "expect, link, size, get_cfg, check, over_write")
ExportSessionCfgFile = namedtuple("ExportSessionCfgFile", "expect, link, get, set_cfg")


class TestSecurityLoad:
    use_cases = {
        "test_get_all_arp_item": {
            "empty": (([], []), ""),
            "normal": ((["a"], ["d"]), "\na b c d")
        },
        "test_check_remote_request_ip": {
            "eq": (True, "", "", ""),
            "normal": ("same", "", "/", "same")
        },
        "test_check_remote_request_mac": {
            "null arp": (True, ["", ""], "", ""),
            "normal": (True, ["ip", ["mac"]], "ip", "mac"),
            "not meet": (False, ["ip", [""]], "ip", "mac")
        },
        "test_check_usr_load_time": {
            "null": (True, {"start_time": None, "end_time": None}, None),
            "convert time error": (True, {"start_time": "", "end_time": ""}, None),
            "valid": (True, {"start_time": "00:00", "end_time": "23:00"}, [0, 0, 0, 10, 10]),
            "valid-1": (True, {"start_time": "23:00", "end_time": "01:50"}, [0, 0, 0, 1, 10]),
            "invalid": (False, {"start_time": "23:00", "end_time": "01:50"}, [0, 0, 0, 2, 10])
        },
        "test_check_one_session_sec_cfg": {
            "not load time": (False, False, None, None),
            "not request ip": (False, True, False, None),
            "not request mac": (False, True, True, False),
            "load cfg met": (True, True, True, True),
        },
        "test_get_usr_session_sec_cfg": {
            "get session sec cfg error-1": ([], ["sec"], ValueError),
            "normal": (
                [{"enable": "enable", "end_time": "end_time", "ip_addr": "ip_addr", "mac_addr": "mac_addr",
                  "start_time": "start_time"}], ["sec"],
                ["enable", "start_time", "end_time", "ip_addr", "mac_addr"]
            )
        },
        "test_check_session_request_usr_data": {
            "no available session sec cfg": (True, None, None),
            "true": (True, [{"enable": "false"}, {"enable": "true"}], True),
            "no session sec cfg is enable": (True, [], False),
            "no session sec cfg is met": (False, [{"enable": "false"},
                                                  {"enable": "true"}], False),
        },
        "test_check_patch_request_data": {
            "Request parameter is empty or not dict": (
                Result(result=False, err_msg="Request parameter is empty or not dict"), None, None, None, None
            ),
            "Request sub parameter is empty or not list": (
                Result(result=False, err_msg="Request sub parameter is empty or not list"), {}, None, None, None
            ),
            "Request parameter list len exceeds": (
                Result(result=False, err_msg="Request parameter list len exceeds"),
                {"load_cfg": [""] * 40}, None, None, None
            ),
            "Request sub parameter is not dict": (
                Result(result=False, err_msg="Request sub parameter is not dict 'str' object has no attribute 'get'"),
                {"load_cfg": [""]}, None, None, None
            ),
            "Parameter enable error": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"), {"load_cfg": [{}]}, None, None, None
            ),
            "Parameter end_time is None but start_time not": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"),
                {"load_cfg": [{"enable": "true", "start_time": ""}]}, None, None, None
            ),
            "Parameter end_time is not None but start_time is None": (
                Result(result=False,
                       err_msg="Request sub parameter is not dict argument of type 'NoneType' is not iterable"),
                {"load_cfg": [{"enable": "true", "end_time": ""}]}, None, None, None
            ),
            "Parameter start_time error": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"),
                {"load_cfg": [{"enable": "true", "start_time": "", "end_time": ""}]}, [False, ], None, None
            ),
            "Parameter end_time error": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"),
                {"load_cfg": [{"enable": "true", "start_time": "", "end_time": ""}]}, [True, False], None, None
            ),
            "Parameter mask invalid": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"),
                {"load_cfg": [{"enable": "true", "start_time": "", "end_time": "", "ip_addr": "/" + "*" * 129}]},
                [True, True], None, None
            ),
            "Parameter ip_addr/mask error": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"),
                {"load_cfg": [{"enable": "true", "start_time": "", "end_time": "", "ip_addr": "/*"}]},
                [True, True], None, None
            ),
            "Parameter ip_addr with mask invalid": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"),
                {"load_cfg": [{"enable": "true", "start_time": "", "end_time": "", "ip_addr": "/1"}]},
                [True, True], False, None
            ),
            "Parameter mask invalid-1": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"),
                {"load_cfg": [{"enable": "true", "start_time": "", "end_time": "", "ip_addr": "/34"}]},
                [True, True], True, None
            ),
            "Parameter single ip_addr invalid": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"),
                {"load_cfg": [{"enable": "true", "start_time": "", "end_time": "", "ip_addr": "34"}]},
                [True, True], False, None
            ),
            "Parameter single ip_addr reserved": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"),
                {"load_cfg": [{"enable": "true", "start_time": "", "end_time": "", "ip_addr": "127"}]},
                [True, True], True, None
            ),
            "Parameter mac_addr len invalid": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"),
                {"load_cfg": [{"enable": "true", "start_time": "", "end_time": "", "ip_addr": "1", "mac_addr": ""}]},
                [True, True], True, None
            ),
            "Parameter mac_addr error": (
                Result(result=False, err_msg="SecurityLoadCfg Parameter error"),
                {"load_cfg": [
                    {"enable": "true", "start_time": "", "end_time": "", "ip_addr": "1", "mac_addr": "*" * 8}
                ]},
                [True, True], True, False
            ),
        },
        "test_patch_request": {
            "SecurityLoad modify is busy": (
                [CommonMethods.ERROR, "ERR.711, SecurityLoad modify is busy."], None, True, None, None, None
            ),
            "check data failed": (
                [CommonMethods.ERROR, "ERR.701, Check patch request data fail"], None, False, Result(False, "err"),
                None, None
            ),
            "Patch request data repeated": (
                [CommonMethods.ERROR, "ERR.710, Patch request data repeated"],
                {"Password": "********", "load_cfg": [{"start_time": "00:00", }, {"start_time": "00:00", }]}, False,
                Result(True), None, None
            ),
            "Ok": (
                [CommonMethods.OK, "set config data ok"],
                {"Password": "********",
                 "load_cfg": [{"start_time": "00:00", }, {"start_time": "01:00", }]},
                False, Result(True), [{"start_time": "00:00", }], None
            )
        },
        "test_post_request": {
            "SecurityLoad modify is busy": (
                [CommonMethods.ERROR, "ERR.711, SecurityLoad modify is busy."], None, True, None, None, None
            ),
            "Check upload para failed": (
                [CommonMethods.ERROR, "ERR.703, Upload parameter error."], None, False, Result(False), None, None
            ),
            "Import Security load config failed": (
                [CommonMethods.ERROR, "err"], {"action": "import"}, False, Result(True),
                Result(result=False, err_msg="err"), None
            ),
            "Import Security load config ok": (
                [CommonMethods.OK, "Import security load config ok"],
                {"action": "import"}, False, Result(True), Result(result=True), None
            ),
            "Export Security load config ok": (
                [CommonMethods.OK, "get Security load config ok"], {"action": "export"}, False, Result(True), None,
                Result(result=True)
            ),
            "action not support": (
                [CommonMethods.ERROR, "action not support"], {}, False, Result(True), None, None
            )
        },
        "test_import_session_cfg_file": {
            "check link err": (
                Result(result=False, err_msg="ERR.712, "),
                Result(result=False), None, None, None, None
            ),
            "too large": (
                Result(result=False, err_msg="ERR.709, Import file is too large."),
                Result(result=True), 102400, None, None, None
            ),
            "Check import file configration failed": (
                Result(result=False, err_msg="ERR.705, Check import file configuration failed."),
                Result(result=True), 1024, [], Result(False, "err"), None
            ),
            "ok": (Result(result=True),
                   Result(result=True), 1024, [], Result(True), Result(True))
        },
    }

    @staticmethod
    def test_get_all_arp_item(mocker: MockerFixture, model: GetAllArpItem):
        mocker.patch("builtins.open", mock_open(read_data=model.read_data))
        assert SecurityLoad.get_all_arp_item() == model.expect

    @staticmethod
    def test_check_remote_request_ip(mocker: MockerFixture, model: CheckRemoteRequestIp):
        mocker.patch.object(NetCheck, "is_same_network_segment", return_value=model.seg)
        assert SecurityLoad.check_remote_request_ip(model.remote_ip, model.expect_ip) == model.expect

    @staticmethod
    def test_check_remote_request_mac(mocker: MockerFixture, model: CheckRemoteRequestMac):
        mocker.patch.object(SecurityLoad, "get_all_arp_item", return_value=model.arp)
        assert SecurityLoad.check_remote_request_mac(model.remote_ip, model.expect_mac) == model.expect

    @staticmethod
    def test_check_usr_load_time(mocker: MockerFixture, model: CheckUsrLoadTime):
        mocker.patch("time.localtime", return_value=model.curr)
        assert SecurityLoad.check_usr_load_time(model.cfg) == model.expect

    @staticmethod
    def test_check_one_session_sec_cfg(mocker: MockerFixture, model: CheckOneSessionSecCfg):
        mocker.patch.object(SecurityLoad, "check_usr_load_time", return_value=model.load)
        mocker.patch.object(SecurityLoad, "check_remote_request_ip", return_value=model.ip)
        mocker.patch.object(SecurityLoad, "check_remote_request_mac", return_value=model.mac)
        assert SecurityLoad.check_one_session_sec_cfg("", {"ip_addr": "ip", "mac_addr": "mac"}) \
               == model.expect

    @staticmethod
    def test_get_usr_session_sec_cfg(mocker: MockerFixture, model: GetUsrSessionSecCfgFile):
        cfg_parser = mocker.patch.object(FileUtils, "get_config_parser").return_value
        cfg_parser.sections.return_value = model.sections
        cfg_parser.get.side_effect = model.value
        assert SecurityLoad.get_usr_session_sec_cfg("") == model.expect

    @staticmethod
    def test_check_session_request_usr_data(mocker: MockerFixture, model: CheckSessionRequestUsrData):
        mocker.patch.object(LoginRuleManager, "get_all", return_value=model.cfg)
        mocker.patch.object(SecurityLoad, "check_one_session_sec_cfg", return_value=model.check)
        assert SecurityLoad.check_session_request_usr_data("") == model.expect

    @staticmethod
    def test_check_patch_request_data(mocker: MockerFixture, model: CheckPatchRequestData):
        mocker.patch.object(NetCheck, "mac_addr_single_cast_check", return_value=model.cast)
        assert SecurityLoad.check_patch_request_data(model.request)._result == model.expect._result
        assert SecurityLoad.check_patch_request_data(model.request).error == model.expect.error

    @staticmethod
    def test_patch_request(mocker: MockerFixture, model: PatchRequest):
        mocker.patch.object(SecurityLoad, "SECURITY_LOCK").locked.return_value = model.locked
        mocker.patch.object(SecurityLoad, "check_patch_request_data", return_value=model.check)
        mocker.patch.object(LoginRuleManager, "get_all", return_value=model.get_all)
        mocker.patch.object(SecurityLoad, "record_diff_rule_opera_type")
        mocker.patch.object(LoginRuleManager, "over_write_database", side_effect=model.overwrite)
        assert SecurityLoad().patch_request(model.request) == model.expect

    @staticmethod
    def test_post_request(mocker: MockerFixture, model: PostRequest):
        mocker.patch.object(SecurityLoad, "SECURITY_LOCK").locked.return_value = model.locked
        mocker.patch.object(SecurityLoad, "check_post_request_data", return_value=model.check)
        mocker.patch.object(SecurityLoad, "import_session_cfg_file", return_value=model.import_cfg)
        mocker.patch.object(LoginRuleManager, "get_all", return_value=model.export)
        assert SecurityLoad().post_request(model.request) == model.expect

    @staticmethod
    def test_import_session_cfg_file(mocker: MockerFixture, model: ImportSessionCfgFile):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.link)
        mocker.patch("os.path.getsize", return_value=model.size)
        mocker.patch.object(SecurityLoad, "get_usr_session_sec_cfg", return_value=model.get_cfg)
        mocker.patch.object(SecurityLoad, "check_patch_request_data", return_value=model.check)
        mocker.patch.object(LoginRuleManager, "over_write_database", return_value=model.over_write)
        mock_os_sync(mocker)
        mock_time_sleep(mocker)
        assert SecurityLoad().import_session_cfg_file("")._result == model.expect._result

