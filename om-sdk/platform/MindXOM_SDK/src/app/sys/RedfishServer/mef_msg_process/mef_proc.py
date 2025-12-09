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
import json
import threading
import time
import uuid
from typing import Optional

from common.log.logger import run_log
from common.utils.timer import RepeatingTimer
from mef_msg_process.mef_msg import MefMsgData
from mef_msg_process.msg_que import msg_que_from_mef
from mef_msg_process.msg_que import msg_que_to_mef
from mef_msg_process.msg_que import alarm_que_from_mef
from net_manager.manager.fd_cfg_manager import FdCfgManager
from net_manager.schemas import HeaderData
from net_manager.schemas import RouteData
from wsclient.ws_client_mef import FailedReason
from wsclient.ws_client_mef import WsClientMef
from wsclient.ws_config import WsConfig


class MefProc:
    """MEF消息处理类"""
    RESOURCE_IMPORT_CERT_INFO = "/edge/system/image-cert-info"
    RESOURCE_ALARM_INFO = "/edge/system/all-alarm"
    # 进程启动和网管切换会开启线程，使用锁控制只有一个线程生效
    MEF_START_LOCK: threading.Lock = threading.Lock()
    mef_connect_timer: Optional[RepeatingTimer] = None

    @staticmethod
    def add_task_from_fd(msg: str):
        """将FD的消息放入消息队列"""
        task_que = msg_que_to_mef

        if task_que.full():
            run_log.warning("The msg_que_to_mef is full, will clear all.")
            task_que.queue.clear()

        try:
            task_que.put(msg, False)
        except Exception as abn:
            run_log.error("put message to queue failed: %s", abn)

    @staticmethod
    def add_task_from_mef(msg: str):
        """将MEF的消息过滤分类后放入消息队列"""
        resource = ""
        try:
            source = json.loads(msg)
            resource = source.get("route").get("resource")
        except Exception as err:
            run_log.warning("The msg content is invalid. reason: %s", err)

        target_que = alarm_que_from_mef if MefProc.RESOURCE_ALARM_INFO == resource else msg_que_from_mef

        if target_que.full():
            run_log.warning("The msg que from mef is full, will clear all.")
            target_que.queue.clear()

        try:
            target_que.put(msg, False)
        except Exception as abn:
            run_log.error("put message to queue failed: %s", abn)

    @staticmethod
    def start_mef_connect_timer():
        """启动MEF对接任务"""
        if MefProc.MEF_START_LOCK.locked():
            run_log.warning("Mef connect start task already exist, no need again, will return.")
            return

        with MefProc.MEF_START_LOCK:
            if not WsClientMef.is_manage_mef():
                run_log.info("No need start mef connect task now.")
                return

            if not WsClientMef.exchange_ca_with_mef():
                WsClientMef().failed_reason = FailedReason.ERR_EXCHANGE_CERT_FAILED

            if MefProc.mef_connect_timer:
                run_log.warning("Mef connect timer is not empty, will discard it, ident: %s.",
                                MefProc.mef_connect_timer.ident)

            WsClientMef().cancelled = False
            MefProc.mef_connect_timer = RepeatingTimer(None, WsClientMef().connect)
            MefProc.mef_connect_timer.start()
            run_log.info("Start connect mef task success, ident: %s.", MefProc.mef_connect_timer.ident)

    @staticmethod
    def stop_mef_connect_timer():
        """关闭mef对接"""
        if MefProc.MEF_START_LOCK.locked():
            run_log.warning("Stop mef connect task already exist, no need again, will return.")
            return

        with MefProc.MEF_START_LOCK:
            if not WsClientMef.is_manage_mef():
                run_log.info("No need start mef connect task now.")
                return

            if not MefProc.mef_connect_timer or not MefProc.mef_connect_timer.is_alive():
                MefProc.mef_connect_timer = None
                run_log.info("Mef connect time is empty or not alive, no need stop.")
                return

            WsClientMef().stop_current_loops()

            WsClientMef().cancelled = True
            MefProc.mef_connect_timer.cancel()

            for _ in range(WsConfig.WAIT_MAX_CNT):
                if not MefProc.mef_connect_timer.is_alive():
                    run_log.info("Stop mef connect timer success, ident: %s.", MefProc.mef_connect_timer.ident)
                    MefProc.mef_connect_timer = None
                    return
                time.sleep(1)

            run_log.warning("Stop mef connect timer timeout, ident: %s.", MefProc.mef_connect_timer.ident)

    @staticmethod
    def send_fd_info_to_mef():
        """向MEF推送FD配置信息"""
        msg = MefProc.gen_fd_info()
        MefProc.add_task_from_fd(msg)

    @staticmethod
    def gen_fd_info():
        header = HeaderData(msg_id=str(uuid.uuid4()), parent_msg_id="", sync=False, timestamp=round(time.time() * 1000))
        route = RouteData(group="resource", operation="update", resource=MefProc.RESOURCE_IMPORT_CERT_INFO,
                          source="device-om")
        content = FdCfgManager.get_fd_info_to_mef()
        return MefMsgData(header, route, content=content).to_ws_msg_str()

    @staticmethod
    def gen_query_alarm_info():
        header = HeaderData(msg_id=str(uuid.uuid4()), parent_msg_id="", sync=False, timestamp=round(time.time() * 1000))
        route = RouteData(group="hardware", operation="query", resource=MefProc.RESOURCE_ALARM_INFO,
                          source="device-om")
        return MefMsgData(header, route, content={}).to_ws_msg_str()

