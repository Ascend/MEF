#!/usr/bin/python3
# -*- coding: UTF-8 -*-
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
from enum import Enum
from typing import Dict, Callable

from common.log.logger import run_log
from common.utils.result_base import Result
from om_event_subscription.report_alarm_task import ReportTaskManager
from om_event_subscription.subscription_mgr import PreSubCertMgr


def start_esight_report_timer():
    report_task_manager = ReportTaskManager()
    # 启动生成告警事件定时任务
    run_log.info("gene_alarm_tasks_timer start")
    threading.Timer(30, report_task_manager.gene_alarm_tasks_timer).start()
    # 启动上报告警事件的定时任务,比生成告警定时任务晚5秒启动，保证先生成再上报
    run_log.info("report_alarm_tasks_timer start")
    threading.Timer(35, report_task_manager.report_alarm_tasks_timer).start()
    # 启动清除已上报告警定时任务
    run_log.info("clear_alarm_tasks_timer start")
    threading.Timer(40, report_task_manager.clear_alarm_tasks_timer).start()


def start_extend_funcs():
    start_esight_report_timer()


class Action(Enum):
    GETUNUSED = "getunusedcert"
    RESTORE = "restorecert"
    DELETE = "deletecert"


def get_unused_cert() -> Result:
    try:
        ret = PreSubCertMgr().get_pre_subs_cert()
    except Exception:
        run_log.error("Get unused subscriptions cert failed.")
        return Result(False, err_msg="Get unused subscriptions cert failed.")

    run_log.info("Get unused subscriptions cert successfully.")
    return Result(True, data=json.dumps(ret.get_cert_crl()))


def restore_cert() -> Result:
    try:
        PreSubCertMgr().restore_pre_subs_cert()
    except Exception:
        run_log.error("Restore subscriptions cert failed.")
        return Result(False, err_msg="Restore subscriptions cert failed.")

    run_log.info("Restore subscriptions cert successfully.")
    return Result(True, data="Restore subscriptions cert successfully.")


def delete_cert() -> Result:
    if not get_unused_cert():
        run_log.error("Delete unused subscriptions cert failed. Because unused subscriptions cert not exist.")
        return Result(False, err_msg="Delete unused subscriptions cert failed.")

    if PreSubCertMgr().delete_cert():
        run_log.info("Delete unused subscriptions cert successfully.")
        return Result(True, data="Delete unused subscriptions cert successfully.")

    run_log.error("Delete unused subscriptions cert failed.")
    return Result(False, err_msg="Delete subscriptions unused cert failed.")


OPERATE: Dict[str, Callable[[], Result]] = {
    Action.GETUNUSED.value: get_unused_cert,
    Action.RESTORE.value: restore_cert,
    Action.DELETE.value: delete_cert,
}


def extend_redfish_cert_manage(action: str) -> Result:
    operate = OPERATE[action]
    return operate()
