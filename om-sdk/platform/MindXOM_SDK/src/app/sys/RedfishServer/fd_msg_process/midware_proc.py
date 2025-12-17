#!/usr/bin/env python
# -*- coding:utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import datetime
import queue
import threading
import time

from common.file_utils import FileCheck
from common.log.logger import operate_log
from common.log.logger import run_log
from fd_msg_process.config import Topic
from fd_msg_process.fd_common_methods import publish_ws_msg
from fd_msg_process.fd_configs import MSG_HANDLING_MAPPING
from fd_msg_process.midware_route import MidwareRoute
from fd_msg_process.om_msg_queue import fd_message_que
from net_manager.manager.fd_cfg_manager import FdCfgManager
from net_manager.manager.fd_cfg_manager import FdMsgData
from wsclient.ws_client_mgr import WsClientMgr

RESTART_FLAG = "/run/restart_flag"
RESTARTING_FLAG = "/run/restarting_flag"


class MidwareProc(MidwareRoute):
    REPORT_ALARM_LOCK = threading.Lock()

    @staticmethod
    def add_task(msg):
        task_que = fd_message_que

        if task_que.full():
            task_que.queue.clear()

        try:
            task_que.put(msg, False)
        except Exception as abn:
            run_log.error("put message to queue failed: %s", abn)

    @staticmethod
    def dispatch_fd_messages():
        while True:
            try:
                msg = fd_message_que.get_nowait()
            except queue.Empty:
                time.sleep(0.5)
                continue

            try:
                MidwareProc.deal_fd_message(msg)
            except Exception as ex:
                run_log.error("deal fd message failed: %s", ex)

    @staticmethod
    def deal_fd_message(msg: FdMsgData):
        for register, handling_func in MSG_HANDLING_MAPPING.items():
            if isinstance(register, str) and register == msg.topic:
                handling_func(msg)

    @staticmethod
    def handling_msg_from_fd():
        message = "Report event failed."
        fd_ip = FdCfgManager.get_cur_fd_ip()
        if not fd_ip:
            run_log.error("get fd ip failed.")
            fd_ip = "AtlasEdge"
        operation_subject = "FD@{}".format(fd_ip)
        if MidwareProc.REPORT_ALARM_LOCK.locked():
            run_log.info("Report alarm is busy")
            operate_log.info("[%s] %s", operation_subject, message)
            return
        with MidwareProc.REPORT_ALARM_LOCK:
            operate_log.info("[%s] %s", operation_subject, "Report event executing.")
            try:
                if FileCheck.check_path_is_exist_and_valid(RESTARTING_FLAG):
                    run_log.info(f"The connection FusionDirector is ready and waits for "
                                 f"120 seconds to report the event.")
                    time.sleep(120)
                    MidwareProc.report_event()
                    message = "Report event successful."
                else:
                    message = "No events detected."
                    run_log.info("No event is detected and does not need to be reported.")
            except Exception as err:
                run_log.error("Exception: err_msg=%s, topic: %s", err, Topic.SUB_RESET_ALARM)
            finally:
                operate_log.info("[%s] %s", operation_subject, message)

    @staticmethod
    def report_event_to_fd(payload_publish):
        msg = FdMsgData.gen_ws_msg_obj(payload_publish, "websocket/alarm")
        # 发布告警上报topic
        publish_ws_msg(msg)
        run_log.info("publish topic: %s, send message: %s", msg.topic, payload_publish)

    @staticmethod
    def report_event():
        run_log.info("Report event start.")
        ret = MidwareProc.view_functions["espmanager/Event"]()
        if isinstance(ret, list) and ret[0] != 0:
            run_log.error("report event error. %s", ret)
            return

        report_list = ret[1]
        for payload_publish in report_list:
            MidwareProc.report_event_to_fd(payload_publish)

        run_log.info("Report event end.")
