import json
from collections import namedtuple

import flask
import pytest
from flask import request
from pytest_mock import MockerFixture

from common.checkers import CheckResult
from common.checkers import IPChecker
from token_auth import get_privilege_auth
from user_manager.token_verification import TokenVerification
from user_manager.user_manager import UserManager

# 获取校验类

PrivilegeAuth = get_privilege_auth()
CheckAccountInsecureUrlCase = namedtuple("CheckAccountInsecureUrlCase",
                                         "excepted, ret_data, request_uri, request_method")
GetUserInfoCase = namedtuple("GetUserInfoCase", "excepted, user_info")
AddOperLogCase = namedtuple("AddOperLogCase", "user_info")


class TestPrivilegeAuth:
    TokenRequiredCase = namedtuple("TokenRequiredCase",
                                   ["expect", "uri", "method", "ip_check", "headers", "token_info"])
    use_cases = {
        "test_check_account_insecure_url": {
            "account_insecure_null": (True, {"message": {"account_insecure_prompt": False}}, "", ""),
            "request_uri_wrong": (False, {"message": {"account_insecure_prompt": True}}, "PATCH",
                                  "/redfish/v1/AccountService/Accounts"),
            "message_not_dict": (False, {"message": "test"}, "", "")
        },
        "test_get_user_info": {
            "get_success": ({"info": "test"}, [{"status": 200, "message": {"result": {"info": "test"}}}]),
            "get_failed": ({}, [{"status": 500, "message": {"result": {"info": "test"}}}]),
            "exception": ({}, Exception())
        },
        "test_add_oper_logger": {
            "user_name_null": ({"user_name": ""},),
        },
        "test_token_required": {
            "token is none": (
                ("Base.1.0.GeneralError", "token is None"), "/redfish/v1/AccountService/Accounts/1", "PATCH",
                CheckResult.make_success(), {"X-Real-Ip": "10.10.10.10"}, {}),
            "request_ip_is_None": (
                ("Base.1.0.GeneralError", "request ip is None"), "/redfish/v1/AccountService/Accounts/1", "PATCH",
                CheckResult.make_success(), {"X-Real-Ip": "null", "X-Auth-Token": "abc"}, {}),
            "session_not_found": (
                ("Base.1.0.AccountForSessionNoLongerExists", "The account for the current session has been removed, "
                                                             "thus the current session has been removed as well."),
                "/redfish/v1/AccountService/Accounts/1", "PATCH", CheckResult.make_success(),
                {"X-Real-Ip": "10.10.10.10", "X-Auth-Token": "abc"},
                {'status': 400, 'message': [110201, 'Session not found.']}),
            "insecure_account_forbidden": (
                ("Base.1.0.GeneralError", "can not operate because of insecure account"),
                "/redfish/v1/AccountService/Acc12xxxx", "PATCH", CheckResult.make_success(),
                {"X-Real-Ip": "10.10.10.10", "X-Auth-Token": "abc"},
                {'status': 200, 'message': {'user_name': 'admin', 'user_id': 1, 'account_insecure_prompt': True}}),
            "insecure_account_pass": (
                ("/redfish/v1/AccountService/Accounts/1", "User Account"),
                "/redfish/v1/AccountService/Accounts/1", "PATCH", CheckResult.make_success(),
                {"X-Real-Ip": "10.10.10.10", "X-Auth-Token": "abc"},
                {'status': 200, 'message': {'user_name': 'admin', 'user_id': 1, 'account_insecure_prompt': True}}),
            "valid_account_ok": (
                ("/redfish/v1/AccountService/Accounts/1", "User Account"),
                "/redfish/v1/AccountService/Accounts/1", "PATCH", CheckResult.make_success(),
                {"X-Real-Ip": "10.10.10.10", "X-Auth-Token": "abc"},
                {'status': 200, 'message': {'user_name': 'admin', 'user_id': 1, 'account_insecure_prompt': False}}),
        }
    }

    @staticmethod
    def test_check_account_insecure_url(model: CheckAccountInsecureUrlCase):
        ret = PrivilegeAuth.check_account_insecure_url(model.ret_data, model.request_uri, model.request_method)
        assert model.excepted == ret

    @staticmethod
    def test_user_make_response_exception():
        with pytest.raises(Exception):
            PrivilegeAuth.user_make_response([])

    @staticmethod
    def test_user_make_response(mocker: MockerFixture):
        mocker.patch("token_auth.make_response", return_value={"message": "success", "status": 200})
        ret = PrivilegeAuth.user_make_response([{"message": "success", "status": 200}, 200])
        assert ret == {"message": "success", "status": 200}

    @staticmethod
    def test_get_user_info(mocker: MockerFixture, model: GetUserInfoCase):
        mocker.patch.object(UserManager, "get_all_info", side_effect=model.user_info)
        ret = PrivilegeAuth.get_user_info("test")
        assert model.excepted == ret

    @staticmethod
    def test_add_oper_logger(mocker: MockerFixture, model: AddOperLogCase):
        mocker.patch.object(PrivilegeAuth, "get_user_info", return_value=model.user_info)
        ret = PrivilegeAuth.add_oper_logger("test")
        assert ret is None

    @staticmethod
    def test_token_required(mocker: MockerFixture, model: TokenRequiredCase):
        def target_func():
            resp = {
                "@odata.context": "/redfish/v1/$metadata#AccountService/Accounts/Members/$entity",
                "@odata.id": "/redfish/v1/AccountService/Accounts/1",
                "@odata.type": "#ManagerAccount.v1_3_4.ManagerAccount",
                "Id": "1",
                "Name": "User Account",
                "Oem": {
                    "LastLoginSuccessTime": "2022-11-16 20:54:03",
                    "LastLoginFailureTime": "2022-11-16 06:15:29",
                    "AccountInsecurePrompt": False,
                    "ConfigNavigatorPrompt": True,
                    "PasswordValidDays": "--",
                    "PwordWrongTimes": 0,
                    "LastLoginIP": "127.0.xx.xx"
                }
            }
            resp_tuple = namedtuple("resp_tuple", ["data"])
            return resp_tuple(data=json.dumps(resp).encode())

        flask_app = flask.Flask(__name__)
        mocker.patch.object(IPChecker, "check", return_value=model.ip_check)
        mocker.patch.object(TokenVerification, "get_all_info", return_value=model.token_info)
        with flask_app.test_request_context(path=model.uri, method=model.method, headers=model.headers):
            wrapper_func = PrivilegeAuth.token_required(target_func)
            assert wrapper_func is not None
            ans = wrapper_func()
            result = json.loads(ans.data.decode())
            error_info = result.get("error")
            if error_info:
                assert error_info.get("code") == model.expect[0]
                assert error_info.get("@Message.ExtendedInfo", [{}])[0].get("Message") == model.expect[1]
            else:
                assert result.get("@odata.id") == model.expect[0]
                assert result.get("Name") == model.expect[1]
