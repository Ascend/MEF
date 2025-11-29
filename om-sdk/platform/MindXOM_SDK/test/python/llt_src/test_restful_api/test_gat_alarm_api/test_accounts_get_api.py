# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
from unittest.mock import patch

from test_restful_api.test_z_main.restful_test_base import GetTest
from user_manager.user_manager import UserManager


class TestAccounts(GetTest):
    """2.1.1 查询指定用户资源信息"""
    ACCOUNTS_URL = "/redfish/v1/AccountService/Accounts/"

    def __init__(self, expect_ret, code: int, label: str, member_id: str, patch_return):
        self.expect_ret = expect_ret
        self.patch_return = patch_return
        self.patch = None
        send_url = self.ACCOUNTS_URL + member_id
        super().__init__(url=send_url,
                         code=code,
                         label=label)

    def before(self):
        self.patch = patch.object(UserManager, "get_all_info", side_effect=self.patch_return)
        self.patch.start()

    def after(self):
        if self.patch:
            self.patch.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def test_querying_resource_information_of_specified_user():
    TestAccounts(
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
        member_id='x',
        label="test Accounts member data_id failed due to member_id is not number.",
        patch_return=None
    )
    TestAccounts(
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
        member_id='177777777777777777777777',
        label="test Accounts member data_id failed due to member_id is between 1~16",
        patch_return=None
    )
    TestAccounts(
        expect_ret=json.dumps({
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
                    }
        }),
        code=404,
        member_id='',
        label="test Accounts member data_id failed due to member_id is NULL",
        patch_return=None
    )
    TestAccounts(
        expect_ret=json.dumps({"@odata.context": "/redfish/v1/$metadata#AccountService/Accounts/Members/$entity",
                               "@odata.id": "/redfish/v1/AccountService/Accounts/1",
                               "@odata.type": "#ManagerAccount.v1_3_4.ManagerAccount",
                               "Id": "1",
                               "Name": "User Account",
                               "Oem": {
                                    "LastLoginSuccessTime": "test",
                                    "LastLoginFailureTime": "test",
                                    "AccountInsecurePrompt": False,
                                    "ConfigNavigatorPrompt": True,
                                    "PasswordValidDays": "test",
                                    "PwordWrongTimes": 0,
                                    "LastLoginIP": "test"
                                }}),
        code=200,
        member_id='1',
        label="get accounts info success",
        patch_return=[{'status': 200, 'message': {"result": {"Id": "1", "LastLoginSuccessTime": "test",
                                                             "LastLoginFailureTime": "test",
                                                             "AccountInsecurePrompt": False,
                                                             "ConfigNavigatorPrompt": True,
                                                             "PasswordValidDays": "test",
                                                             "PwordWrongTimes": 0,
                                                             "LastLoginIP": "test"
                                                             }}}]
    )
    TestAccounts(
        expect_ret=json.dumps({
            "error": {"code": "Base.1.0.GeneralError",
                      "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                      "@Message.ExtendedInfo": [
                          {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                           "Description": "Indicates that a general error has occurred.",
                           "Message": "Query specified account info failed.",
                           "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                           "Resolution": "None",
                           "Oem": {"status": None}}]}
        }),
        code=400,
        member_id='1',
        label="get accounts info exception",
        patch_return=Exception()
    )


test_querying_resource_information_of_specified_user()
