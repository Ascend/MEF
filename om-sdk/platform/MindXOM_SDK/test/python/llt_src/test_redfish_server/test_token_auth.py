from collections import namedtuple

import pytest
from pytest_mock import MockerFixture

from token_auth import get_privilege_auth
from user_manager.user_manager import UserManager
# 获取校验类

PrivilegeAuth = get_privilege_auth()
CheckAccountInsecureUrlCase = namedtuple("CheckAccountInsecureUrlCase",
                                         "excepted, ret_data, request_uri, request_method")
GetUserInfoCase = namedtuple("GetUserInfoCase", "excepted, user_info")
AddOperLogCase = namedtuple("AddOperLogCase", "user_info")


class TestPrivilegeAuth:
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
        }
    }

    def test_check_account_insecure_url(self, model: CheckAccountInsecureUrlCase):
        ret = PrivilegeAuth.check_account_insecure_url(model.ret_data, model.request_uri, model.request_method)
        assert model.excepted == ret

    def test_user_make_response_exception(self):
        with pytest.raises(Exception):
            PrivilegeAuth.user_make_response([])

    def test_user_make_response(self, mocker: MockerFixture):
        mocker.patch("token_auth.make_response", return_value={"message": "success", "status": 200})
        ret = PrivilegeAuth.user_make_response([{"message": "success", "status": 200}, 200])
        assert ret == {"message": "success", "status": 200}

    def test_get_user_info(self, mocker: MockerFixture, model: GetUserInfoCase):
        mocker.patch.object(UserManager, "get_all_info", side_effect=model.user_info)
        ret = PrivilegeAuth.get_user_info("test")
        assert model.excepted == ret

    def test_add_oper_logger(self, mocker: MockerFixture, model: AddOperLogCase):
        mocker.patch.object(PrivilegeAuth, "get_user_info", return_value=model.user_info)
        ret = PrivilegeAuth.add_oper_logger("test")
        assert ret is None
