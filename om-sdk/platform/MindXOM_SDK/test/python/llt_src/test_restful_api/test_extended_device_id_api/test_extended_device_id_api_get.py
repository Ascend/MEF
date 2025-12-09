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

from test_restful_api.test_z_main.restful_test_base import GetTest


class TestGetExtendedDeviceIdFailed(GetTest):
    EXTENDED_DEVICE_URL = "/redfish/v1/Systems/ExtendedDevices/"

    def __init__(self, expect_ret, code: int, label: str, device_id: str):
        send_url = self.EXTENDED_DEVICE_URL + str(device_id)
        super().__init__(url=send_url, code=code, label=label)
        self.expect_ret = expect_ret

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def init_get_extended_device_id_instances():
    TestGetExtendedDeviceIdFailed(expect_ret=json.dumps({
                "error": {
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "The requested URL was not found on the server",
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
                                  code=404,
                                  label="test extended device id failed due to device_id is NULL",
                                  device_id='')
    TestGetExtendedDeviceIdFailed(expect_ret=json.dumps(
        {"error": {"code": "Base.1.0.GeneralError",
                   "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                   "@Message.ExtendedInfo": [
                       {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                        "Description": "Indicates that a general error has occurred.",
                        "Message": "Parameter is invalid.",
                        "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                        "Resolution": "None",
                        "Oem": {"status": 100024}}]}}),
                                  code=400,
                                  label="test extended device id failed due to device_id is '!'",
                                  device_id='!')
    TestGetExtendedDeviceIdFailed(expect_ret=json.dumps(
        {"error": {"code": "Base.1.0.GeneralError",
                   "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                   "@Message.ExtendedInfo": [
                       {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                        "Description": "Indicates that a general error has occurred.",
                        "Message": "Parameter is invalid.",
                        "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                        "Resolution": "None",
                        "Oem": {"status": 100024}}]}}),
                                  code=400,
                                  label="test extended device id failed due to device_id is 'x'",
                                  device_id='x')


init_get_extended_device_id_instances()
