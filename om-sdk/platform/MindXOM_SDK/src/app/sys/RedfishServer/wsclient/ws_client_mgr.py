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
import threading
import time
from asyncio import AbstractEventLoop
from asyncio import futures
from typing import Optional

from common.utils.result_base import Result
from websockets.legacy.client import WebSocketClientProtocol

from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.singleton import Singleton
from common.utils.timer import RepeatingTimer
from net_manager.constants import NetManagerConstants
from net_manager.manager.fd_cfg_manager import FdCfgManager
from net_manager.manager.fd_cfg_manager import FdConfigData
from net_manager.manager.fd_cfg_manager import FdMsgData
from wsclient.connect_status import FdConnectStatus
from wsclient.ws_config import WsConfig


class WsClientMgr(Singleton):
    """Websocket客户端管理单例对象"""
    # 对接FD IP被锁定
    ERR_IP_LOCKED = "FusionDirector.1.0.IAMRequestIPLocked"
    # 对接FD的账号密码错误
    ERR_INVALID_ACCOUNT = ("Base.1.0.InsufficientPrivilege", "FusionDirector.1.0.AuthenticationFailure", )
    ERR_INVALID_SPARE_NODE_ID = "FusionDirector.1.0.SpareNodeIDInCorrect"
    ERR_INVALID_FD_CERT = "invalid fd cert"
    ERR_INVALID_FD_IP_PORT = "invalid fd ip or port"

    # websocket消息发送任务超时时间，实测1秒也可以正常使用
    SEND_TIMEOUT = 30
    # 停止所有协程任务的最大尝试次数30次
    _MAX_RETRY_CNT = 30
    _STOP_LOCK = threading.Lock()
    _CONNECT_TEST_RESULT_LOCK = threading.Lock()

    def __init__(self) -> None:
        super().__init__()
        self.conn_timer: Optional[RepeatingTimer] = None
        self.connected: bool = False
        self.client_obj: Optional[WebSocketClientProtocol] = None
        self.connect_loop: Optional[AbstractEventLoop] = None
        self.event_loop: Optional[AbstractEventLoop] = None
        self.fd_ip: str = ""
        self.fd_server_name: str = ""
        self.connect_test_result = ""
        self.spare_node_id = ""

    @staticmethod
    async def sleep_for_a_while(seconds):
        """
        因为正在sleep的任务无法取消，为了更快响应取消pending任务，需要用以下写法
        :param seconds: 等待任务的秒数
        :return: None
        """
        for _ in range(seconds):
            await asyncio.sleep(1)

    @staticmethod
    def stop_event_loop_now(loop: AbstractEventLoop) -> Result:
        """停止事件循环loop中的所有协程任务停止并关闭"""
        if not loop:
            run_log.info("Current loop is None.")
            return Result(True)

        if loop.is_closed():
            run_log.info("Current loop is already closed.")
            return Result(True)

        asyncio.set_event_loop(loop)
        WsClientMgr._cancel_all_tasks(loop)
        unfinished_tasks = asyncio.all_tasks(loop)
        cur_cnt = 0
        while unfinished_tasks:
            unfinished_tasks = asyncio.all_tasks(loop)
            time.sleep(1)
            cur_cnt += 1
            if cur_cnt >= WsClientMgr._MAX_RETRY_CNT:
                run_log.warning("Try count of stopping all tasks in current event loop is up to limit.")
                break

        run_log.info("All tasks in current loop are cancelled.")

        loop.stop()
        run_log.info("Stop current loop done.")

        loop.close()
        run_log.info("Close current loop done.")

        WsClientMgr._wait_for_loop_closed(loop)
        run_log.info("Wait for current loop to be closed done.")

        return Result(result=loop.is_closed())

    @staticmethod
    def _cancel_all_tasks(event_loop):
        tasks = asyncio.all_tasks(loop=event_loop)
        group_task = asyncio.gather(*tasks, return_exceptions=True)
        group_task.cancel()

    @staticmethod
    async def _send_msg_by_ws(msg: str, client: WebSocketClientProtocol) -> Result:
        if not client or client.closed:
            run_log.error("Websocket client is empty or closed, can not send msg now.")
            return Result(result=False, err_msg="websocket client is empty or closed, can not send msg now")

        try:
            await asyncio.wait_for(client.send(msg), timeout=WsClientMgr.SEND_TIMEOUT)
        except futures.TimeoutError:
            run_log.warning("Websocket client send msg time out.")
            return Result(result=False, err_msg="websocket client send msg time out")
        except futures.CancelledError:
            run_log.warning("Websocket client send msg task cancelled.")
            return Result(result=False, err_msg="websocket client send msg task cancelled")
        except Exception as err:
            run_log.error("Send msg failed, caught exception: %s, type is %s.", err, type(err))
            return Result(result=False, err_msg=f"send response failed, caught exception: {err}, type is {type(err)}")

        run_log.info("Websocket client send msg success.")
        return Result(True)

    @staticmethod
    def _wait_for_loop_closed(loop):
        cur_cnt = 0
        while not loop.is_closed:
            time.sleep(1)
            cur_cnt += 1
            if cur_cnt > WsClientMgr._MAX_RETRY_CNT:
                run_log.warning("Try count of waiting for loop closing is up to limit.")
                break

    def clear_connect_result(self):
        """清空连接测试结果"""
        with self._CONNECT_TEST_RESULT_LOCK:
            self.connect_test_result = ""

    def check_account_info_invalid(self):
        """连接测试结果，账号密码不对"""
        with self._CONNECT_TEST_RESULT_LOCK:
            return self.connect_test_result in WsClientMgr.ERR_INVALID_ACCOUNT

    def check_ip_locked(self):
        """连接测试结果，账号密码不对"""
        with self._CONNECT_TEST_RESULT_LOCK:
            return self.connect_test_result == WsClientMgr.ERR_IP_LOCKED

    def check_cert_is_invalid(self):
        """是否证书、吊销列表有问题"""
        with self._CONNECT_TEST_RESULT_LOCK:
            return self.connect_test_result == WsClientMgr.ERR_INVALID_FD_CERT

    def check_ip_port_is_invalid(self):
        """是否ip、port有问题"""
        with self._CONNECT_TEST_RESULT_LOCK:
            return self.connect_test_result == WsClientMgr.ERR_INVALID_FD_IP_PORT

    def set_connect_result(self, res):
        with self._CONNECT_TEST_RESULT_LOCK:
            self.connect_test_result = res

    def need_reconnect(self):
        """
        判断是否需要重新启动连接动作
        :return:
        """
        ret = FdCfgManager().get_ws_config()
        if not ret:
            run_log.error("get config failed: %s, need reconnect", ret.error)
            return True

        cfg: FdConfigData = ret.data
        if cfg.net_mgmt_type != NetManagerConstants.FUSION_DIRECTOR:
            return False

        if not self.conn_timer:
            return True

        if not hasattr(self.conn_timer, 'is_alive'):
            return True

        if not self.conn_timer.is_alive():
            return True

        if self.ready_for_send_msg():
            # 当前已经纳管到FD1且就绪，重新触发配置FD2，如果FD2的配置不对，不影响当前的纳管状态
            return False

        if self.check_account_info_invalid():
            # 账号密码错误不重连
            FdConnectStatus().trans_to_err_configured()
            from wsclient.ws_monitor import WsMonitor
            WsMonitor.stop_all_connect_threads()
            return False

        if self.check_ip_locked():
            # IP被锁不重连
            FdConnectStatus().trans_to_err_configured()
            from wsclient.ws_monitor import WsMonitor
            WsMonitor.stop_all_connect_threads()
            return False

        if not self.connected:
            return True

        if cfg.status in ("ready", "connecting"):
            return False

        fd_ip_in_hosts = AppCommonMethod.get_fd_ip_from_etc_hosts(cfg.server_name)
        if self.fd_ip != fd_ip_in_hosts:
            return True

        return False

    def ready_for_send_msg(self) -> bool:
        """
        客户端是否可以与服务端发送消息
        """
        return self.connected and self.client_obj and self.client_obj.open and \
            self.event_loop and not self.event_loop.is_closed()

    def send_msg(self, msg: FdMsgData) -> Result:
        """
        通过websocket client发送消息
        :param msg: 上行消息
        :return:
        """
        if not self.conn_timer or not self.conn_timer.is_alive():
            run_log.error("connect task inst not initialized")
            return Result(result=False, err_msg="connect task inst not initialized")

        if not self.connected or not self.client_obj:
            run_log.error("status is not connected or client obj is invalid")
            return Result(result=False, err_msg="status is not connected or client obj is invalid")

        if not isinstance(msg, FdMsgData):
            run_log.error("invalid msg format")
            return Result(result=False, err_msg="invalid msg format")

        ret = self.send_msg_by_ws(msg.to_ws_msg_str())
        if not ret:
            run_log.error("send msg by ws failed: %s", ret.error)
            return ret

        return Result(True)

    def send_msg_by_ws(self, msg: str) -> Result:
        if not self.event_loop or self.event_loop.is_closed():
            run_log.error("Current event loop is None or closed, can not send msg.")
            return Result(result=False, err_msg="current event loop is closed, can not send msg now")

        if not self.client_obj or not self.client_obj.open:
            run_log.error("Current ws client obj is not open state, can not send msg now.")
            return Result(result=False, err_msg="current ws client obj is not open state, can not send msg now")

        asyncio.set_event_loop(self.event_loop)

        wait_time, max_time = 0, 5
        while self.event_loop.is_running():
            time.sleep(wait_time)
            if wait_time > max_time:
                run_log.warning("Event loop is busy, wait up to max time.")
                break
            wait_time += 1

        try:
            ret = self.event_loop.run_until_complete(WsClientMgr._send_msg_by_ws(msg, self.client_obj))
        except Exception as err:
            run_log.error("send msg failed, caught exception: %s", err)
            return Result(result=False, err_msg=f"send msg failed, caught exception: {err}")

        if not ret:
            run_log.error("send msg failed: %s", ret.error)
            return ret

        return Result(True)

    def stop_current_fd_loops(self):
        """停止当前与FD之间的连接事件循环与发送消息事件循环"""
        if self._STOP_LOCK.locked():
            run_log.warning("Stop current loops is busy.")
            return

        with self._STOP_LOCK:
            if not self._need_stop_loops():
                run_log.info("No need stop loops, all loops is closed.")
                return

            for loop_name, loop in ("connect", self.connect_loop), ("event", self.event_loop):
                try:
                    ret = self.stop_event_loop_now(loop)
                except Exception as err:
                    run_log.warning("Stop current %s loop failed, caught exception: %s.", loop_name, err)
                    continue

                if not ret:
                    run_log.warning("Stop current %s loop failed, loop is not closed.", loop_name)
                    continue

                run_log.info("Stop current %s loop successfully.", loop_name)

            if not self._need_stop_loops() and self.fd_ip:
                run_log.info("Disconnect to FD %s successfully.", self.fd_ip)

            self.connected = False
            self.fd_ip = ""
            self.clear_connect_result()
            run_log.info("Stop current loops done.")

    def stop_current_fd_connect_timer(self):
        """停止当前与FD之间的connect线程"""
        if self._STOP_LOCK.locked():
            run_log.warning("Stop current connect timer is busy.")
            return

        with self._STOP_LOCK:
            if not self.conn_timer or not self.conn_timer.is_alive():
                self.conn_timer = None
                run_log.info("Current connect timer is empty or not alive, no need stop.")
                return

            self.conn_timer.cancel()
            # 等待线程取消成功
            for _ in range(WsConfig.WAIT_MAX_CNT):
                if not self.conn_timer.is_alive():
                    run_log.info("Stop connect timer success, ident: %s.", self.conn_timer.ident)
                    self.conn_timer = None
                    return
                time.sleep(1)

            run_log.warning("Stop connect timer timeout, ident: %s.", self.conn_timer.ident)

    def _need_stop_loops(self):
        return any(loop and not loop.is_closed() for loop in (self.connect_loop, self.event_loop))
