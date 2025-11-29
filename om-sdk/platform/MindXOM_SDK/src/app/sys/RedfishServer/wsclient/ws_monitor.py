# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

import asyncio
import json
import threading
import time
from asyncio import CancelledError
from typing import Optional

from websockets.legacy.client import connect

from common.log.logger import run_log
from common.utils.result_base import Result
from common.utils.singleton import Singleton
from common.utils.timer import DynamicRepeatingTimer
from common.utils.timer import RepeatingTimer
from fd_msg_process.config import SysInfoTaskStatus
from net_manager.exception import NetManagerException
from net_manager.manager.fd_cfg_manager import FdCfgManager
from net_manager.manager.fd_cfg_manager import FdConfigData
from net_manager.manager.net_cfg_manager import NetCfgManager
from wsclient.connect_status import FdConnectStatus
from wsclient.fd_connect_check import FdConnectCheck
from wsclient.fd_connect_check import WsClientMgr
from wsclient.ws_client_mef import WsClientMef
from wsclient.ws_config import WsConfig
from wsclient.ws_info_reporter import HeartBeatInfoHandle, SystemInfoHandler
from wsclient.ws_msg_dispatcher import WsMsgDispatcher
from wsclient.ws_util import WsClientProtocol
from wsclient.ws_util import WsInvalidStatusCode


