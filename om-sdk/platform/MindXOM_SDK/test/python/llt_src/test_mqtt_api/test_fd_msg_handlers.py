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

from concurrent.futures.thread import ThreadPoolExecutor

from pytest_mock import MockerFixture

from fd_msg_process.fd_msg_handlers import FDMessageHandler, FdMsgData
from fd_msg_process.midware_route import MidwareRoute
from fd_msg_process.midware_urls import MidwareUris
from common.utils.result_base import Result
from mef_msg_process.mef_proc import MefProc
from wsclient.ws_monitor import WsMonitor


class TestUtils1:
    topic = "$hw/edge/v1/hardware/operate/restart"

    @staticmethod
    def clear():
        return ""


class TestUtils:
    data = [0, 0]
    topic = "$hw/edge/v1/hardware/operate/config_hostname"
    up_resource = False

    @staticmethod
    def deal_func_name():
        return False

    @staticmethod
    def error():
        return ""

    @staticmethod
    def payload():
        return ""

    @staticmethod
    def content():
        return ""

    @staticmethod
    def empty():
        return ""
    queue = TestUtils1


class TestGetJsonInfoObj:
    def test_check_ack_response(self):
        assert not FDMessageHandler.check_ack_response(TestUtils)

    def test_handle_config_net_manager_task_first(self, mocker: MockerFixture):
        mocker.patch.object(FDMessageHandler, "call_view_func", side_effect=[Result(True), ])
        mocker.patch.object(FdMsgData, "gen_ws_msg_obj", side_effect=[Result(True), ])
        assert FDMessageHandler.handle_config_net_manager_task(TestUtils)

    def test_handle_config_net_manager_task(self, mocker: MockerFixture):
        mocker.patch.object(FDMessageHandler, "call_view_func", side_effect=[TestUtils(), ])
        mocker.patch.object(FdMsgData, "gen_ws_msg_obj", side_effect=[Result(True), ])
        mocker.patch.object(MefProc, "stop_mef_connect_timer", side_effect=[Result(True), ])
        mocker.patch.object(WsMonitor, "start_fd_connect_monitor", side_effect=[Result(True), ])
        mocker.patch.object(MefProc, "start_mef_connect_timer", side_effect=[Result(True), ])
        mocker.patch.object(FDMessageHandler, "clear_fd_msg_queue", side_effect=[TestUtils(), ])
        assert FDMessageHandler.handle_config_net_manager_task(TestUtils) == [0, ""]

    def test_handle_msg_from_fd_by_mqtt_first(self, mocker: MockerFixture):
        mocker.patch.object(ThreadPoolExecutor, "submit", side_effect=Exception())
        assert not FDMessageHandler.handle_msg_from_fd_by_mqtt(TestUtils1)

    def test_handle_msg_from_fd_by_mqtt(self, mocker: MockerFixture):
        mocker.patch.object(FDMessageHandler, "call_func_and_report", side_effect=[Result(True), ])
        assert not FDMessageHandler.handle_msg_from_fd_by_mqtt(TestUtils)

    def test_handle_upgrade_service(self, mocker: MockerFixture):
        mocker.patch.object(FDMessageHandler, "call_func_and_report", side_effect=[Result(True), ])
        assert not FDMessageHandler.handle_upgrade_service(TestUtils)

    def test_call_view_func_except(self):
        assert not FDMessageHandler.call_view_func("a", "")

    def test_call_view_func(self, mocker: MockerFixture):
        mocker.patch.object(MidwareRoute, "view_functions", side_effect=[Result(True), ])
        assert FDMessageHandler.call_view_func("a", "")

    def test_handle_restart_operation_task(self, mocker: MockerFixture):
        mocker.patch.object(MidwareRoute, "view_functions", side_effect=[Result(True), ])
        mocker.patch.object(FdMsgData, "gen_ws_msg_obj", side_effect=[Result(True), Result(True), ])
        assert not FDMessageHandler.handle_restart_operation_task(TestUtils)

    def test_handle_profile_effect_task(self, mocker: MockerFixture):
        mocker.patch.object(MidwareRoute, "view_functions", side_effect=[Result(True), ])
        mocker.patch.object(FdMsgData, "gen_ws_msg_obj", side_effect=[Result(True), Result(True), ])
        assert not FDMessageHandler.handle_profile_effect_task(TestUtils)

    def test_handle_reset_alarm_from_fd_first(self, mocker: MockerFixture):
        mocker.patch("json.loads", side_effect=Exception())
        assert not FDMessageHandler.handle_reset_alarm_from_fd(TestUtils)

    def test_handle_reset_alarm_from_fd_second(self, mocker: MockerFixture):
        mocker.patch("json.loads", side_effect=[Result(True), ])
        mocker.patch.object(FDMessageHandler, "call_view_func", side_effect=[TestUtils(), ])
        assert not FDMessageHandler.handle_reset_alarm_from_fd(TestUtils)

    def test_handle_reset_alarm_from_fd(self, mocker: MockerFixture):
        mocker.patch("json.loads", side_effect=[Result(True), ])
        mocker.patch.object(FDMessageHandler, "call_view_func", side_effect=[Result(False), ])
        assert not FDMessageHandler.handle_reset_alarm_from_fd(TestUtils)

    def test_handle_info_collect_progress_task(self, mocker: MockerFixture):
        mocker.patch.object(MidwareRoute, "view_functions", side_effect=[Result(True), ])
        mocker.patch.object(FdMsgData, "gen_ws_msg_obj", side_effect=[Result(True), ])
        assert not FDMessageHandler.handle_info_collect_progress_task(TestUtils)

    def test_handle_firmware_effective_task(self, mocker: MockerFixture):
        mocker.patch.object(MidwareRoute, "view_functions", side_effect=[Result(True), ])
        mocker.patch.object(FdMsgData, "gen_ws_msg_obj", side_effect=[Result(True), ])
        assert not FDMessageHandler.handle_firmware_effective_task(TestUtils)

    def test_handle_ntp_sync_task_first(self):
        assert not FDMessageHandler.handle_ntp_sync_task()

    def test_handle_ntp_sync_task(self, mocker: MockerFixture):
        mocker.patch.object(MidwareUris, "snyc_ntp_config", side_effect=[["1", "1"]] * 10)
        assert not FDMessageHandler.handle_ntp_sync_task()

    def test_call_func_and_report_first(self):
        assert not FDMessageHandler.call_func_and_report("", TestUtils1)

    def test_call_func_and_report(self, mocker: MockerFixture):
        mocker.patch.object(FDMessageHandler, "call_view_func", side_effect=[TestUtils(), ])
        mocker.patch.object(FdMsgData, "gen_ws_msg_obj", side_effect=[Result(True), ])
        assert not FDMessageHandler.call_func_and_report("1", TestUtils)

    def test_call_func_and_report_second(self, mocker: MockerFixture):
        mocker.patch.object(FDMessageHandler, "call_view_func", side_effect=[TestUtils(), ])
        mocker.patch.object(FdMsgData, "gen_ws_msg_obj", side_effect=[Result(True), ])
        assert not FDMessageHandler.call_func_and_report("1", TestUtils)

    def test_clear_fd_msg_queue(self):
        assert not FDMessageHandler.clear_fd_msg_queue([TestUtils, ])
