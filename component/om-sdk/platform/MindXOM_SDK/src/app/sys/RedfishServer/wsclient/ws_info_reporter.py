# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

import asyncio
import time
from asyncio import CancelledError
from concurrent import futures
from datetime import datetime
from datetime import timezone
from queue import Empty

from common.utils.result_base import Result
from websockets.legacy.client import WebSocketClientProtocol

from common.log.logger import run_log
from common.utils.date_utils import DateUtils
from fd_msg_process.midware_route import MidwareRoute
from mef_msg_process.msg_que import msg_que_from_mef
from net_manager.manager.fd_cfg_manager import FdMsgData
from net_manager.manager.net_cfg_manager import NetCfgManager
from user_manager.user_manager import UserManager
from wsclient.connect_status import FdConnectStatus
from wsclient.ws_client_mgr import WsClientMgr


class SystemInfoHandler:
    """信息主动上报"""
    CHANGE_PWD_EVENT_MAX_DAYS = 30  # 修改密码事件, 超过30天未修改密码, 则上报
    CHANGE_PWD_EVENT_REPORT_INTERVAL = 43200  # 修改密码事件上报周期, 每12个小时检查上报

    @staticmethod
    async def ws_send_sys_info(client: WebSocketClientProtocol):
        """
        系统静态信息周期性主动上报协程
        :param client: websocket客户端
        :return: None
        """
        while True:
            try:
                ret = await SystemInfoHandler._send_sys_info(client)
            except CancelledError:
                run_log.warning("Send sys info task cancelled.")
                WsClientMgr().connected = False
                FdConnectStatus().trans_to_connecting()
                return
            except Exception as err:
                run_log.error("Send sys info failed, caught exception: %s", err)
                WsClientMgr().connected = False
                FdConnectStatus().trans_to_connecting()
                return

            if not ret:
                return

    @staticmethod
    async def ws_send_sys_status(client: WebSocketClientProtocol):
        """
        系统状态信息周期性主动上报协程
        :param client: websocket客户端
        :return: None
        """
        while True:
            try:
                ret = await SystemInfoHandler._send_sys_status(client)
            except CancelledError:
                run_log.warning("send sys status task canceled")
                WsClientMgr().connected = False
                FdConnectStatus().trans_to_connecting()
                return
            except Exception as err:
                run_log.error("send sys status failed, caught exception: %s", err)
                WsClientMgr().connected = False
                FdConnectStatus().trans_to_connecting()
                return

            if not ret:
                return

    @staticmethod
    async def ws_send_alarm(client: WebSocketClientProtocol):
        """
        告警周期性主动上报协程
        :param client: websocket客户端
        :return: None
        """
        while True:
            try:
                ret = await SystemInfoHandler._send_alarm(client)
            except CancelledError:
                run_log.warning("send alarm info task canceled")
                WsClientMgr().connected = False
                FdConnectStatus().trans_to_connecting()
                return
            except Exception as err:
                run_log.error("send alarm info failed, caught exception: %s", err)
                WsClientMgr().connected = False
                FdConnectStatus().trans_to_connecting()
                return

            if not ret:
                FdConnectStatus().trans_to_connecting()
                return

    @staticmethod
    async def ws_send_event(client: WebSocketClientProtocol):
        """
        告警周期性主动上报事件
        :param client: websocket客户端
        :return: None
        """
        while True:
            try:
                ret = await SystemInfoHandler._send_event(client)
            except CancelledError:
                run_log.warning("send event info task canceled")
                WsClientMgr().connected = False
                FdConnectStatus().trans_to_connecting()
                return
            except Exception as err:
                run_log.error("send event info failed, caught exception: %s", err)
                WsClientMgr().connected = False
                FdConnectStatus().trans_to_connecting()
                return

            if not ret:
                FdConnectStatus().trans_to_connecting()
                return

    @staticmethod
    async def route_msg_to_fd(client: WebSocketClientProtocol):
        """将MEF返回的消息转发给FD"""
        while True:
            if not WsClientMgr().ready_for_send_msg():
                await asyncio.sleep(1)
                continue

            try:
                msg = msg_que_from_mef.get_nowait()
            except Empty:
                await asyncio.sleep(0.5)
                continue

            try:
                await asyncio.wait_for(client.send(msg), timeout=WsClientMgr.SEND_TIMEOUT)
            except futures.TimeoutError:
                run_log.warning("Websocket client send msg time out.")
                continue
            except futures.CancelledError:
                run_log.warning("Websocket client send msg task cancelled.")
                return Result(result=False, err_msg="websocket client send msg task cancelled")
            except Exception as err:
                run_log.error("Send msg failed, caught exception: %s, type is %s.", err, type(err))
                return Result(result=False,
                              err_msg=f"send response failed, caught exception: {err}, type is {type(err)}")

    @staticmethod
    async def _send_sys_info(client: WebSocketClientProtocol):
        sys_info = SystemInfoHandler._get_sys_info()
        if client.closed:
            run_log.error("websocket client connection is closed, can not send msg now")
            return Result(result=False)

        try:
            await asyncio.wait_for(client.send(sys_info), timeout=WsClientMgr.SEND_TIMEOUT)
        except futures.TimeoutError:
            run_log.warning("Send sys info time out.")

        await WsClientMgr.sleep_for_a_while(120)

        return Result(result=True)

    @staticmethod
    async def _send_sys_status(client: WebSocketClientProtocol):
        sys_status = SystemInfoHandler._get_sys_status()
        if client.closed:
            run_log.error("websocket client connection is closed, can not send msg now")
            return Result(result=False)

        try:
            await asyncio.wait_for(client.send(sys_status), timeout=WsClientMgr.SEND_TIMEOUT)
        except futures.TimeoutError:
            run_log.warning("Send sys status time out.")

        await WsClientMgr.sleep_for_a_while(60)
        return Result(result=True)

    @staticmethod
    async def _send_alarm(client: WebSocketClientProtocol):
        alarm_info = SystemInfoHandler._get_alarm()
        if client.closed:
            run_log.error("websocket client connection is closed, can not send msg now")
            return Result(result=False)

        try:
            await asyncio.wait_for(client.send(alarm_info), timeout=WsClientMgr.SEND_TIMEOUT)
        except futures.TimeoutError:
            run_log.warning("Send alarm info time out.")

        await WsClientMgr.sleep_for_a_while(60)
        return Result(result=True)

    @staticmethod
    async def _send_event(client: WebSocketClientProtocol):
        """
        告警周期性主动上报事件
        :param client: websocket客户端
        :return: None
        """
        # 延时5分钟，保证重纳管场景首次上报时，能在FD上展示此事件
        await WsClientMgr.sleep_for_a_while(300)
        event_info = SystemInfoHandler._get_event()
        if client.closed:
            run_log.error("websocket client connection is closed, can not send msg now")
            return Result(result=False)

        try:
            await asyncio.wait_for(client.send(event_info), timeout=WsClientMgr.SEND_TIMEOUT)
        except futures.TimeoutError:
            run_log.warning("Send event info time out.")

        await WsClientMgr.sleep_for_a_while(SystemInfoHandler.CHANGE_PWD_EVENT_REPORT_INTERVAL)
        return Result(result=True)

    @staticmethod
    def _get_sys_info():
        """系统静态信息 - 连接ws后必传"""
        sys_info = {}
        # 一机一密下发后，状态变为ready后才上报具体的系统信息
        if NetCfgManager().get_net_cfg_info().status == "ready":
            sys_info = MidwareRoute.view_functions["espmanager/SysInfo"]()[1]
        msg = FdMsgData.gen_ws_msg_obj(sys_info, "websocket/sys_info")
        return msg.to_ws_msg_str()

    @staticmethod
    def _get_sys_status():
        """系统状态信息 - 连接ws后必传"""
        sys_status = {}
        # 一机一密下发后，状态变为ready后才上报具体的系统信息
        if NetCfgManager().get_net_cfg_info().status == "ready":
            sys_status = MidwareRoute.view_functions["espmanager/SysStatus"]()[1]
        msg = FdMsgData.gen_ws_msg_obj(sys_status, "websocket/sys_status")
        return msg.to_ws_msg_str()

    @staticmethod
    def _get_alarm():
        """告警信息"""
        alarm_data = {"alarm": []}
        ret = MidwareRoute.view_functions["espmanager/Alarm"]("payload")
        ret_valid = isinstance(ret, list) and ret and ret[0] == 0
        if ret_valid:
            alarm_data = ret[1]
        else:
            # 获取alarm失败，记录日志，继续上报空消息
            run_log.error("get alarm info failed")

        msg = FdMsgData.gen_ws_msg_obj(alarm_data, "websocket/alarm")
        return msg.to_ws_msg_str()

    @staticmethod
    def _get_event():
        """事件信息"""
        msg_dict = {"alarm": []}
        user_info = UserManager.find_user_by_id(1)
        if not user_info or not user_info.account_insecure_prompt:
            run_log.info("No events detected.")
            msg = FdMsgData.gen_ws_msg_obj(msg_dict, "websocket/alarm")
            return msg.to_ws_msg_str()

        pwd_modify_time = DateUtils.get_time(user_info.pword_modify_time)
        current_time = DateUtils.get_time(DateUtils.get_format_time(time.time()))
        if abs((current_time - pwd_modify_time).days) < SystemInfoHandler.CHANGE_PWD_EVENT_MAX_DAYS:
            run_log.info("No events detected.")
            msg = FdMsgData.gen_ws_msg_obj(msg_dict, "websocket/alarm")
            return msg.to_ws_msg_str()

        time_stamp = round(current_time.timestamp())
        event_msg = {
            "type": "event",
            "alarmId": "0x01000006",
            "alarmName": "Change the web password",
            "resource": "system",
            "perceivedSeverity": "MAJOR",
            "timestamp": datetime.utcfromtimestamp(time_stamp).replace(microsecond=0, tzinfo=timezone.utc).isoformat(),
            "notificationType": "",
            "detailedInformation": "The web password is not changed, please change it.",
            "suggestion": "Log in to the Atlas 500 WebUI and change the web password.",
            "reason": "",
            "impact": "",
        }
        run_log.info("Check the default cipher is not changed for a long time, start report event to FD.")
        msg = FdMsgData.gen_ws_msg_obj({"alarm": [event_msg]}, "websocket/alarm")
        return msg.to_ws_msg_str()