class WsMonitor(Singleton):
    """Websocket客户端对接任务监控"""
    MONITOR_LOCK = threading.Lock()
    # FD固定的初始对接账号
    FD_INIT_ACCOUNT = "EdgeAccount"
    # websocket通道重连等待间隔时间初始值1分钟
    _CONNECT_RETRY_TIME = 60
    _NAME = "WsMonitorWithFD"

    def __init__(self) -> None:
        self.monitor_timer: Optional[DynamicRepeatingTimer] = None

    @staticmethod
    def wait_sys_info_proc_done():
        """针对重启场景，等待系统静态信息收集任务完成，才能上报系统信息，否则信息不全, 最多等待1分钟，实测40秒可以完成"""
        run_log.info("Wait for sys info proc task done once.")
        for _ in range(60):
            if SysInfoTaskStatus.get():
                run_log.info("Sys info proc task done once after started.")
                return

            time.sleep(1)
        run_log.warning("Wait for sys info proc task time out.")

    @staticmethod
    def start_fd_connect_monitor():
        """连接FD的定时任务，监控是否需要重新启动连接网管协程"""
        if WsMonitor.MONITOR_LOCK.locked():
            run_log.warning("The websocket monitor timer task is busy.")
            return

        with WsMonitor.MONITOR_LOCK:
            if WsMonitor().check_monitor_timer_valid():
                run_log.info("The websocket monitor timer task is already active, no need start.")
                return

            if not SysInfoTaskStatus.get():
                WsMonitor.wait_sys_info_proc_done()

            WsMonitor().monitor_timer = DynamicRepeatingTimer(WsMonitor._NAME, WsMonitor._CONNECT_RETRY_TIME,
                                                              WsMonitor._start_monitor)
            WsMonitor().monitor_timer.start()
            for _ in range(WsConfig.WAIT_MAX_CNT):
                if WsMonitor().monitor_timer.is_alive():
                    run_log.info("Start websocket monitor timer task success, ident: %d.",
                                 WsMonitor().monitor_timer.ident)
                    return
                time.sleep(1)

            run_log.warning("Start websocket monitor timer task timeout, ident: %d.", WsMonitor().monitor_timer.ident)

    @staticmethod
    def stop_fd_connect_monitor():
        """取消连接FD的定时任务"""
        if WsMonitor.MONITOR_LOCK.locked():
            run_log.warning("The websocket monitor timer task is busy.")
            return

        with WsMonitor.MONITOR_LOCK:
            if not WsMonitor().monitor_timer or not WsMonitor().monitor_timer.is_alive():
                run_log.info("The websocket monitor timer task is already inactive, no need stop.")
                return

            WsMonitor().monitor_timer.cancel()
            for _ in range(WsConfig.WAIT_MAX_CNT):
                if not WsMonitor().monitor_timer.is_alive():
                    run_log.info("Stop websocket monitor timer task success, ident: %d.",
                                 WsMonitor().monitor_timer.ident)
                    WsMonitor().monitor_timer = None
                    return
                time.sleep(1)

            run_log.warning("Cancel websocket monitor timer task timeout, ident: %d.", WsMonitor().monitor_timer.ident)

    @staticmethod
    def check_spare_node_error(err: Exception, node_id):
        if not isinstance(err, WsInvalidStatusCode):
            return

        error_mes = ""
        try:
            res = json.loads(err.body.decode("utf-8"))
        except Exception as abn:
            run_log.error("json loads response data failed, find exception: %s", abn)
            return

        if "error" in res.keys():
            error_mes = res.get("error").get("@Message.ExtendedInfo")[0].get("MessageId")

        if error_mes == WsClientMgr.ERR_INVALID_SPARE_NODE_ID:
            WsClientMgr().spare_node_id = res.get("error").get("@Message.ExtendedInfo")[0].get("MessageArgs")[0]
            run_log.info("Current node is spare node, current node id is not equal to old device node id: %s.",
                         WsClientMgr().spare_node_id)
            # 备件替换需要更换node id
            if WsClientMgr().spare_node_id:
                try:
                    NetCfgManager().update_net_cfg_info({"node_id": WsClientMgr().spare_node_id})
                except Exception as err:
                    run_log.error("Update current spare node id failed: %s.", err)
                    return

                run_log.info("Update spare device node id success: %s -> %s.", node_id, WsClientMgr().spare_node_id)
                WsClientMgr().spare_node_id = ""

    @staticmethod
    def init_data_after_connected(client, server_ip: str, server_name: str):
        WsClientMgr().client_obj = client
        WsClientMgr().connected = True
        WsClientMgr().fd_ip = server_ip
        WsClientMgr().event_loop = asyncio.new_event_loop()
        WsClientMgr().fd_server_name = server_name

    @staticmethod
    async def start_all_task(client):
        await asyncio.gather(
            WsMsgDispatcher.dispatch(client),
            HeartBeatInfoHandle.ws_send_keepalive(client),
            SystemInfoHandler.ws_send_sys_info(client),
            SystemInfoHandler.ws_send_sys_status(client),
            SystemInfoHandler.ws_send_alarm(client),
            SystemInfoHandler.ws_send_event(client),
            SystemInfoHandler.route_msg_to_fd(client),
        )

    @staticmethod
    def stop_all_connect_threads():
        """停止对接线程任务"""
        from mef_msg_process.mef_proc import MefProc
        WsMonitor.stop_fd_connect_monitor()
        WsClientMgr().stop_current_fd_loops()
        WsClientMgr().stop_current_fd_connect_timer()
        MefProc.stop_mef_connect_timer()

    @staticmethod
    def _start_monitor() -> bool:
        """返回值表示是否需要屏蔽循环等待时间日志"""
        client_mgr = WsClientMgr()
        try:
            reconnect = client_mgr.need_reconnect()
        except Exception as err:
            run_log.error("Get reconnect status failed, %s", err)
            return True

        if not reconnect:
            # 不需要重连的场景，屏蔽循环线程的等待日志
            return False

        FdConnectStatus().trans_to_connecting()

        if client_mgr.conn_timer:
            run_log.info("Try to stop current loops before connect timer cancel.")
            client_mgr.stop_current_fd_loops()
            client_mgr.stop_current_fd_connect_timer()
            run_log.info("Old FD %s connect timer cancelled.", client_mgr.fd_ip or "initial")

        client_mgr.conn_timer = RepeatingTimer(None, function=WsMonitor._start_aio_loop)
        client_mgr.conn_timer.start()
        return True

    @staticmethod
    def _start_aio_loop():
        loop = asyncio.new_event_loop()
        WsClientMgr().connect_loop = loop

        try:
            ret = loop.run_until_complete(WsMonitor._start_ws_connect())
        except Exception as err:
            run_log.warning("Start websocket connect task failed, caught exception: %s", err)
            return Result(False)

        if not ret:
            run_log.warning("Websocket connect task completed, result: %s.", ret.error)
            return ret

        return ret

    @staticmethod
    async def _start_ws_connect():
        ret = FdConnectCheck.get_fd_config_data()
        if not ret:
            run_log.error("Get fd config data failed: %s.", ret.error)
            return Result(False, err_msg=f"Get fd config data failed: {ret.error}")

        config: FdConfigData = ret.data
        # 1.测试连接
        ret = FdConnectCheck.connect_test(config)

        if not ret:
            run_log.error("Test connecting to FD %s failed: %s.", config.server_ip, ret.error)
            if WsClientMgr().check_account_info_invalid() or WsClientMgr().check_cert_is_invalid() or \
                    WsClientMgr().check_ip_locked():
                FdConnectStatus().trans_to_err_configured()
                WsMonitor.stop_all_connect_threads()
                err_msg = "invalid account info or cert info or ip locked."
                run_log.error("%s", err_msg)
            return Result(False, err_msg=f"test connecting to FD failed: {ret.error}")

        run_log.info("Test connecting to FD %s successfully.", config.server_ip)

        # 上次连接任务更新的状态
        pre_status = config.status
        if pre_status == "":
            # 如果是首次纳管，初始账号密码校验通过之后，更新数据库状态为connecting，避免连接任务在周期内被重启
            WsMonitor._update_status("connecting")

        # 2.发起websocket连接
        headers = config.gen_extra_headers()
        try:
            ssl_context = config.gen_client_ssl_context(set_state=True)
        except NetManagerException as err:
            run_log.error("Invalid ssl context, catch %s", err)
            return Result(False)

        try:
            async with connect(uri=config.wss_uri, ssl=ssl_context, extra_headers=headers,
                               klass=WsClientProtocol) as client:
                run_log.info("Websocket connect to FD %s successfully.", config.server_ip)

                # 更新/etc/hosts
                old_server_name = WsClientMgr().fd_server_name
                if not FdCfgManager.update_etc_hosts(config.server_ip, old_server_name, config.server_name):
                    run_log.error("Update FD info to etc hosts failed.")
                    return Result(False)

                # 更新WsClientMgr实例属性
                WsMonitor.init_data_after_connected(client, config.server_ip, config.server_name)

                if pre_status == "connected":
                    # 使用FD下发新的一机一密进行连接，连接成功之后，状态从connected改为ready就绪状态
                    WsMonitor._update_status("ready")
                    # 纳管后查询一次FD证书即将过期告警
                    FdCfgManager().check_cert_status()

                if config.cloud_user != WsMonitor.FD_INIT_ACCOUNT:
                    FdConnectStatus().trans_to_connected()

                if WsClientMef().ready_for_send_msg():
                    WsClientMef.send_fd_info_to_mef()

                await WsMonitor.start_all_task(client)
        except CancelledError:
            run_log.warning("Websocket connect task cancelled.")
            WsClientMgr().connected = False
            FdConnectStatus().trans_to_connecting()
            return Result(False, err_msg="websocket connect task cancelled")
        except WsInvalidStatusCode as err:
            run_log.warning(f"invalid status code: {err.status_code}, {err}")
            WsMonitor.check_spare_node_error(err, config.node_id)
            FdConnectStatus().trans_to_connecting()
            return Result(False, err_msg="invalid status code, connect task cancelled")
        except Exception as err:
            run_log.error("Websocket client connect failed, caught exception: %s.", err)
            WsClientMgr().connected = False
            FdConnectStatus().trans_to_connecting()
            return Result(False, err_msg="websocket client connect failed, caught exception")

        return Result(True)

    @staticmethod
    def _update_status(status):
        try:
            NetCfgManager().update_net_cfg_info({"status": status})
        except Exception as err:
            run_log.error("Update status failed: %s.", err)

        run_log.info("Update connect status successfully: %s.", status)

    def check_monitor_timer_valid(self):
        return self.monitor_timer and self.monitor_timer.is_alive()
