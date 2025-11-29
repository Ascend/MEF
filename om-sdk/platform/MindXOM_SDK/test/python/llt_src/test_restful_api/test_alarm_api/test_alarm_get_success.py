# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json

from unittest.mock import patch

from test_restful_api.test_z_main.restful_test_base import GetTest


class TestGetAlarmSuccess(GetTest):
    ALARM_URL = "/redfish/v1/Systems/Alarm/AlarmInfo"

    def __init__(self, expect_ret, code: int, label: str):
        self.expect_ret = expect_ret
        self.patch = None
        super().__init__(url=self.ALARM_URL, code=code, label=label)

    def before(self):
        self.patch = patch("lib_restful_adapter.LibRESTfulAdapter.lib_restful_interface",
                           return_value={"status": 200, "message": {"AlarMessages": ["test"]}})
        self.patch.start()

    def after(self):
        if self.patch:
            self.patch.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def init_test_get_alarm():
    TestGetAlarmSuccess(expect_ret=json.dumps({
        "@odata.context": "/redfish/v1/$metadata#Systems/Alarm/AlarmInfo",
        "@odata.id": "/redfish/v1/Systems/Alarm/AlarmInfo",
        "@odata.type": "MindXEdgeAlarm.v1_0_0.MindXEdgeAlarm",
        "Id": "Alarm Info",
        "Name": "Alarm Info",
        "AlarMessages": ["test"]}),
        label="test get alarm success",
        code=200,
    )


init_test_get_alarm()
