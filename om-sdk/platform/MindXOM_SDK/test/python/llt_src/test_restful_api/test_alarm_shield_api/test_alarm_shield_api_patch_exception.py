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
from unittest.mock import patch

from test_restful_api.test_z_main.restful_test_base import PatchTest


class TestPatchAlarmShieldException(PatchTest):
    ALARM_SHIELD_URL = "/redfish/v1/Systems/Alarm/AlarmShield/Increase"

    def __init__(self, expect_ret, code: int, label: str):
        self.expect_ret = expect_ret
        self.patch = None
        super().__init__(url=self.ALARM_SHIELD_URL, code=code, data=self.get_data(), label=label)

    def get_data(self):
        return (
            {
                "AlarmShieldMessages":
                    [
                        {
                            "UniquelyIdentifies": "a000000000",
                            "AlarmId": "00000000",
                            "PerceivedSeverity": "2",
                            "AlarmInstance": "M.2"
                        }
                    ]
            }
        )

    def before(self):
        self.patch = patch("lib_restful_adapter.LibRESTfulAdapter.lib_restful_interface",
                           side_effect=Exception())
        self.patch.start()

    def after(self):
        if self.patch:
            self.patch.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def init_patch_alarm_shield_instances_exception():
    TestPatchAlarmShieldException(expect_ret=json.dumps({
        "error": {
            "code": "Base.1.0.GeneralError",
            "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
            "@Message.ExtendedInfo": [
                {
                    "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                    "Description": "Indicates that a general error has occurred.",
                    "Message": "Increase alarm shield failed.",
                    "Severity": "Critical",
                    "NumberOfArgs": None,
                    "ParamTypes": None,
                    "Resolution": "None",
                    "Oem": {
                        "status": None
                    }
                }
            ]
        }
    }),
                                  label="test patch alarm shield failed due to lib_restful_interface exception!",
                                  code=400,
                                  )


init_patch_alarm_shield_instances_exception()