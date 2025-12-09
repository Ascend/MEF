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
import asyncio
import json
import os
import time
from asyncio import CancelledError
from collections import namedtuple
from concurrent import futures
from datetime import datetime
from tempfile import NamedTemporaryFile

import pytest
from mock.mock import patch
from pytest_mock import MockFixture
from websockets.legacy.client import WebSocketClientProtocol

from common.db.database import DataBase
from common.db.migrate import Migrate
from common.utils.date_utils import DateUtils
from common.utils.result_base import Result
from fd_msg_process.midware_route import MidwareRoute
from net_manager.manager.net_cfg_manager import NetCfgManager
from redfish_db.init_structure import INIT_COLUMNS
from test_mqtt_api.get_log_info import GetLogInfo
from user_manager import user_manager
from user_manager.models import User
from wsclient.ws_client_mgr import WsClientMgr
from wsclient.ws_info_reporter import SystemInfoHandler, HeartBeatInfoHandle

getLog = GetLogInfo()


@pytest.fixture(scope="module")
def database() -> DataBase:
    with NamedTemporaryFile(suffix=".db") as tmp_file:
        db_path = tmp_file.name
    type("TestMigrate", (Migrate,), {"models": {User.__tablename__: User}})\
        .execute_on_install(db_path, INIT_COLUMNS)
    test_db = type("TestDB", (DataBase,), {})(db_path)
    with patch.object(user_manager, "session_maker", test_db.session_maker):
        yield test_db
    os.remove(db_path)


