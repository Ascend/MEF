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
from collections import namedtuple

from pytest_mock import MockerFixture

from common.utils.result_base import Result
from redfish_database_monitor import RedfishDatabaseMonitor

MonitorDatabaseStatusCase = namedtuple("MonitorDatabaseStatusCase", "check_database_is_valid")


class TestRedfishDatabaseMonitor:
    use_cases = {
        "test_monitor_database_status": {
            "normal": (Result(False), ),
            "false": (Result(True),),
        },

    }

    def test_rf_start_server(self):
        assert not RedfishDatabaseMonitor.report_database_alarm_to_fd(RedfishDatabaseMonitor(), "alarm")

    def test_monitor_database_status(self, mocker: MockerFixture, model: MonitorDatabaseStatusCase):
        mocker.patch.object(RedfishDatabaseMonitor, "check_database_is_valid",
                            return_value=model.check_database_is_valid)
        assert not RedfishDatabaseMonitor.monitor_database_status(RedfishDatabaseMonitor())
