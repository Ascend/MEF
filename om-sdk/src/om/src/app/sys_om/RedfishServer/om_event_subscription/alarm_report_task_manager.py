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

from typing import List, Iterable, NoReturn

from sqlalchemy import not_

from om_event_subscription.models import AlarmReportTask
from om_event_subscription.subscription_mgr import BaseManager
from redfish_db.session import session_maker


class AlarmReportTaskManager(BaseManager):
    model = AlarmReportTask

    def get_by_status(self, *task_status: int) -> List[AlarmReportTask]:
        with session_maker() as session:
            task_list = session.query(self.model).filter(self.model.task_status.in_(task_status)).all()
            for task in task_list:
                session.expunge(task)
            return task_list

    def update_task_by_subscriber_id_event_ids_task_status(
            self,
            subscriber_id: int,
            event_ids: List[str],
            original_status: int,
            update_data: dict
    ) -> NoReturn:
        with session_maker() as session:
            session.query(self.model).filter(
                self.model.event_id.in_(event_ids),
                self.model.subscriber_id == subscriber_id,
                self.model.task_status == original_status
            ).update(update_data)

    def delete_by_status_and_ids(self, task_status: Iterable[int], subscriber_ids: Iterable[int]) -> NoReturn:
        with session_maker() as session:
            session.query(self.model).filter(
                self.model.task_status.in_(task_status),
                not_(self.model.subscriber_id.in_(subscriber_ids))
            ).delete()