class TestSystemInfoHandler:
    use_cases = {
        "test_ws_send_sys_info": {
            "normal": ("", [None, ]),
            "CancelledError": ("Send sys info task cancelled.", CancelledError),
            "Exception": ("Send sys info failed, caught exception:", Exception),
        },
        "test_ws_send_sys_status": {
            "normal": ("", [None, ]),
            "CancelledError": ("send sys status task canceled", CancelledError),
            "Exception": ("send sys status failed, caught exception:", Exception),
        },
        "test_ws_send_alarm": {
            "normal": ("", [None, ]),
            "CancelledError": ("send alarm info task canceled", CancelledError),
            "Exception": ("send alarm info failed, caught exception:", Exception),
        },
        "test_ws_send_event": {
            "normal": ("", [None, ]),
            "CancelledError": ("send event info task canceled", CancelledError),
            "Exception": ("send event info failed, caught exception:", Exception),
        },
        "test_route_msg_to_fd": {
            "CancelledError": ("Websocket client send msg task cancelled.", futures.CancelledError),
            "Exception": ("Send msg failed, caught exception: ", Exception),
        },
        "test_send_sys_info": {
            "closed": (Result(result=False), None, True, None),
            "normal": (Result(result=True), "test", False, None)
        },
        "test_send_sys_status": {
            "closed": (Result(result=False), None, True, None),
            "normal": (Result(result=True), "test", False, None)
        },
        "test_send_alarm": {
            "closed": (Result(result=False), None, True, None),
            "normal": (Result(result=True), "test", False, None)
        },
        "test_send_event": {
            "closed": (Result(result=False), None, True, None),
            "normal": (Result(result=True), "test", False, None)
        },
        "test_get_sys_info": {
            "not ready": ("{}", "test", [None, None]),
            "ready": ('{"test": "test"}', "ready", [None, {"test": "test"}]),
        },
        "test_get_alarm": {
            "ret not list": ('{"alarm": []}', "test"),
            "ret[0] not 0": ('{"alarm": []}', [1, "test"]),
            "ret is none": ('{"alarm": []}', None),
            "normal": ('{"alarm": ["test"]}', [0, {"alarm": ["test"]}]),
        },
    }

    WsSendSysStatusCase = namedtuple("WsSendSysStatusCase", "expect, send_sys_status")
    WsSendSysInfoCase = namedtuple("WsSendSysInfoCase", "expect, send_sys_info")
    WsSendAlarmCase = namedtuple("WsSendAlarmCase", "expect, send_alarm")
    WsSendEventCase = namedtuple("WsSendEventCase", "expect, send_event")
    RouteMsgToFdCase = namedtuple("RouteMsgToFdCase", "expect, wait_for")
    SendSysInfoCase = namedtuple("SendSysInfoCase", "expect, get_sys_info, closed, wait_for")
    SendSysStatusCase = namedtuple("SendSysStatusCase", "expect, get_sys_status, closed, wait_for")
    SendAlarmCase = namedtuple("SendAlarmCase", "expect, get_alarm, closed, wait_for")
    SendEventCase = namedtuple("SendEventCase", "expect, get_event, closed, wait_for")
    GetSysInfoCase = namedtuple("GetSysInfoCase", "expect, status, view_functions")
    GetAlarmCase = namedtuple("GetAlarmCase", "expect, view_functions")

    @staticmethod
    def test_get_alarm(mocker: MockFixture, model: GetAlarmCase):
        mocker.patch.object(MidwareRoute, "view_functions").__getitem__.return_value.return_value = model.view_functions
        assert json.loads(SystemInfoHandler._get_alarm()).get("content") == model.expect

    @staticmethod
    def test_get_sys_info(mocker: MockFixture, model: GetSysInfoCase):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info").return_value.status = model.status
        mocker.patch.object(MidwareRoute, "view_functions").__getitem__.return_value.return_value = model.view_functions
        assert json.loads(SystemInfoHandler._get_sys_info()).get("content") == model.expect

    @staticmethod
    @pytest.mark.asyncio
    async def test_send_event(mocker: MockFixture, model: SendEventCase):
        mocker.patch.object(SystemInfoHandler, "_get_event", return_value=model.get_event)
        mocker.patch.object(WebSocketClientProtocol, "closed", model.closed)
        mocker.patch.object(WebSocketClientProtocol, "send")
        mocker.patch.object(asyncio, "wait_for", side_effect=model.wait_for)
        mocker.patch.object(WsClientMgr, "sleep_for_a_while")
        result = await SystemInfoHandler._send_event(WebSocketClientProtocol())
        assert bool(result) == bool(model.expect)

    @staticmethod
    @pytest.mark.asyncio
    async def test_send_alarm(mocker: MockFixture, model: SendAlarmCase):
        mocker.patch.object(SystemInfoHandler, "_get_alarm", return_value=model.get_alarm)
        mocker.patch.object(WebSocketClientProtocol, "closed", model.closed)
        mocker.patch.object(WebSocketClientProtocol, "send")
        mocker.patch.object(asyncio, "wait_for", side_effect=model.wait_for)
        mocker.patch.object(WsClientMgr, "sleep_for_a_while")
        result = await SystemInfoHandler._send_alarm(WebSocketClientProtocol())
        assert bool(result) == bool(model.expect)

    @staticmethod
    @pytest.mark.asyncio
    async def test_send_sys_status(mocker: MockFixture, model: SendSysStatusCase):
        mocker.patch.object(SystemInfoHandler, "_get_sys_status", return_value=model.get_sys_status)
        mocker.patch.object(WebSocketClientProtocol, "closed", model.closed)
        mocker.patch.object(WebSocketClientProtocol, "send")
        mocker.patch.object(asyncio, "wait_for", side_effect=model.wait_for)
        mocker.patch.object(WsClientMgr, "sleep_for_a_while")
        result = await SystemInfoHandler._send_sys_status(WebSocketClientProtocol())
        assert bool(result) == bool(model.expect)

    @staticmethod
    @pytest.mark.asyncio
    async def test_send_sys_info(mocker: MockFixture, model: SendSysInfoCase):
        mocker.patch.object(SystemInfoHandler, "_get_sys_info", return_value=model.get_sys_info)
        mocker.patch.object(WebSocketClientProtocol, "closed", model.closed)
        mocker.patch.object(WebSocketClientProtocol, "send")
        mocker.patch.object(asyncio, "wait_for", side_effect=model.wait_for)
        mocker.patch.object(WsClientMgr, "sleep_for_a_while")
        result = await SystemInfoHandler._send_sys_info(WebSocketClientProtocol())
        assert bool(result) == bool(model.expect)

    @staticmethod
    async def route_msg_to_fd_loop():
        task = asyncio.create_task(SystemInfoHandler.route_msg_to_fd(WebSocketClientProtocol()))
        await task

    @staticmethod
    def test_route_msg_to_fd(mocker: MockFixture, model: RouteMsgToFdCase):
        mocker.patch.object(WsClientMgr, "ready_for_send_msg", return_value=True)
        mocker.patch("mef_msg_process.msg_que.msg_que_from_mef.get_nowait", return_value="test")
        mocker.patch.object(WebSocketClientProtocol, "send")
        mocker.patch.object(asyncio, "wait_for", side_effect=model.wait_for)
        getLog.clear_log()
        asyncio.run(TestSystemInfoHandler.route_msg_to_fd_loop())
        assert model.expect in getLog.get_log()

    @staticmethod
    async def ws_send_event_loop():
        task = asyncio.create_task(SystemInfoHandler.ws_send_event(WebSocketClientProtocol()))
        await task

    @staticmethod
    def test_ws_send_event(mocker: MockFixture, model: WsSendEventCase):
        mocker.patch.object(SystemInfoHandler, "_send_event", side_effect=model.send_event)
        getLog.clear_log()
        asyncio.run(TestSystemInfoHandler.ws_send_event_loop())
        assert model.expect in getLog.get_log()

    @staticmethod
    async def ws_send_alarm_loop():
        task = asyncio.create_task(SystemInfoHandler.ws_send_alarm(WebSocketClientProtocol()))
        await task

    @staticmethod
    def test_ws_send_alarm(mocker: MockFixture, model: WsSendAlarmCase):
        mocker.patch.object(SystemInfoHandler, "_send_alarm", side_effect=model.send_alarm)
        getLog.clear_log()
        asyncio.run(TestSystemInfoHandler.ws_send_alarm_loop())
        assert model.expect in getLog.get_log()

    @staticmethod
    async def ws_send_sys_status_loop():
        task = asyncio.create_task(SystemInfoHandler.ws_send_sys_status(WebSocketClientProtocol()))
        await task

    @staticmethod
    def test_ws_send_sys_status(mocker: MockFixture, model: WsSendSysStatusCase):
        mocker.patch.object(SystemInfoHandler, "_send_sys_status", side_effect=model.send_sys_status)
        getLog.clear_log()
        asyncio.run(TestSystemInfoHandler.ws_send_sys_status_loop())
        assert model.expect in getLog.get_log()

    @staticmethod
    async def ws_send_sys_info_loop():
        task = asyncio.create_task(SystemInfoHandler.ws_send_sys_info(WebSocketClientProtocol()))
        await task

    @staticmethod
    def test_ws_send_sys_info(mocker: MockFixture, model: WsSendSysInfoCase):
        mocker.patch.object(SystemInfoHandler, "_send_sys_info", side_effect=model.send_sys_info)
        getLog.clear_log()
        asyncio.run(TestSystemInfoHandler.ws_send_sys_info_loop())
        assert model.expect in getLog.get_log()

    @staticmethod
    def test_get_event_beyond_change_pwd_event_max_days(database: DataBase, mocker: MockFixture):
        mocker.patch.object(DateUtils, "get_format_time",
                            return_value=time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(1707523200.0)))
        with database.session_maker() as session:
            session.query(User).delete()
            session.add(User(
                username_db="admin",
                pword_hash="test",
                create_time="2024-01-01 00:00:00",
                pword_modify_time="2024-01-02 00:00:00",
                pword_wrong_times=0,
                start_lock_time=0,
                last_login_success_time="",
                last_login_failure_time="",
                account_insecure_prompt=True,
                config_navigator_prompt=True,
                log_in_time="",
                lock_state=False,
                enabled=True,
                role_id=1,
            ))
        expect = {
            'alarm': [
                {
                    'type': 'event',
                    'alarmId': '0x01000006',
                    'alarmName': 'Change the web password',
                    'resource': 'system',
                    'perceivedSeverity': 'MAJOR',
                    'timestamp': '2024-02-10T00:00:00+00:00',
                    'notificationType': '',
                    'detailedInformation': 'The web password is not changed, please change it.',
                    'suggestion': 'Log in to the Atlas 500 WebUI and change the web password.',
                    'reason': '',
                    'impact': ''
                }
            ]
        }
        assert json.loads(json.loads(SystemInfoHandler._get_event()).get("content")) == expect

    @staticmethod
    def test_get_event_within_change_pwd_event_max_days(database: DataBase, mocker: MockFixture):
        mocker.patch.object(DateUtils, "get_time",
                            return_value=datetime.strptime("2024-01-23 00:00:00", "%Y-%m-%d %H:%M:%S"))
        with database.session_maker() as session:
            session.query(User).delete()
            session.add(User(
                username_db="admin",
                pword_hash="test",
                create_time="2024-01-01 00:00:00",
                pword_modify_time="2024-01-02 00:00:00",
                pword_wrong_times=0,
                start_lock_time=0,
                last_login_success_time="",
                last_login_failure_time="",
                account_insecure_prompt=True,
                config_navigator_prompt=True,
                log_in_time="",
                lock_state=False,
                enabled=True,
                role_id=1,
            ))
        expect = {"alarm": []}
        assert json.loads(json.loads(SystemInfoHandler._get_event()).get("content")) == expect

    @staticmethod
    def test_get_event_without_userinfo(database: DataBase, mocker: MockFixture):
        mocker.patch.object(DateUtils, "get_format_time",
                            return_value=time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(1707523200.0)))
        with database.session_maker() as session:
            session.query(User).delete()
            session.add(User(
                username_db="admin",
                pword_hash="test",
                create_time="2024-01-01 00:00:00",
                pword_modify_time="2024-01-02 00:00:00",
                pword_wrong_times=0,
                start_lock_time=0,
                last_login_success_time="",
                last_login_failure_time="",
                account_insecure_prompt=False,
                config_navigator_prompt=True,
                log_in_time="",
                lock_state=False,
                enabled=True,
                role_id=1,
            ))
        expect = {"alarm": []}
        assert json.loads(json.loads(SystemInfoHandler._get_event()).get("content")) == expect


