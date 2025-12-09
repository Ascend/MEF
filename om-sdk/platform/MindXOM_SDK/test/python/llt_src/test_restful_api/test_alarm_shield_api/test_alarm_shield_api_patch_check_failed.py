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

from test_restful_api.test_z_main.restful_test_base import PatchTest


class TestPatchAlarmShieldFailed(PatchTest):
    ALARM_SHIELD_URL = "/redfish/v1/Systems/Alarm/AlarmShield/Increase"

    def __init__(self, expect_ret, code: int, label: str, data, json_flag=True):
        self.expect_ret = expect_ret
        self.data = data
        self.json_flag = json_flag
        super().__init__(url=self.ALARM_SHIELD_URL, code=code, data=self.get_data(), label=label)

    @staticmethod
    def get_template():
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

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response

    def get_data(self):
        if not self.json_flag:
            return self.data
        ret = self.get_template()
        ret.get("AlarmShieldMessages")[0].update(self.data)
        return ret


def init_patch_alarm_shield_failed_instances():
    TestPatchAlarmShieldFailed(expect_ret=json.dumps({
        "error": {
            "code": "Base.1.0.GeneralError",
            "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
            "@Message.ExtendedInfo": [
                {
                    "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                    "Description": "Indicates that a general error has occurred.",
                    "Message": "Parameter is invalid.",
                    "Severity": "Critical",
                    "NumberOfArgs": None,
                    "ParamTypes": None,
                    "Resolution": "None",
                    "Oem": {
                        "status": 100024
                    }
                }
            ]
        }
    }),
        label="test patch alarm shield failed due to UniquelyIdentifies is too invalid!",
        code=400,
        data={"UniquelyIdentifies": "a00000000"},
        json_flag=False)

    TestPatchAlarmShieldFailed(expect_ret=json.dumps({
        "error": {
            "code": "Base.1.0.GeneralError",
            "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
            "@Message.ExtendedInfo": [
                {
                    "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                    "Description": "Indicates that a general error has occurred.",
                    "Message": "Parameter is invalid.",
                    "Severity": "Critical",
                    "NumberOfArgs": None,
                    "ParamTypes": None,
                    "Resolution": "None",
                    "Oem": {
                        "status": 100024
                    }
                }
            ]
        }
    }),
        label="test patch edgeSystem failed due to AlarmId is too invalid!",
        code=400,
        data={"AlarmId": "0000000"},
        json_flag=False)

    TestPatchAlarmShieldFailed(expect_ret=json.dumps({
        "error": {
            "code": "Base.1.0.GeneralError",
            "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
            "@Message.ExtendedInfo": [
                {
                    "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                    "Description": "Indicates that a general error has occurred.",
                    "Message": "Parameter is invalid.",
                    "Severity": "Critical",
                    "NumberOfArgs": None,
                    "ParamTypes": None,
                    "Resolution": "None",
                    "Oem": {
                        "status": 100024
                    }
                }
            ]
        }
    }),
        label="test patch edgeSystem failed due to PerceivedSeverity is too invalid!",
        code=400,
        data={"PerceivedSeverity": "5"},
        json_flag=False)

    TestPatchAlarmShieldFailed(expect_ret=json.dumps({
        "error": {
            "code": "Base.1.0.GeneralError",
            "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
            "@Message.ExtendedInfo": [
                {
                    "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                    "Description": "Indicates that a general error has occurred.",
                    "Message": "Parameter is invalid.",
                    "Severity": "Critical",
                    "NumberOfArgs": None,
                    "ParamTypes": None,
                    "Resolution": "None",
                    "Oem": {
                        "status": 100024
                    }
                }
            ]
        }
    }),
        label="test patch edgeSystem failed due to AlarmInstance is too invalid!",
        code=400,
        data={"AlarmInstance": "MM"},
        json_flag=False)

    TestPatchAlarmShieldFailed(expect_ret=json.dumps({
        "error": {"code": "Base.1.0.MalformedJSON",
                  "message": "A MalformedJSON has occurred. See ExtendedInfo for more information.",
                  "@Message.ExtendedInfo":
                      [{"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                        "Description": "Indicates that the request body was malformed JSON.  "
                                       "Could be duplicate, syntax error,etc.",
                        "Message": "The request body submitted was malformed JSON "
                                   "and could not be parsed by the receiving service.",
                        "Severity": "Critical", "NumberOfArgs": 0, "ParamTypes": None,
                        "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                        "Oem": {"status": None}}]}}),
        label="data is not json!",
        code=400,
        data="AlarmInstance",
        json_flag=False)


init_patch_alarm_shield_failed_instances()
