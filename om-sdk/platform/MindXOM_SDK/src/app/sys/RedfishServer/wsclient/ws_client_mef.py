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
import asyncio
import json
import os
import ssl
import threading
import time
from asyncio import AbstractEventLoop
from asyncio import CancelledError
from enum import Enum
from queue import Empty
from socket import create_connection
from typing import Optional

import websockets
from websockets.exceptions import ConnectionClosedError
from websockets.legacy.client import WebSocketClientProtocol

from common.constants.base_constants import MefOperate, MefNetStatus
from common.log.logger import run_log
from common.token_bucket import TokenBucket
from common.utils.ability_policy import OmAbility
from common.utils.ability_policy import is_allow
from common.utils.app_common_method import AppCommonMethod
from common.utils.singleton import Singleton
from common.utils.result_base import Result
from lib_restful_adapter import LibRESTfulAdapter
from mef_msg_process.mef_config_mgr import MefConfigData
from mef_msg_process.msg_que import msg_que_to_mef, alarm_que_from_mef
from net_manager.checkers.contents_checker import CertContentsChecker
from net_manager.manager.fd_cfg_manager import FdCfgManager
from wsclient.ws_client_mgr import WsClientMgr
from wsclient.ws_config import WsConfig


class FailedReason(Enum):
    """连接失败原因"""
    EMPTY = ""
    ERR_INVALID_SSL_CONTEXT = "invalid ssl context"
    ERR_CERT_VERIFIED_FAILED = "certification verify failed"
    ERR_EXCHANGE_CERT_FAILED = "exchange mef cert failed"


