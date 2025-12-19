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

from test_restful_api.test_z_main.restful_test_base import PatchTest


class TestPatchNTPService(PatchTest):
    NTP_SERVICE_URL = "/redfish/v1/Systems/NTPService"

    def __init__(self, expect_ret, code: int, label: str, data: dict, patch_return=None, json_flag=True):
        self.expect_ret = expect_ret
        self.json_flag = json_flag
        self.data = data
        self.patch = None
        self.patch_return = patch_return
        super().__init__(url=self.NTP_SERVICE_URL, code=code, data=self.get_data(), label=label)

    @staticmethod
    def get_template():
        return (
            {
                "ClientEnabled": False,
                "ServerEnabled": False,
                "NTPRemoteServers": "192.168.2.10",
                "NTPRemoteServersbak": None,
                "NTPLocalServers": "51.38.68.206",
                "Target": "Client"
            }
        )

    def before(self):
        self.patch = patch("lib_restful_adapter.LibRESTfulAdapter.lib_restful_interface",
                           return_value=self.patch_return)
        self.patch.start()

    def after(self):
        if self.patch:
            self.patch.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response

    def get_data(self):
        if not self.json_flag:
            return self.data
        ret = self.get_template()
        ret.update(self.data)
        return ret


def init_patch_ntp_service_instances():
    TestPatchNTPService(expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 100024}}]}}),
                        label="test patch NTP Service failed due to NTPRemoteServers  IP is invalid!",
                        code=400,
                        data={"ClientEnabled": True, "NTPRemoteServers": "192.168.2.300"})

    TestPatchNTPService(expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 100024}}]}}),
                        label="test patch NTP Service failed due to parameter NTPRemoteServers \
                              is the same with NTPRemoteServersbak!",
                        code=400,
                        data={"ClientEnabled": True, "NTPRemoteServers": "192.168.2.3",
                              "NTPRemoteServersbak": "192.168.2.3"})

    TestPatchNTPService(expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 100024}}]}}),
                        label="test patch NTP Service failed due to Target is invalid!",
                        code=400,
                        data={"Target": "Client1"})

    TestPatchNTPService(expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 100024}}]}}),
                        label="test patch NTP Service failed due to ClientEnabled is invalid!",
                        code=400,
                        data={"ClientEnabled": "True"})

    TestPatchNTPService(expect_ret=json.dumps(
        {"error": {"code": "Base.1.0.MalformedJSON",
                   "message": "A MalformedJSON has occurred. See ExtendedInfo for more information.",
                   "@Message.ExtendedInfo": [
                       {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                        "Description": "Indicates that the request body was malformed JSON.  "
                                       "Could be duplicate, syntax error,etc.",
                        "Message": "The request body submitted was malformed JSON "
                                   "and could not be parsed by the receiving service.",
                        "Severity": "Critical", "NumberOfArgs": 0, "ParamTypes": None,
                        "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                        "Oem": {"status": None}}]}}),
        label="test patch NTP Service failed due to data is not json!",
        code=400,
        data="ClientEnabled",
        json_flag=False
    )

    TestPatchNTPService(expect_ret=json.dumps({
                            "error": {
                                "code": "Base.1.0.GeneralError",
                                "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                                "@Message.ExtendedInfo": [{
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Internal server error",
                                "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {
                                    "status": 100011
                                }}]}
                        }),
                        label="test patch NTP Service exception!",
                        code=500,
                        data={"ClientEnabled": False},
                        patch_return={"status": 200, "message": "test"})
    TestPatchNTPService(expect_ret=json.dumps({
                                "@odata.type": "#MindXEdgeNTPService.v1_0_0.MindXEdgeNTPService",
                                "@odata.context": "/redfish/v1/$metadata#Systems/NTPService",
                                "@odata.id": "/redfish/v1/Systems/NTPService",
                                "Id" : "NTPService",
                                "Name": "NTPService",
                                "ClientEnabled": None,
                                "ServerEnabled": None,
                                "Port": None,
                                "NTPRemoteServers": None,
                                "NTPRemoteServersbak": None,
                                "NTPLocalServers": None
                                }),
                        label="test patch NTP Service success!",
                        code=200,
                        data={"ClientEnabled": False},
                        patch_return={"status": 200, "message": {"test": "test"}})


init_patch_ntp_service_instances()
