# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
import asyncio
import json
import logging
from asyncio import AbstractEventLoop
from collections import namedtuple
from queue import Empty

import pytest
from _pytest.logging import LogCaptureFixture
from pytest_mock import MockFixture
from websockets.legacy.client import WebSocketClientProtocol

from common.checkers import CheckResult
from common.constants.base_constants import MefNetStatus
from common.utils.app_common_method import AppCommonMethod
from lib_restful_adapter import LibRESTfulAdapter
from net_manager.checkers.contents_checker import CertContentsChecker
from net_manager.manager.fd_cfg_manager import FdCfgManager
from wsclient.ws_client_mef import FailedReason, WsClientMef
from wsclient.ws_client_mgr import WsClientMgr


class TestWsClientMef:
    use_cases = {
        "test_is_manage_mef": {
            "not allow": (False, False, None),
            "status is not ok": (False, True, {"status": 400, "message": ""}),
            "ok": (True, True, {"status": 200, "message": {"mef_net_status": MefNetStatus.FD_OM.value}}),
            "not FD_OM": (False, True, {"status": 200, "message": {"mef_net_status": MefNetStatus.MEF.value}}),
        },
        "test_check_mef_port_available": {
            "OK": ("Ip or port of MEF available.", None),
            "ConnectionRefusedError": ("Ip or port of MEF not available", ConnectionRefusedError),
            "Exception": ("Check ip or port of MEF failed, caught exception", Exception),
        },
        "test_exchange_ca_with_mef": {
            "post failed": ("exchange root ca cert failed",
                            [{"status": AppCommonMethod.ERROR, "message": ""}, ],
                            None, None, None, None),
            "get failed": ("get mef root ca data failed",
                           [{"status": AppCommonMethod.OK, "message": ""},
                            {"status": AppCommonMethod.ERROR, "message": ""}, ],
                           None, None, None, None,),
            "not ca_data": ("MEF root ca data is empty",
                            [{"status": AppCommonMethod.OK, "message": ""},
                             {"status": AppCommonMethod.OK, "message": {"mef_ca_data": ""}}, ],
                            None, None, None, None),
            "check_dict failed": ("invalid cert content",
                                  [{"status": AppCommonMethod.OK, "message": ""},
                                   {"status": AppCommonMethod.OK, "message": {"mef_ca_data": "test"}}, ],
                                  CheckResult.make_failed("check failed"), None, None, None),
            "unlink exception": ("unlink file failed",
                                 [{"status": AppCommonMethod.OK, "message": ""},
                                  {"status": AppCommonMethod.OK, "message": {"mef_ca_data": "test"}}, ],
                                 CheckResult.make_success(), True, Exception, None),
            "fdopen exception": ("save mef ca data to file failed",
                                 [{"status": AppCommonMethod.OK, "message": ""},
                                  {"status": AppCommonMethod.OK, "message": {"mef_ca_data": "test"}}, ],
                                 CheckResult.make_success(), False, None, Exception),
            "OK": ("",
                   [{"status": AppCommonMethod.OK, "message": ""},
                    {"status": AppCommonMethod.OK, "message": {"mef_ca_data": "test"}}, ],
                   CheckResult.make_success(), False, None, "123"),
        },
        "test_stop_mef_process": {
            "not manage mef": (True, False, None),
            "post failed": (False, True, [{"status": AppCommonMethod.ERROR, "message": ""}, ]),
            "OK": (True, True, [{"status": AppCommonMethod.OK, "message": ""}, ]),
        },
        "test_restart_mef_process": {
            "post failed": (False, [{"status": AppCommonMethod.ERROR, "message": ""}, ]),
            "OK": (True, [{"status": AppCommonMethod.OK, "message": ""}, ]),
        },
        "test_route_msg_to_mef": {
            "not ready for send msg": ("The websocket connection with MEF is invalid, "
                                       "will stop the route msg to MEF task.",
                                       False, None),
            "CancelledError": ("Websocket client send msg task cancelled.",
                               True, asyncio.CancelledError),
            "Exception": ("Send msg failed, caught exception: ",
                          True, Exception),
        },
        "test_connect": {
            "locked": ("Mef connect task already exist, no need start again.", True, None),
            "cancelled": ("The MEF connect timer has been cancelled.", False, True),
        },
        "test_stop_current_loops": {
            "locked": ("Stop current loops is busy.", True, None, None),
            "not need stop loops": ("No need stop loops, all loops is closed.", False, False, None),
            "Exception": ("Stop current MEF", False, True, Exception),
            "normal": ("Stop current MEF loops done.", False, True, None),
        },
        "test_get_mef_alarm_info": {
            "not ready": ([], False, None),
            "Empty": ([], True, Empty),
            "no alarm": ([], True, None),
            "not json": ([], True, "test"),
            "no content": ([], True, {"1": "test"}),
            "normal": ("alarm test", True, {json.dumps({"content": "{\"alarm\": \"alarm test\"}"})}),
        },
    }

    IsManageMefCase = namedtuple("IsManageMefCase", "expect, is_allow, lib_restful")
    CheckMefPortAvailableCase = namedtuple("CheckMefPortAvailableCase", "expect, create_connection")
    ExchangeCaWithMefCase = namedtuple("ExchangeCaWithMefCase",
                                       "expect, lib_restful, check_dict, islink, unlink, fdopen")
    StopMefProcessCase = namedtuple("StopMefProcessCase", "expect, is_manage_mef, lib_restful")
    RestartMefProcessCase = namedtuple("RestartMefProcessCase", "expect, lib_restful")
    RouteMsgToMefCase = namedtuple("RouteMsgToMefCase", "expect, ready_for_send_msg, wait_for")
    ConnectCase = namedtuple("ConnectCase", "expect, locked, cancelled")
    StopCurrentLoopsCase = namedtuple("StopCurrentLoopsCase", "expect, locked, need_stop_loops, stop_event_loop_now")
    GetMefAlarmInfoCase = namedtuple("GetMefAlarmInfoCase", "expect, ready_for_send_msg, alarm_info")

    @staticmethod
    def test_get_mef_alarm_info(mocker: MockFixture, model: GetMefAlarmInfoCase):
        mocker.patch.object(WsClientMef(), "ready_for_send_msg", return_value=model.ready_for_send_msg)
        mocker.patch("mef_msg_process.msg_que.alarm_que_from_mef.get", side_effect=model.alarm_info)
        assert model.expect == WsClientMef().get_mef_alarm_info()

    @staticmethod
    def test_stop_current_loops(caplog: LogCaptureFixture, mocker: MockFixture, model: StopCurrentLoopsCase):
        caplog.set_level(logging.INFO)
        mocker.patch.object(WsClientMef(), "_STOP_LOCK").locked.return_value = model.locked
        mocker.patch.object(WsClientMef(), "_need_stop_loops", return_value=model.need_stop_loops)
        mocker.patch.object(WsClientMgr, "stop_event_loop_now", side_effect=model.stop_event_loop_now)
        mocker.patch.object(WsClientMef, "ready_for_send_msg", return_value=False)
        WsClientMef().stop_current_loops()
        assert model.expect in [rec.message for rec in caplog.records][-1]

    @staticmethod
    def test_ready_for_send_msg_client_obj_is_false(mocker: MockFixture):
        mocker.patch.object(WsClientMef(), "client_obj", False)
        assert not WsClientMef().ready_for_send_msg()

    @staticmethod
    def test_ready_for_send_msg_client_obj_not_open(mocker: MockFixture):
        mocker.patch.object(WsClientMef(), "client_obj", WebSocketClientProtocol())
        mocker.patch.object(WebSocketClientProtocol, "open", False)
        assert not WsClientMef().ready_for_send_msg()

    @staticmethod
    def test_ready_for_send_msg_connect_loop_is_false(mocker: MockFixture):
        mocker.patch.object(WsClientMef(), "client_obj", WebSocketClientProtocol())
        mocker.patch.object(WebSocketClientProtocol, "open", True)
        mocker.patch.object(WsClientMef(), "connect_loop", False)
        assert not WsClientMef().ready_for_send_msg()

    @staticmethod
    def test_ready_for_send_msg_connect_loop_is_closed(mocker: MockFixture):
        mocker.patch.object(WsClientMef(), "client_obj", WebSocketClientProtocol())
        mocker.patch.object(WebSocketClientProtocol, "open", True)
        mocker.patch.object(WsClientMef(), "connect_loop", AbstractEventLoop())
        mocker.patch.object(AbstractEventLoop, "is_closed", return_value=True)
        assert not WsClientMef().ready_for_send_msg()

    @staticmethod
    def test_ready_for_send_msg(mocker: MockFixture):
        mocker.patch.object(WsClientMef(), "client_obj", WebSocketClientProtocol())
        mocker.patch.object(WebSocketClientProtocol, "open", True)
        mocker.patch.object(WsClientMef(), "connect_loop", AbstractEventLoop())
        mocker.patch.object(AbstractEventLoop, "is_closed", return_value=False)
        assert WsClientMef().ready_for_send_msg()

    @staticmethod
    def test_connect(caplog: LogCaptureFixture, mocker: MockFixture, model: ConnectCase):
        caplog.set_level(logging.INFO)
        mocker.patch.object(WsClientMef(), "MEF_CONNECT_LOCK").locked.return_value = model.locked
        mocker.patch.object(WsClientMef(), "cancelled", model.cancelled)
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch.object(WsClientMef(), "ready_for_send_msg", return_value=True)
        mocker.patch.object(FdCfgManager, "check_fd_mode_and_status_ready", return_value=True)
        mocker.patch.object(WsClientMef(), "ready_for_send_msg", return_value=False)
        mocker.patch.object(WsClientMef(), "check_mef_port_available", return_value=True)
        mocker.patch.object(WsClientMef(), "restart_mef_process", return_value=True)
        mocker.patch.object(WsClientMef(), "exchange_ca_with_mef", return_value=True)
        mocker.patch.object(WsClientMef(), "connect_loop")
        WsClientMef().connect()
        assert model.expect in [rec.message for rec in caplog.records][-1]

    @staticmethod
    @pytest.mark.asyncio
    async def test_route_msg_to_mef(caplog: LogCaptureFixture, mocker: MockFixture, model: RouteMsgToMefCase):
        caplog.set_level(logging.INFO)
        mocker.patch.object(WsClientMef(), "cancelled", False)
        mocker.patch.object(WsClientMef(), "ready_for_send_msg", return_value=model.ready_for_send_msg)
        mocker.patch("mef_msg_process.msg_que.msg_que_to_mef.get_nowait", return_value="test")
        mocker.patch.object(WebSocketClientProtocol, "send", side_effect=[None, None])
        mocker.patch.object(asyncio, "wait_for", side_effect=model.wait_for)
        await WsClientMef.route_msg_to_mef(WebSocketClientProtocol())
        assert model.expect in [rec.message for rec in caplog.records][-1]

    @staticmethod
    def test_restart_mef_process(mocker: MockFixture, model: RestartMefProcessCase):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_restful)
        assert bool(WsClientMef.restart_mef_process()) == model.expect

    @staticmethod
    def test_stop_mef_process(mocker: MockFixture, model: StopMefProcessCase):
        mocker.patch.object(WsClientMef, "is_manage_mef", return_value=model.is_manage_mef)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_restful)
        assert bool(WsClientMef.stop_mef_process()) == model.expect

    @staticmethod
    def test_exchange_ca_with_mef(mocker: MockFixture, model: ExchangeCaWithMefCase):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", side_effect=model.lib_restful)
        mocker.patch.object(CertContentsChecker, "check_dict", return_value=model.check_dict)
        mocker.patch("os.path.islink", return_value=model.islink)
        mocker.patch("os.unlink", side_effect=model.unlink)
        mocker.patch("os.open")
        mocker.patch("os.fdopen").return_value.__enter__.return_value.write.side_effect = model.fdopen
        assert WsClientMef.exchange_ca_with_mef().error == model.expect

    @staticmethod
    def test_check_mef_port_available(caplog: LogCaptureFixture, mocker: MockFixture, model: CheckMefPortAvailableCase):
        caplog.set_level(logging.INFO)
        mocker.patch("wsclient.ws_client_mef.create_connection").side_effect = model.create_connection
        WsClientMef.check_mef_port_available()
        assert model.expect in [rec.message for rec in caplog.records][-1]

    @staticmethod
    def test_is_manage_mef(mocker: MockFixture, model: IsManageMefCase):
        mocker.patch("wsclient.ws_client_mef.is_allow", return_value=model.is_allow)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_restful)
        assert WsClientMef.is_manage_mef() == model.expect


class TestFailedReason:

    @staticmethod
    def test_enum():
        assert FailedReason.EMPTY.value == ""
        assert FailedReason.ERR_INVALID_SSL_CONTEXT.value == "invalid ssl context"
        assert FailedReason.ERR_CERT_VERIFIED_FAILED.value == "certification verify failed"
        assert FailedReason.ERR_EXCHANGE_CERT_FAILED.value == "exchange mef cert failed"
