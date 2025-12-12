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

from common.log.logger import run_log
from om_event_subscription.report_alarm_task import ReportTaskManager


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
