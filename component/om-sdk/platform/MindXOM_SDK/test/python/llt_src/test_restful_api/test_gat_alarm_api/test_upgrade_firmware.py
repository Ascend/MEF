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

from test_restful_api.test_z_main.restful_test_base import PostTest
from upgrade_service.upgrade_blueprint import privilege_auth


class TestUpgradeFirmwareFailed(PostTest):
    """2.3.1 升级固件"""
    COMPUTER_SYSTEM_RESET_URL = "/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate"

    def __init__(self, expect_ret, code: int, label: str, data: dict):
        self.expect_ret = expect_ret
        self.data = data
        super().__init__(url=self.COMPUTER_SYSTEM_RESET_URL,
                         code=code,
                         data=data,
                         label=label)

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def test_upgrade_firmware(mocker):
    mocker.patch.object(privilege_auth, "token_required")
    TestUpgradeFirmwareFailed(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": None}}]}}),
        label="test upgrade firmware failed due to ImageURI is invalid",
        code=400,
        data={
            "ImageURI": "",
            "TransferProtocol": "https"
        })
