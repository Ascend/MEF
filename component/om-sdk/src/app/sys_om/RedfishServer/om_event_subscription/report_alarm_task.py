#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

import contextlib
import itertools
import json
import time
from datetime import datetime, timezone
from functools import partial
from typing import List, Iterable

from urllib3 import PoolManager
from urllib3.exceptions import ConnectTimeoutError, SSLError

from common.log.logger import run_log
from common.utils.timer import RepeatingTimer
from lib_restful_adapter import LibRESTfulAdapter
from om_event_subscription.active_alarm_manager import ActiveAlarmManager
from om_event_subscription.alarm_report_task_manager import AlarmReportTaskManager
from om_event_subscription.constants import (TaskStatus, TASK_ALARM_STATUS_RELATIONSHIP,
                                             ALARM_SEVERITY_MAPPING, RESP_STATUS_LIST, ALARM_ID_NUM, ALARM_SEPARATOR)
from om_event_subscription.models import (ActiveAlarm, AlarmReportTask, Subscription)
from om_event_subscription.param_checker import DestinationChecker
from om_event_subscription.subscription_client import create_clients
from om_event_subscription.subscription_mgr import SubscriptionsMgr
from om_event_subscription.tools import group_by


class ReportTaskManager:
    """
    告警上报任务管理类，
    用于计算需要上报的告警任务，并上报告警
    """

    def __init__(self):
        self.subscription_manager = SubscriptionsMgr()
        self.task_manager = AlarmReportTaskManager()
        self.alarm_manager = ActiveAlarmManager()

    @staticmethod
    def alarms_to_tasks(subscribers: List[Subscription], alarms: List[ActiveAlarm], task_status: int):
        for subscriber, alarm in itertools.product(subscribers, alarms):
            if not alarm.severity and not ALARM_SEVERITY_MAPPING.get("eSight").get(int(alarm.severity)):
                continue

            yield AlarmReportTask(
                subscriber_id=subscriber.id,
                event_id=ALARM_SEPARATOR.join((alarm.alarm_id, alarm.alarm_instance, str(alarm.id))),
                event_type=alarm.type,
                event_name=alarm.alarm_name,
                event_timestamp=(alarm.create_time
                                 if task_status == TaskStatus.WAIT_ADD.value
                                 else datetime.now(tz=timezone.utc).astimezone()),
                task_status=task_status,
                severity=ALARM_SEVERITY_MAPPING.get("eSight").get(int(alarm.severity))
            )

    @staticmethod
    def get_rf_alarms() -> Iterable[ActiveAlarm]:
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Alarm", "GET", None, False, None)
        if not all((isinstance(ret_dict.get("message"), dict),
                    isinstance(ret_dict.get("message").get("AlarMessages"), list))):
            return

        for data in ret_dict["message"]["AlarMessages"]:
            if not isinstance(data, dict):
                continue
            yield ActiveAlarm.from_dict(data)

    @staticmethod
    def report_tasks(subscriber: Subscription, manager: PoolManager, report_data: dict):
        if not manager or not subscriber:
            return False

        target = subscriber.get_decrypt_destination()
        is_target_valid = DestinationChecker("target").check({"target": target})
        if not is_target_valid:
            run_log.error("report to subscriber %d failed reason: %s", subscriber.id, is_target_valid.reason)
            return False

        try:
            resp = manager.request(
                "POST", target,
                body=json.dumps(report_data),
                preload_content=False,
                headers={"X-Auth-Token": subscriber.get_decrypt_credential()}
            )
        except ConnectTimeoutError:
            run_log.error("report to subscriber %d failed reason: connect time out", subscriber.id)
            return False
        except SSLError:
            run_log.error("report to subscriber %d failed reason: cert is wrong", subscriber.id)
            return False
        except Exception:
            run_log.error("report to subscriber %d failed", subscriber.id)
            return False

        if resp and resp.status in RESP_STATUS_LIST:
            run_log.info("report to subscriber %d success", subscriber.id)
            return True

        run_log.warning("report to subscriber %d failed", subscriber.id)
        return False

    @staticmethod
    def generate_report_data(tasks: List[dict], subscriber_id: int):

        event_list = []
        for task in tasks:
            task_status = task.get("task_status")
            alarm_status = TASK_ALARM_STATUS_RELATIONSHIP.get(task_status)
            if alarm_status is None:
                continue

            event_id = task.get("event_id", "")
            alarm_id_instance = event_id.split(ALARM_SEPARATOR)
            if len(alarm_id_instance) != ALARM_ID_NUM:
                continue
            alarm_id = alarm_id_instance[0]
            event_time = time.strftime(
                "%Y-%m-%d %H:%M:%S",
                time.localtime(int(task.get("event_timestamp", datetime.now(tz=timezone.utc)).timestamp()))
            )
            event = {
                "EventType": task.get("event_type"),
                "EventId": event_id,
                "Severity": task.get("severity"),
                "EventTimestamp": event_time,
                "Oem": {
                    "Public": {
                        "specificProblemID": alarm_id,
                        "specificProblem": task.get("event_name"),
                        "objectName": alarm_id_instance[1],
                        "alarmStatus": alarm_status
                    }
                }
            }
            event_list.append(event)

        if not event_list:
            return {}

        return {
            "@odata.context": "/redfish/v1/$metadata#Event.Event",
            "@odata.id": f"/redfish/v1/EventService/Events/{subscriber_id}",
            "@odata.type": "#Event.v1_1_0.Event",
            "Id": subscriber_id,
            "Name": "Event Array",
            "Context": "event subscription context string",
            "Events": event_list
        }

    def delete_task_without_subscriber(self):
        subscriber_ids = (subscriber.id for subscriber in self.subscription_manager.get_valid_subscriptions())
        self.task_manager.delete_by_status_and_ids(TaskStatus.wait_add_and_delete(), subscriber_ids)

    def delete_invalid_by_event_id(self, wait_add_list, wait_delete_list, all_alarm_list):
        ids_to_delete: List[int] = []

        # 获取当前在数据库中待添加但不在总告警中的告警id
        for item in wait_add_list:
            if item.event_id[:item.event_id.rindex(ALARM_SEPARATOR)] not in all_alarm_list:
                ids_to_delete.append(item.id)

        # 获取当前在数据库中待删除且在总告警中的告警id
        for item in wait_delete_list:
            if item.event_id[:item.event_id.rindex(ALARM_SEPARATOR)] in all_alarm_list:
                ids_to_delete.append(item.id)

        if not ids_to_delete:
            return

        self.task_manager.delete_by_ids(ids_to_delete)

    def delete_invalid_task(self, rf_alarms: Iterable[ActiveAlarm]):
        # 查询所有的待新增、待清除任务
        exist_wait_add_list = self.task_manager.get_by_status(TaskStatus.WAIT_ADD.value)
        exist_wait_delete_list = self.task_manager.get_by_status(TaskStatus.WAIT_DELETE.value)

        # 根据告警id和告警实体，当前发生的告警与任务做比较，计算出失效任务
        rf_alarm_event_list = [ALARM_SEPARATOR.join([alarm.alarm_id, alarm.alarm_instance]) for alarm in rf_alarms]

        # 删除当前失效任务
        self.delete_invalid_by_event_id(exist_wait_add_list, exist_wait_delete_list, rf_alarm_event_list)

    def generate_tasks(self, subscribers: List[Subscription], alarms: List[ActiveAlarm], status: int):
        if not subscribers:
            return

        self.task_manager.add_objs(self.alarms_to_tasks(subscribers, alarms, status))

    def modify_task_status_to_finished(self, task_list, subscriber_id):
        task_status_dict = group_by(task_list, "task_status")

        add_task_list = task_status_dict.get(TaskStatus.WAIT_ADD.value)
        if add_task_list:
            event_id_list = [task.get("event_id") for task in add_task_list]
            self.task_manager.update_task_by_subscriber_id_event_ids_task_status(
                subscriber_id, event_id_list, TaskStatus.WAIT_ADD.value, {"task_status": TaskStatus.ADDED.value})

        delete_task_list = task_status_dict.get(TaskStatus.WAIT_DELETE.value)
        if delete_task_list:
            event_id_list = [task.get("event_id") for task in delete_task_list]
            self.task_manager.update_task_by_subscriber_id_event_ids_task_status(
                subscriber_id, event_id_list, TaskStatus.WAIT_DELETE.value, {"task_status": TaskStatus.DELETED.value})

    def report_to_all_subscribers(self, cert_crl_client, cert_client, dict_tasks, subscribers):
        task_group = group_by(dict_tasks, "subscriber_id")
        for subscriber in subscribers:
            task_list = task_group.get(subscriber.id)
            report_data = self.generate_report_data(task_list, subscriber.id)
            if not report_data.get("Events"):
                continue
            # 先发送包含吊销列表的请求
            report = partial(ReportTaskManager.report_tasks, subscriber=subscriber, report_data=report_data)
            if report(manager=cert_crl_client) or report(manager=cert_client):
                # 上报成功，修改任务状态为已完成
                self.modify_task_status_to_finished(task_list, subscriber.id)

    def gen_alarm_report_task(self):
        try:
            # 1. 获取/run/all_active_alarm文件里的告警
            cur_active_alarms = set(self.get_rf_alarms())

            # 2. 查询订阅者，删除已取消订阅的订阅者的任务
            self.delete_task_without_subscriber()

            # 3. 删除失效任务
            self.delete_invalid_task(cur_active_alarms)

            # 4. 获取active_alarm表里的告警
            last_active_alarms = set(self.alarm_manager.get_all()) if self.alarm_manager.get_all() else set()

            # 5. 计算待新增待消除告警
            wait_adds = cur_active_alarms.difference(last_active_alarms)
            wait_deletes = last_active_alarms.difference(cur_active_alarms)

            # 6. 生成告警任务,更新active_alarm表
            subscribers = self.subscription_manager.get_valid_subscriptions()
            if wait_adds:
                self.alarm_manager.add_objs(list(wait_adds))
                self.generate_tasks(subscribers, list(wait_adds), TaskStatus.WAIT_ADD.value)

            if wait_deletes:
                self.generate_tasks(subscribers, list(wait_deletes), TaskStatus.WAIT_DELETE.value)
                self.alarm_manager.delete_by_ids([alarm.id for alarm in wait_deletes])
        except Exception:
            run_log.error("generate alarm report task failed.")

    def cal_and_report_task(self):
        try:
            # 1. 查询待消除待新增任务
            tasks = self.task_manager.get_by_status(*TaskStatus.wait_add_and_delete())
            if not tasks:
                return

            dict_tasks = (task.to_dict() for task in tasks)

            # 2. 查询订阅者
            subscribers = self.subscription_manager.get_valid_subscriptions()
            if not subscribers:
                return

            # 3. 创建客户端
            cert_crl_manager, cert_manager = create_clients()
            if not cert_crl_manager and not cert_manager:
                run_log.error("Clients are None")
                return

            with contextlib.ExitStack() as stack:
                cert_crl_client = stack.enter_context(cert_crl_manager()) if cert_crl_manager else None
                cert_client = stack.enter_context(cert_manager()) if cert_manager else None
                self.report_to_all_subscribers(cert_crl_client, cert_client, dict_tasks, subscribers)
        except Exception:
            run_log.error("calculate report task failed.")

    def delete_added_deleted_task(self):
        try:
            self.task_manager.delete_by_status_and_ids(TaskStatus.added_and_deleted(), ())
        except Exception:
            run_log.error("delete added and deleted task failed.")

    def gene_alarm_tasks_timer(self):
        RepeatingTimer(60, self.gen_alarm_report_task).start()

    def report_alarm_tasks_timer(self):
        RepeatingTimer(60, self.cal_and_report_task).start()

    def clear_alarm_tasks_timer(self):
        RepeatingTimer(86400, self.delete_added_deleted_task).start()