class TestHeartBeatInfoHandle:
    use_cases = {
        "test_ws_send_keepalive": {
            "normal": ("", [None, ]),
            "CancelledError": ("send keep info task canceled", CancelledError),
            "Exception": ("send alarm info failed, caught exception:", Exception),
        },
        "test_send_keepalive": {
            "closed": (Result(result=False), None, True, None),
            "normal": (Result(result=True), "test", False, None)
        },
    }

    WsSendKeepaliveCase = namedtuple("WsSendKeepaliveCase", "expect, send_keepalive")
    SendKeepaliveCase = namedtuple("SendKeepaliveCase", "expect, get_keepalive, closed, wait_for")

    @staticmethod
    @pytest.mark.asyncio
    async def test_send_keepalive(mocker: MockFixture, model: SendKeepaliveCase):
        mocker.patch.object(HeartBeatInfoHandle, "_get_keepalive", return_value=model.get_keepalive)
        mocker.patch.object(WebSocketClientProtocol, "closed", model.closed)
        mocker.patch.object(WebSocketClientProtocol, "send")
        mocker.patch.object(asyncio, "wait_for", side_effect=model.wait_for)
        mocker.patch.object(WsClientMgr, "sleep_for_a_while")
        result = await HeartBeatInfoHandle._send_keepalive(WebSocketClientProtocol())
        assert bool(result) == bool(model.expect)

    @staticmethod
    async def ws_send_keepalive_loop():
        task = asyncio.create_task(HeartBeatInfoHandle.ws_send_keepalive(WebSocketClientProtocol()))
        await task

    @staticmethod
    def test_ws_send_keepalive(mocker: MockFixture, model: WsSendKeepaliveCase):
        mocker.patch.object(HeartBeatInfoHandle, "_send_keepalive", side_effect=model.send_keepalive)
        getLog.clear_log()
        asyncio.run(TestHeartBeatInfoHandle.ws_send_keepalive_loop())
        assert model.expect in getLog.get_log()

    @staticmethod
    def test_get_keepalive():
        assert json.loads(HeartBeatInfoHandle._get_keepalive()).get("content") == "ping"
