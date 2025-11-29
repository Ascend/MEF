# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json

from mock import patch

from test_restful_api.test_z_main.restful_test_base import GetTest


class TestGetAlarmFailed(GetTest):
    ALARM_URL = "/redfish/v1/Systems/Alarm/AlarmInfo"

    def __init__(self, expect_ret, code: int, label: str):
        self.expect_ret = expect_ret
        self.patch = None
        super().__init__(url=self.ALARM_URL, code=code, label=label)

    def before(self):
        self.patch = patch("lib_restful_adapter.LibRESTfulAdapter.lib_restful_interface", side_effect=Exception())
        self.patch.start()

    def after(self):
        if self.patch:
            self.patch.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def init_test_get_alarm():
    TestGetAlarmFailed(expect_ret=json.dumps({
        "error": {
            "code": "Base.1.0.GeneralError",
            "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
            "@Message.ExtendedInfo": [
                {
                    "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                    "Description": "Indicates that a general error has occurred.",
                    "Message": "Get Alarm info failed.",
                    "Severity": "Critical",
                    "NumberOfArgs": None,
                    "ParamTypes": None,
                    "Resolution": "None",
                    "Oem": {
                        "status": None
                    }
                }
            ]
        }}),
        label="test get alarm failed due to lib_restful_interface exception!",
        code=400,
    )


init_test_get_alarm()
