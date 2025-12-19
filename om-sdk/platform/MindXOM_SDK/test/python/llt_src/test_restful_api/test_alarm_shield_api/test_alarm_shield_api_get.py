# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json
from unittest.mock import patch

from test_restful_api.test_z_main.restful_test_base import GetTest


class TestGetAlarmShield(GetTest):
    ALARM_URL = "/redfish/v1/Systems/Alarm/AlarmShield"

    def __init__(self, expect_ret, code: int, label: str):
        self.expect_ret = expect_ret
        self.patch = None
        super().__init__(url=self.ALARM_URL, code=code, label=label)

    def before(self):
        self.patch = patch("lib_restful_adapter.LibRESTfulAdapter.lib_restful_interface",
                           return_value={"status": 200, "message": {"AlarmShieldMessages": ["test"]}})
        self.patch.start()

    def after(self):
        if self.patch:
            self.patch.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def init_test_get_alarm():
    TestGetAlarmShield(expect_ret=json.dumps({
        "@odata.context": "/redfish/v1/$metadata#Systems/Alarm/AlarmShield",
        "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield",
        "@odata.type": "MindXEdgeAlarm.v1_0_0.MindXEdgeAlarm",
        "Id": "Alarm Shield",
        "Name": "Alarm Shield",
        "AlarmShieldMessages": ["test"],
        "Increase": {
            "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield/Increase"
        },
        "Decrease": {
            "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield/Decrease"
        }
    }),
                       label="test get alarm success",
                       code=200,
                       )


init_test_get_alarm()
