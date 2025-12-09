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
from unittest.mock import patch

from test_restful_api.test_z_main.restful_test_base import PostTest

from user_manager.user_manager import SessionManager, UserManager


class TestPostSession(PostTest):
    ACCOUNTS_URL = "/redfish/v1/SessionService/Sessions"

    def __init__(self, expect_ret, data, code: int, label: str, patch_session_return, patch_user_return):
        self.expect_ret = expect_ret
        self.patch_session = None
        self.patch_user = None
        self.patch_session_return = patch_session_return
        self.patch_user_return = patch_user_return
        super().__init__(url=self.ACCOUNTS_URL,
                         code=code,
                         label=label,
                         data=data)

    def before(self):
        self.patch_session = patch.object(SessionManager, "get_all_info", return_value=self.patch_session_return)
        self.patch_session.start()
        self.patch_user = patch.object(UserManager, "find_user_by_username", side_effect=self.patch_user_return)
        self.patch_user.start()

    def after(self):
        if self.patch_session:
            self.patch_session.stop()

        if self.patch_user:
            self.patch_user.stop()

    def call_back_assert(self, test_response: str):
        assert self.expect_ret == test_response


def init_post_session():
    TestPostSession(
        expect_ret=json.dumps(
            {"@odata.context": "/redfish/v1/$metadata#Session.Session",
             "@odata.id": "/redfish/v1/SessionService/Sessions/1",
             "@odata.type": "#Session.v1_4_0.Session", "Id": "1", "Name": "User Session",
             "UserName": "test1", "Oem": {"UserId": "1", "AccountInsecurePrompt": False, "message": "[id:name]"}}),
        code=201,
        data={"UserName": "123a123", "Password": "test"},
        label="Login successfully.",
        patch_session_return={"status": 200, "message": {"result": {"Id": "1", "Token": "test"}}},
        patch_user_return={"status": 200, "message": {"result": {"Id": "1", "name": "test"}}}),


    TestPostSession(
        expect_ret=json.dumps(
            {"error": {"code": "Base.1.0.GeneralError",
                       "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                       "@Message.ExtendedInfo": [
                           {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                            "Description": "Indicates that a general error has occurred.",
                            "Message": "The user name or password error.",
                            "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                            "Resolution": "None",
                            "Oem": {"status": 110207}}]}}),
        code=400,
        data={"UserName": "123", "Password": "test"},
        label="Login failed.",
        patch_session_return={"status": 110210, "message": [110210, "Login failed."]},
        patch_user_return={"status": 200, "message": {"result": {"Id": "1", "name": "test"}}}),


init_post_session()
