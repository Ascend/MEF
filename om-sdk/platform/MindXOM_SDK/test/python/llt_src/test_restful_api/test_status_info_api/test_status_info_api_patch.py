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

from test_restful_api.test_z_main.restful_test_base import PatchTest


class TestPatchStatusInfoFailed(PatchTest):
    NTP_SERVICE_URL = "redfish/v1/Systems/LTE/StatusInfo"

    def __init__(self, expect_ret, code: int, label: str, data: dict):
        self.expect_ret = expect_ret
        self.data = data
        super().__init__(url=self.NTP_SERVICE_URL, code=code, data=self.get_data(), label=label)

    @staticmethod
    def get_template():
        return (
            {
                "default_gateway": False,
                "lte_enable": True,
                "sim_exist": True,
                "state_lte": True,
                "state_data": True,
                "network_signal_level": 4,
                "network_type": "4G",
                "ip_addr": "10.21.33.44"
            }
        )

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response

    def get_data(self):
        ret = self.get_template()
        ret.update(self.data)
        return ret


def init_patch_status_info_instances():
    TestPatchStatusInfoFailed(expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 100024}}]}}),
                              label="test patch status info failed due to parameter state_data is invalid!",
                              code=400,
                              data={"state_lte": True, "state_data": "false"})

    TestPatchStatusInfoFailed(expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 100024}}]}}),
                              label="test patch status info failed due to parameter state_lte is invalid!",
                              code=400,
                              data={"state_lte": "true", "state_data": False})

    TestPatchStatusInfoFailed(expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 100024}}]}}),
                              label="test patch status info failed due to parameter state_data is None!",
                              code=400,
                              data={"state_lte": True, "state_data": ""})

    TestPatchStatusInfoFailed(expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 100024}}]}}),
                              label="test patch status info failed due to parameter state_lte is None!",
                              code=400,
                              data={"state_lte": "", "state_data": False})


init_patch_status_info_instances()
