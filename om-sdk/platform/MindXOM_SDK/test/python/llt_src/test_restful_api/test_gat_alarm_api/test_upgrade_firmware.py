# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json

from test_restful_api.test_z_main.restful_test_base import PostTest
from ut_utils.models import MockPrivilegeAuth


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
    mocker.patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth)
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
