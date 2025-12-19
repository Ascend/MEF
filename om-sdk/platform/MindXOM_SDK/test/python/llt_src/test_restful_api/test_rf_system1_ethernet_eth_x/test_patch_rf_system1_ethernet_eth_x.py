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


class TestPatchRfSystem1EthernetCollection(PatchTest):
    """修改以太网接口资源信息"""
    RF_SYSTEM1_ETHERNET_ETH_X_URL = "/redfish/v1/Systems/EthernetInterfaces/"

    def __init__(self, expect_ret, code: int, label: str, data, eth_id, patch_return):
        self.expect_ret = expect_ret
        self.patch = None
        self.patch_check_external_parameter = None
        self.patch_return = patch_return
        super().__init__(url=self.RF_SYSTEM1_ETHERNET_ETH_X_URL + str(eth_id),
                         code=code,
                         label=label,
                         data=data
                         )

    def before(self):
        self.patch = patch("lib_restful_adapter.LibRESTfulAdapter.lib_restful_interface",
                           return_value=self.patch_return)
        self.patch.start()

        self.patch_check_external_parameter = patch("ibma_redfish_urls.RedfishGlobals.check_external_parameter",
                                                    return_value=None)
        self.patch_check_external_parameter.start()

    def after(self):
        if self.patch:
            self.patch.stop()

        if self.patch_check_external_parameter:
            self.patch_check_external_parameter.stop()

    def call_back_assert(self, test_response: str):
        # 调用成功时，时间会一直变化，没法比较时间，所以比较里面的其他字段
        if self._code == 200:
            assert "TaskState" in test_response
        else:
            assert self.expect_ret == test_response


def test_patch_rf_system1_ethernet_eth_x():
    TestPatchRfSystem1EthernetCollection(
        expect_ret=json.dumps(
            {"error": {
                "code": "Base.1.0.MalformedJSON",
                "message": "A MalformedJSON has occurred. See ExtendedInfo for more information.",
                "@Message.ExtendedInfo": [
                    {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                     "Description": "Indicates that the request body was malformed JSON.  "
                                    "Could be duplicate, syntax error,etc.",
                     "Message": "The request body submitted was malformed JSON "
                                "and could not be parsed by the receiving service.",
                     "Severity": "Critical", "NumberOfArgs": 0, "ParamTypes": None,
                     "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                     "Oem": {"status": None}}]}}
        ),
        code=400,
        label="Modify ethernet interface failed due to data is not json.",
        data="test",
        patch_return=None,
        eth_id="GMAC0"
    )
    TestPatchRfSystem1EthernetCollection(
        expect_ret=json.dumps(
            {"@odata.context": "/redfish/v1/$metadata#Systems/Members/EthernetInterfaces/Members/$entity",
             "@odata.id": "/redfish/v1/Systems/EthernetInterfaces/GMAC0",
             "@odata.type": "#EthernetInterface.v1_8_0.EthernetInterface",
             "Id": "GMAC0", "Name": "Set mainboardNIC1Port1 odatainfo",
             "IPv4Addresses": [{"Address": "test", "Tag": "default"}],
             "Oem": {"StartTime": "2022-08-25T19:21:16+0800", "TaskState": "Running", "TaskPercentage": "ok"}}
        ),
        code=200,
        label="Modify ethernet interface success.",
        data={"IPv4Addresses": [{"Address": "test"}]},
        patch_return={"status": 200, "message": {"eth_id": "GMAC0"}},
        eth_id="GMAC0"
    )


test_patch_rf_system1_ethernet_eth_x()
