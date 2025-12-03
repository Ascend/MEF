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


class TestPatchSessionServiceCollection(PatchTest):
    ACCOUNTS_SERVICE_COLLECTION_URL = "/redfish/v1/AccountService"

    def __init__(self, expect_ret, data, code: int, label: str):
        self.patch = None
        self.expect_ret = expect_ret
        super().__init__(url=self.ACCOUNTS_SERVICE_COLLECTION_URL,
                         code=code,
                         label=label,
                         data=data
                         )

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def init_patch_account_service_collection():
    TestPatchSessionServiceCollection(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.MalformedJSON",
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
        code=400,
        label="modify password validity failed due to data is not json.",
        data="test",
    ),
    TestPatchSessionServiceCollection(
        expect_ret=json.dumps(
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
        label="modify password validity failed due to PasswordExpirationDays is wrong.",
        data={"PasswordExpirationDays": 366, "Password": "test1234!"},
    )


init_patch_account_service_collection()