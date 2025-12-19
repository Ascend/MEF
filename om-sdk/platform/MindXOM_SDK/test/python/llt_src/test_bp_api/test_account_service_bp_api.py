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
from collections import namedtuple
from unittest.mock import patch
from flask.testing import FlaskClient


from ut_utils.models import MockPrivilegeAuth
from test_bp_api.create_client import get_client

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from account_service.account_blueprint import account_service_bp

PATCHAccountMember = namedtuple("PATCHAccountMember", "expect, expect_code, data, member_id")


class TestAccountMember:
    client: FlaskClient = get_client(account_service_bp)
    use_cases = {
        "test_patch_account_member": {
            "test Accounts member data_id failed due to member_id is 'x'": (
                {"error": {"code": "Base.1.0.GeneralError",
                           "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                           "@Message.ExtendedInfo": [
                               {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": 100024}}]}},
                400, {"UserName": "123"}, 'x'
            ),
            "test Accounts member data_id failed due to member_id is between 1~16": (
                {"error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The user name or password error.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 110207
                            }
                        }
                    ]
                }},
                400, {"UserName": "admin123"}, "17"
            ),
            "test get accounts user_name false for all numbers": (
                {"error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The user name or password error.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 110207
                            }
                        }
                    ]
                }},
                400, {"UserName": "123"}, "1"
            ),
            "test get accounts user_name false for exceeding 16 numbers": (
                {"error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The user name or password error.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 110207
                            }
                        }
                    ]
                }},
                400, {"UserName": "123qwrweafferaferscs3414"}, "1"
            ),
            "test get accounts user_name false for NULL": (
                {"error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The user name or password error.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 110207
                            }
                        }
                    ]
                }},
                400, {"UserName": ""}, "1"
            ),
            "test get Accounts old_password false for Not more than 8 length": (
                {"error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The user name or password error.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 110207
                            }
                        }
                    ]
                }},
                400, {"old_password": "123456"}, "1"
            ),
            "test get Accounts old_password false for exceeding 20 length": (
                {"error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The user name or password error.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 110207
                            }
                        }
                    ]
                }},
                400, {"old_password": "123456123qwrweafferaferscs3414fwqswd"}, "1"
            ),
            "test get Accounts Password false for Not more than 8 length": (
                {"error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The user name or password error.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 110207
                            }
                        }
                    ]
                }},
                400, {"Password": "34456"}, "1"
            ),
            "test get Accounts Password false for exceeding 20 length.": (
                {"error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The user name or password error.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 110207
                            }
                        }
                    ]
                }},
                400, {"Password": "123456123qwrweafferaferscs3414fweqfwffqq"}, "1"
            ),
            "test get Accounts new_password_second false for Not more than 8 length.": (
                {"error": {
                    "code": "Base.1.0.GeneralError",
                    "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                    "@Message.ExtendedInfo": [
                        {
                            "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The user name or password error.",
                            "Severity": "Critical",
                            "NumberOfArgs": None,
                            "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {
                                "status": 110207
                            }
                        }
                    ]
                }},
                400, {"Password": "Huawei1234", "new_password_second": "Huawei12345"}, "1"
            )
        }
    }

    def test_patch_account_member(self, model: PATCHAccountMember):
        response = self.client.patch("/redfish/v1/AccountService/Accounts/{}".format(model.member_id),
                                     data=json.dumps(model.data))
        assert response.status_code == model.expect_code
        assert response.get_json(force=True) == model.expect
