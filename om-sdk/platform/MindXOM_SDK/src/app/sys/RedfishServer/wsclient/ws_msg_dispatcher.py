# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from asyncio import CancelledError

from websockets.exceptions import ConnectionClosedError
from websockets.legacy.client import WebSocketClientProtocol

from common.utils.result_base import Result

from common.log.logger import run_log
from common.token_bucket import TokenBucket
from fd_msg_process.config import Topic
from fd_msg_process.midware_proc import MidwareProc
from mef_msg_process.mef_proc import MefProc
from net_manager.constants import NetManagerConstants
from net_manager.manager.fd_cfg_manager import FdMsgData
from wsclient.connect_status import FdConnectStatus
from wsclient.fd_connect_check import FdConnectCheck
from wsclient.ws_client_mef import WsClientMef
from wsclient.ws_client_mgr import WsClientMgr


class WsMsgDispatcher:
    """Websocket消息分派"""

    token_bucket = TokenBucket()

    @staticmethod
    async def dispatch(client: WebSocketClientProtocol):
        """
        websocket消息分派处理协程
        :param client: websocket客户端对象
        :return: None
        """
        try:
            async for msg in client:
                if not WsMsgDispatcher.token_bucket.consume_token(1):
                    continue

                ret = WsMsgRouter.parse_msg(msg)
                if not ret:
                    run_log.error("parse msg failed")
                    continue

                msg_obj: FdMsgData = ret.data
                # 未到连接就绪状态，只允许处理FD下发的网管配置一机一密的消息
                connect_ready = FdConnectCheck.check_fd_connect_status_ready()
                if not connect_ready and not WsMsgRouter.is_white_list_when_not_ready(msg_obj):
                    run_log.warning("fd connect status is not ready, cannot dispatch msg")
                    continue

                ret = WsMsgRouter.accept(msg_obj, msg)
                if not ret:
                    run_log.error("accept msg failed: %s", ret.error)
        except CancelledError:
            run_log.warning("Websocket msg dispatch task cancelled.")
            WsClientMgr().connected = False
        except ConnectionClosedError as err:
            run_log.warning("Websocket msg dispatch task done because of websocket connect closed: %s.", err)
            WsClientMgr().connected = False
        except Exception as err:
            run_log.error("Websockets dispatch process failed: caught exception: %s.", err)
            WsClientMgr().connected = False

        FdConnectStatus().trans_to_connecting()


class WsMsgRouter:
    """Websocket消息处理路由"""
    _MSG_ROUTE_MAP = {
        "websocket/profile": {"deal_func_route": "espmanager/SysConfig", "up_resource": "websocket/profile"},
        "websocket/profile_effect": {
            "deal_func_route": "espmanager/SysConfigEffect",
            "up_resource": "websocket/profile_effect",
        },
        "websocket/tag": {"deal_func_route": "espmanager/SysAssetTag", "up_resource": "websocket/config_result"},
        "websocket/restart": {
            "deal_func_route": "espmanager/ComputerSystemReset",
            "up_resource": "websocket/restart_result",
        },
        "websocket/firmware_effective": {
            "deal_func_route": "espmanager/UpdateService/FirmwareEffective",
            "up_resource": "websocket/restart_result",
        },
        "websocket/info_collect": {
            "deal_func_route": "espmanager/InfoCollect",
            "up_resource": "websocket/info_collect_process",
        },
        "websocket/rearm": {"deal_func_route": Topic.SUB_RESET_ALARM, "up_resource": ""},
        "websocket/config_hostname": {
            "deal_func_route": "espmanager/Hostname",
            "up_resource": "websocket/config_result",
        },
        "websocket/netmanager": {"deal_func_route": "espmanager/netmanager", "up_resource": "websocket/config_result"},
        "websocket/passthrough/account_modify": {
            "deal_func_route": "espmanager/passthrough/account_modify",
            "up_resource": "websocket/config_result",
        },
        "websocket/config_dflc": {
            "deal_func_route": "espmanager/config_dflc",
            "up_resource": "websocket/config_dflc_result",
        },
        "websocket/install": {
            "deal_func_route": "",
            "up_resource": "websocket/upgrade_progress",
        },
        "websocket/cert_update": {
            "deal_func_route": "espmanager/cert_update",
            "up_resource": "websocket/config_result",
        },
        "websocket/crl_update": {
            "deal_func_route": "espmanager/crl_update",
            "up_resource": "websocket/config_result",
        },
        "websocket/cert_delete": {
            "deal_func_route": "espmanager/cert_delete",
            "up_resource": "websocket/config_result",
        },
        "websocket/cert_query": {
            "deal_func_route": "espmanager/cert_query",
            "up_resource": "websocket/cert_info",
        },
        "websocket/min_recovery": {
            "deal_func_route": "espmanager/handle_recover_mini_os",
            "up_resource": "websocket/config_result",
        },
    }

    @staticmethod
    def accept(msg: FdMsgData, ori_msg) -> Result:
        """
        将白名单的消息对象实例加入消息处理队列
        :param msg: FdMsgData消息对象
        :param ori_msg: FD下发的原始消息内容
        :return: Result
        """
        down_resource = msg.route.resource
        # FD下发的消息source需要在白名单中，防伪造
        if msg.route.source not in NetManagerConstants.SOURCE_WHITELIST:
            run_log.error("Received source is not supported: %s.", msg.route.source)
            return Result(result=False)

        if down_resource not in WsMsgRouter._MSG_ROUTE_MAP:
            # 当前MEF通道建立，需要转发给MEF处理
            if WsClientMef.is_manage_mef():
                run_log.info("Dispatch FD msg to MEF.")
                MefProc.add_task_from_fd(ori_msg)
                return Result(True)

            run_log.error("Received resource is not supported: %s.", down_resource)
            return Result(result=False, err_msg="received resource is not supported")

        msg.deal_func_name = WsMsgRouter._MSG_ROUTE_MAP.get(down_resource).get("deal_func_route")
        msg.up_resource = WsMsgRouter._MSG_ROUTE_MAP.get(down_resource).get("up_resource")
        MidwareProc.add_task(msg)
        return Result(result=True)

    @staticmethod
    def is_white_list_when_not_ready(msg_obj: FdMsgData) -> bool:
        """
        判断是否是配置一机一密的消息
        :param msg_obj: 解析的消息实例
        :return: Result
        """
        return msg_obj.route.resource == "websocket/netmanager"

    @staticmethod
    def parse_msg(msg) -> Result:
        """
        解析消息
        :param msg: 接收的websocket消息
        :return: Result
        """
        try:
            msg_obj: FdMsgData = FdMsgData.from_str(msg)
        except Exception as err:
            run_log.error("parse msg failed: %s", err)
            return Result(result=False)

        return Result(result=True, data=msg_obj)
