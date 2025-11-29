# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
from collections import namedtuple
from unittest.mock import patch
from flask.testing import FlaskClient

from pytest_mock import MockerFixture

from lib_restful_adapter import LibRESTfulAdapter
from user_manager.user_manager import UserManager
from ut_utils.models import MockPrivilegeAuth
from account_service import account_views
from test_bp_api.create_client import get_client

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from account_service.account_blueprint import account_service_bp

GetAccountPwdExpirationDays = namedtuple("GetAccountPwdExpirationDays",
                                         "get_resource, check_status_is_ok, code, get_all_info")
ModifyAccountPwdExpirationDay = namedtuple("ModifyAccountPwdExpirationDay",
                                           "headers, lock, expect, data, lib_interface, code, get_resource")
GetAccountInfoCollection = namedtuple("GetAccountInfoCollection", "get_resource, get_all_info, code")
GetSpecifiedAccountInfo = namedtuple("GetSpecifiedAccountInfo", "member_id, get_resource, get_all_info, code")
PATCHAccountMember = namedtuple("PATCHAccountMember", "expect, expect_code, data, member_id")


class TestAccountMember:
    client: FlaskClient = get_client(account_service_bp)
    use_cases = {
        "test_get_account_password_expiration_days": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#AccountService",
                    "@odata.id": "/redfish/v1/AccountService",
                    "@odata.type": "#AccountService.v1_11_0.AccountService",
                    "Id": "AccountService",
                    "Name": "Account Service",
                    "PasswordExpirationDays": None,
                    "Accounts": {
                        "@odata.id": "/redfish/v1/AccountService/Accounts"
                    }
                },
                True,
                200,
                {"status": 200, "message": dict()}
            ),
            "get_all_info_failed": (
                {
                    "@odata.context": "/redfish/v1/$metadata#AccountService",
                    "@odata.id": "/redfish/v1/AccountService",
                    "@odata.type": "#AccountService.v1_11_0.AccountService",
                    "Id": "AccountService",
                    "Name": "Account Service",
                    "PasswordExpirationDays": None,
                    "Accounts": {
                        "@odata.id": "/redfish/v1/AccountService/Accounts"
                    }
                },
                False,
                400,
                {"status": 400, "message": [100011, "Internal server error"]}
            ),
        },
        "test_modify_account_password_expiration_days": {
            "locked": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                True,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Internal server error",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100011
                                },
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Severity": "Critical"
                            }
                        ],
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information."
                    }
                },
                None, None, 500,
                {
                    "error": {
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Internal server error",
                                "NumberOfArgs": None,
                                "Oem": {
                                    "status": 100011
                                },
                                "ParamTypes": None,
                                "Resolution": "None",
                                "Severity": "Critical"
                            }
                        ],
                        "code": "Base.1.0.GeneralError",
                        "message": "A GeneralError has occurred. See ExtendedInfo for more information."
                    }
                },
            ),
            "invalid_param": (
                {"X-Real-Ip": "null", "X-Auth-Token": "abc"},
                False,
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Alarm/AlarmShield",
                    "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield",
                    "@odata.type": "MindXEdgeAlarm.v1_0_0.MindXEdgeAlarm",
                    "AlarmShieldMessages": [],
                    "Decrease": {
                        "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield/Decrease"
                    },
                    "Id": "Alarm Shield",
                    "Increase": {
                        "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield/Increase"
                    },
                    "Name": "Alarm Shield"
                },
                {"PasswordExpirationDays": 10, "Password": 123},
                [{"status": 200, "message": "pass"}], 400,
                {
                    "@odata.context": "/redfish/v1/$metadata#Systems/Alarm/AlarmShield",
                    "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield",
                    "@odata.type": "MindXEdgeAlarm.v1_0_0.MindXEdgeAlarm",
                    "Id": "Alarm Shield",
                    "Name": "Alarm Shield",
                    "AlarmShieldMessages": [],
                    "Increase": {
                        "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield/Increase"
                    },
                    "Decrease": {
                        "@odata.id": "/redfish/v1/Systems/Alarm/AlarmShield/Decrease"
                    }
                },
            ),
        },
        "test_get_account_info_collection": {
            "success": (
                {
                  "@odata.context": "/redfish/v1/$metadata#AccountService/Accounts/$entity",
                  "@odata.id": "/redfish/v1/AccountService/Accounts",
                  "@odata.type": "#ManagerAccountCollection.ManagerAccountCollection",
                  "Name": "Accounts Collection",
                  "Members@odata.count": 0,
                  "Members": [
                    {
                      "@odata.id": "/redfish/v1/AccountService/Accounts/entityID"
                    }
                  ]
                },
                {"status": 200, "message": {"result": ["OK"]}},
                200,
            ),
            "get_all_info_failed": (
                {
                    "@odata.context": "/redfish/v1/$metadata#AccountService/Accounts/$entity",
                    "@odata.id": "/redfish/v1/AccountService/Accounts",
                    "@odata.type": "#ManagerAccountCollection.ManagerAccountCollection",
                    "Name": "Accounts Collection",
                    "Members@odata.count": 0,
                    "Members": [
                        {
                            "@odata.id": "/redfish/v1/AccountService/Accounts/entityID"
                        }
                    ]
                },
                {"status": 400, "message": [100011, "Internal server error"]},
                400,
            ),
        },
        "test_get_specified_account_info": {
            "invalid_member_id": (
                "fdsa", None, None, 400
            ),
            "pass": (
                "1",
                {
                    "@odata.context": "/redfish/v1/$metadata#AccountService/Accounts/Members/$entity",
                    "@odata.id": "/redfish/v1/AccountService/Accounts/1",
                    "@odata.type": "#ManagerAccount.v1_3_4.ManagerAccount",
                    "Id": None,
                    "Name": "User Account",
                    "Oem": {
                        "LastLoginSuccessTime": None,
                        "LastLoginFailureTime": None,
                        "AccountInsecurePrompt": None,
                        "ConfigNavigatorPrompt": None,
                        "PasswordValidDays": None,
                        "PwordWrongTimes": 	None,
                        "LastLoginIP": None
                    }
                },
                {"status": 200, "message": dict()}, 200
            ),
        },
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

    def test_get_account_password_expiration_days(self, mocker: MockerFixture, model: GetAccountPwdExpirationDays):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        mocker.patch.object(UserManager, "get_all_info", return_value=model.get_all_info)
        response = self.client.get("/redfish/v1/AccountService")
        assert response.status_code == model.code

    def test_modify_account_password_expiration_days(self, mocker: MockerFixture, model: ModifyAccountPwdExpirationDay):
        mocker.patch.object(account_views, "ACCOUNT_SERVICE_LOCK").locked.return_value = model.lock
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        mocker.patch.object(UserManager, "patch_request", side_effect=model.lib_interface)
        response = self.client.patch("/redfish/v1/AccountService", data=json.dumps(model.data), headers=model.headers)
        assert response.status_code == model.code
        assert response.get_json(force=True) == model.expect

    def test_get_account_info_collection(self, mocker: MockerFixture, model: GetAccountInfoCollection):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        mocker.patch.object(UserManager, "get_all_info", return_value=model.get_all_info)
        response = self.client.get("/redfish/v1/AccountService/Accounts")
        assert response.status_code == model.code

    def test_get_specified_account_info(self, mocker: MockerFixture, model: GetSpecifiedAccountInfo):
        mocker.patch.object(json, "loads", return_value=model.get_resource)
        mocker.patch.object(UserManager, "get_all_info", return_value=model.get_all_info)
        response = self.client.get(f"/redfish/v1/AccountService/Accounts/{model.member_id}")
        assert response.status_code == model.code

    def test_patch_account_member(self, model: PATCHAccountMember):
        response = self.client.patch("/redfish/v1/AccountService/Accounts/{}".format(model.member_id),
                                     data=json.dumps(model.data))
        assert response.status_code == model.expect_code
        assert response.get_json(force=True) == model.expect
