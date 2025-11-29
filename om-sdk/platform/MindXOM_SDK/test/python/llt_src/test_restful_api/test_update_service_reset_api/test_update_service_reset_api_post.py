# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json

from test_restful_api.test_z_main.restful_test_base import PostTest
from ut_utils.models import MockPrivilegeAuth


class TestPostUpdateServiceResetFailed(PostTest):
    UPDATE_SERVICE_RESET_URL = "/redfish/v1/UpdateService/Actions/UpdateService.Reset"

    def __init__(self, expect_ret, code: int, label: str, data: dict):
        self.expect_ret = expect_ret
        self.data = data
        super().__init__(url=self.UPDATE_SERVICE_RESET_URL, code=code, label=label, data=data)

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def test_post_update_service_reset_instances(mocker):
    mocker.patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth)
    TestPostUpdateServiceResetFailed(expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": None}}]}}),
                                     label="test post update service reset failed \
                                     due to parameter ResetType is null!",
                                     code=400,
                                     data={"ResetType": ""})

    TestPostUpdateServiceResetFailed(expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "Parameter is invalid.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": None}}]}}),
                                     label="test post update service reset failed \
                                     due to parameter ResetType is invalid!",
                                     code=400,
                                     data={"ResetType": "xxx"})