class WsClientMef(Singleton):
    """MEF对接的Websocket客户端类"""
    # 进程启动和网管切换会开启线程，使用锁控制只有一个线程生效
    MEF_CONNECT_LOCK: threading.Lock = threading.Lock()
    # websocket消息发送任务超时时间
    SEND_TIMEOUT = 30
    # MEF service
    MEF_SERVICE_PATH = "/usr/lib/systemd/system/mef-edge-main.service"
    MEF_SERVICE_NAME = "mef-edge-main"
    DOCKER_SERVICE_NAME = "docker"

    # 停止所有协程任务的最大尝试次数30次
    _MAX_RETRY_CNT = 30
    _STOP_LOCK = threading.Lock()
    token_bucket = TokenBucket()

    def __init__(self) -> None:
        super().__init__()
        self.client_obj: Optional[WebSocketClientProtocol] = None
        self.connect_loop: Optional[AbstractEventLoop] = None
        self.failed_reason: FailedReason = FailedReason.EMPTY
        self.cancelled: bool = False

    @staticmethod
    def is_manage_mef() -> bool:
        """
        MEF是否被OM管理：当能力项关闭和当MEF网管状态不是FD_OM时，OM不管理MEF
        """
        if not is_allow(OmAbility.MEF_CONFIG):
            run_log.warning("Ability %s is disable, Do not manage the MEF process.", OmAbility.MEF_CONFIG.value)
            return False

        ret_dict = LibRESTfulAdapter.lib_restful_interface("Mef", "GET", None)
        if not AppCommonMethod.check_status_is_ok(ret_dict):
            run_log.warning("Get MEF net status failed, Do not manage the MEF process.")
            return False

        if ret_dict["message"]["mef_net_status"] == MefNetStatus.FD_OM.value:
            return True

        run_log.warning("Get MEF net status is not [FD_OM], Do not manage the MEF process.")
        return False

    @staticmethod
    def check_mef_port_available() -> bool:
        """服务存在但是可能mef服务没有拉起来，需要检查端口是否可用"""
        ip_port = (MefConfigData.host, int(MefConfigData.port))

        try:
            with create_connection(address=ip_port, timeout=WsConfig.SOCKET_CONNECT_TIMEOUT):
                run_log.info("Ip or port of MEF available.")
                return True
        except ConnectionRefusedError as err:
            # ip不可达情况
            run_log.warning("Ip or port of MEF not available, err: %s, type is %s.", err, type(err))
        except Exception as err:
            run_log.error("Check ip or port of MEF failed, caught exception: %s, err type is %s.", err, type(err))

        return False

    @staticmethod
    def exchange_ca_with_mef() -> Result:
        # 交换根证书
        request_data = {
            "operate_type": MefOperate.EXCHANGE_CRT.value,
        }
        ret = LibRESTfulAdapter.lib_restful_interface("Mef", "POST", request_data, False)
        if not AppCommonMethod.check_status_is_ok(ret):
            run_log.error("Exchange root ca cert failed.")
            return Result(False, err_msg="exchange root ca cert failed")

        # 查询根证书
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Mef", "GET", None, False, True)
        if not AppCommonMethod.check_status_is_ok(ret_dict):
            run_log.error("Get mef root ca data failed.")
            return Result(False, err_msg="get mef root ca data failed")

        # 校验根证书内容合法性
        ca_data = ret_dict.get("message", {}).get("mef_ca_data")
        if not ca_data:
            run_log.error("MEF root ca data is empty.")
            return Result(False, err_msg="MEF root ca data is empty")

        ret = CertContentsChecker("cert").check_dict({"cert": ca_data})
        if not ret:
            run_log.error("invalid mef ca cert content: %s", ret.reason)
            return Result(False, err_msg="invalid cert content")

        # 校验根证书路径合法性
        root_crt_path = MefConfigData.root_crt_path
        try:
            if os.path.islink(root_crt_path):
                run_log.warning("%s path invalid: is link", root_crt_path)
                os.unlink(root_crt_path)
        except Exception as error:
            run_log.error("Unlink file %s failed, caught exception: %s", root_crt_path, error)
            return Result(False, err_msg="unlink file failed")

        # 保存根证书至文件
        try:
            with os.fdopen(os.open(root_crt_path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as file:
                file.write(ca_data)
        except Exception as error:
            run_log.error("save mef ca data to file failed, caught exception: %s", error)
            return Result(False, err_msg="save mef ca data to file failed")

        return Result(True)

    @staticmethod
    def send_fd_info_to_mef():
        from mef_msg_process.mef_proc import MefProc
        MefProc.send_fd_info_to_mef()

    @staticmethod
    def stop_mef_process() -> Result:
        """停止MEF进程"""
        if not WsClientMef.is_manage_mef():
            return Result(True)

        request_data = {
            "operate_type": MefOperate.STOP.value,
        }
        ret = LibRESTfulAdapter.lib_restful_interface("Mef", "POST", request_data)
        if not AppCommonMethod.check_status_is_ok(ret):
            run_log.error("Stop MEF process failed.")
            return Result(False)

        run_log.info("Stop MEF process success.")
        return Result(True)

    @staticmethod
    def restart_mef_process() -> Result:
        """重启MEF进程"""
        request_data = {
            "operate_type": MefOperate.RESTART.value,
        }
        ret = LibRESTfulAdapter.lib_restful_interface("Mef", "POST", request_data, False)
        if not AppCommonMethod.check_status_is_ok(ret):
            run_log.error("Restart MEF process failed.")
            return Result(False)

        return Result(True)

    @staticmethod
    async def receive_msg_from_mef(client: WebSocketClientProtocol):
        """
        接收来自MEF的消息
        :param client: websocket客户端对象
        :return: None
        """
        # 需要局部导入
        from mef_msg_process.mef_proc import MefProc
        try:
            async for msg in client:
                if not WsClientMef.token_bucket.consume_token(1):
                    continue

                if WsClientMef().cancelled:
                    run_log.warning("MEF connect timer task is cancelled.")
                    return

                MefProc.add_task_from_mef(msg)
        except CancelledError:
            run_log.warning("MEF msg dispatch task cancelled.")
        except ConnectionClosedError as err:
            run_log.warning("MEF msg dispatch task done because of connect closed: %s.", err)
        except Exception as err:
            run_log.error("MEF dispatch process failed: caught exception: %s, type is %s.", err, type(err))

    @staticmethod
    async def route_msg_to_mef(client: WebSocketClientProtocol):
        """消息分发"""
        while not WsClientMef().cancelled:
            if not WsClientMef().ready_for_send_msg():
                run_log.warning("The websocket connection with MEF is invalid, will stop the route msg to MEF task.")
                return

            try:
                msg = msg_que_to_mef.get_nowait()
            except Empty:
                await asyncio.sleep(0.5)
                continue

            try:
                await asyncio.wait_for(client.send(msg), timeout=WsClientMef.SEND_TIMEOUT)
            except asyncio.TimeoutError:
                run_log.warning("Websocket client send msg time out.")
                continue
            except asyncio.CancelledError:
                run_log.warning("Websocket client send msg task cancelled.")
                return Result(result=False, err_msg="websocket client send msg task cancelled")
            except Exception as err:
                run_log.error("Send msg failed, caught exception: %s, type is %s.", err, type(err))
                return Result(result=False,
                              err_msg=f"send response failed, caught exception: {err}, type is {type(err)}")

    def connect(self):
        """与MEF建立Websocket通道"""
        if self.MEF_CONNECT_LOCK.locked():
            run_log.warning("Mef connect task already exist, no need start again.")
            return

        with self.MEF_CONNECT_LOCK:
            while not self.cancelled:
                time.sleep(5)

                # 如果MEF未安装，则跳过
                if not os.path.exists(self.MEF_SERVICE_PATH):
                    continue

                if self._need_stop_mef():
                    run_log.info("Try stopping MEF...")
                    WsClientMef.stop_mef_process()
                    continue

                if not WsClientMgr().ready_for_send_msg() or not FdCfgManager.check_fd_mode_and_status_ready():
                    continue

                if self.ready_for_send_msg():
                    continue

                # 如果MEF端口未起来，则需要拉起MEF进程
                if not self.check_mef_port_available() and not WsClientMef.restart_mef_process():
                    continue

                if self.failed_reason.value and not WsClientMef.exchange_ca_with_mef():
                    continue

                run_log.info("Try connecting to MEF...")
                self.connect_loop = asyncio.new_event_loop()
                try:
                    self.connect_loop.run_until_complete(self._connect())
                except Exception as err:
                    run_log.warning("Start MEF websocket connect task failed, caught exception: %s", err)

                self.stop_current_loops()

            run_log.info("The MEF connect timer has been cancelled.")

    def ready_for_send_msg(self) -> bool:
        """
        客户端是否可以与服务端发送消息
        """
        return self.client_obj and self.client_obj.open and self.connect_loop and not self.connect_loop.is_closed()

    def stop_current_loops(self):
        """停止当前连接事件循环与发送消息事件循环"""
        if self._STOP_LOCK.locked():
            run_log.warning("Stop current loops is busy.")
            return

        with self._STOP_LOCK:
            if not self._need_stop_loops():
                run_log.info("No need stop loops, all loops is closed.")
                return

            loop_name = "connect"
            try:
                ret = WsClientMgr.stop_event_loop_now(self.connect_loop)
            except Exception as err:
                run_log.warning("Stop current MEF %s loop failed, caught exception: %s.", loop_name, err)
                return

            if not ret:
                run_log.warning("Stop current MEF %s loop failed, loop is not closed.", loop_name)
            else:
                run_log.info("Stop current MEF %s loop successfully.", loop_name)

            if not self.ready_for_send_msg():
                run_log.info("Disconnect to MEF successfully.")

            run_log.info("Stop current MEF loops done.")

    def get_mef_alarm_info(self):
        """
        获取MEF告警信息
        """
        if not self.ready_for_send_msg():
            return []

        from mef_msg_process.mef_proc import MefProc
        query_alarm_msg = MefProc.gen_query_alarm_info()
        MefProc.add_task_from_fd(query_alarm_msg)

        # 从告警队列里获取mef返回的告警信息
        try:
            alarm_info = alarm_que_from_mef.get(timeout=10)
        except Empty:
            run_log.error("Get MEF alarm info failed, timeout.")
            return []

        if not alarm_info:
            return []

        try:
            content = json.loads(alarm_info).get("content")
            return json.loads(content).get("alarm") if content else []
        except Exception:
            run_log.error("MEF alarm info is invalid.")
            return []

    async def _connect(self):
        config = MefConfigData()
        ssl_context_ret = config.gen_client_ssl_context()
        if not ssl_context_ret:
            run_log.error("Invalid ssl context: %s.", ssl_context_ret.error)
            self.failed_reason: FailedReason = FailedReason.ERR_INVALID_SSL_CONTEXT
            return

        ssl_context = ssl_context_ret.data
        try:
            async with websockets.connect(uri=config.wss_uri, ssl=ssl_context, ping_interval=10,
                                          ping_timeout=5) as client:
                run_log.info("Websocket connect to MEF successfully.")
                self.failed_reason: FailedReason = FailedReason.EMPTY
                self.client_obj = client

                WsClientMef.send_fd_info_to_mef()

                await asyncio.gather(
                    WsClientMef.receive_msg_from_mef(client),
                    WsClientMef.route_msg_to_mef(client),
                )
        except CancelledError:
            run_log.warning("MEF websocket connect task cancelled.")
        except Exception as err:
            if isinstance(err, ssl.SSLCertVerificationError):
                self.failed_reason: FailedReason = FailedReason.ERR_CERT_VERIFIED_FAILED
            run_log.error("MEF websocket client connect failed, caught exception: %s, err type is %s.", err, type(err))

    def _need_stop_loops(self):
        return self.connect_loop and not self.connect_loop.is_closed()

    def _need_stop_mef(self):
        """Web模式, docker与MEF服务启动, 需要停止MEF进程，减少资源占用"""
        is_web_mode = FdCfgManager.check_manager_type_is_web()
        is_mef_active = AppCommonMethod.check_service_status_is_active(self.MEF_SERVICE_NAME)
        is_docker_active = AppCommonMethod.check_service_status_is_active(self.DOCKER_SERVICE_NAME)
        return is_web_mode and is_mef_active and is_docker_active