class HeartBeatInfoHandle:
    """心跳信息处理"""

    @staticmethod
    async def ws_send_keepalive(client: WebSocketClientProtocol):
        """
        websocket心跳消息上报协程
        :param client: websocket客户端
        :return: None
        """
        while True:
            try:
                ret = await HeartBeatInfoHandle._send_keepalive(client)
            except CancelledError:
                run_log.warning("send keep info task canceled")
                WsClientMgr().connected = False
                FdConnectStatus().trans_to_connecting()
                return
            except Exception as err:
                run_log.error("send alarm info failed, caught exception: %s", err)
                WsClientMgr().connected = False
                FdConnectStatus().trans_to_connecting()
                return

            if not ret:
                FdConnectStatus().trans_to_connecting()
                return

    @staticmethod
    async def _send_keepalive(client: WebSocketClientProtocol):
        keep_info = HeartBeatInfoHandle._get_keepalive()
        if client.closed:
            run_log.error("websocket client connection is closed, can not send msg now")
            WsClientMgr().connected = False
            return Result(result=False)

        try:
            await asyncio.wait_for(client.send(keep_info), timeout=WsClientMgr.SEND_TIMEOUT)
        except futures.TimeoutError:
            run_log.warning("Send keepalive info time out.")
        except CancelledError:
            # 遇到任务取消时，需要抛出外层以结束任务
            run_log.warning("Send keep info task canceled")
            raise
        except Exception as err:
            run_log.warning("Common exception, err type :%s, err: %s.", type(err), err)

        await WsClientMgr.sleep_for_a_while(15)
        return Result(result=True)

    @staticmethod
    def _get_keepalive():
        resp_msg_obj = FdMsgData.gen_ws_msg_obj(
            "ping", "node", "", False, group="resource", operation="keepalive", source="websocket"
        )
        return resp_msg_obj.to_ws_msg_str()
