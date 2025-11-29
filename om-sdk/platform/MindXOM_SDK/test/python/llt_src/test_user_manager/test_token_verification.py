# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import time
from collections import namedtuple

import pytest
from pytest_mock import MockerFixture

from common.constants import error_codes
from user_manager.models import User, Session, HisPwd, EdgeConfig, LastLoginInfo
from user_manager.token_verification import TokenVerification
from user_manager.user_manager import EdgeConfigManage
from user_manager.user_manager import UserManager
from user_manager.user_manager import SessionManager
from user_manager.user_manager import UserUtils


JudgeToken = namedtuple("JudgeToken", "request_ip, token, auto_refresh, check_and_get_session,"
                                      "find_edge_config, find_user_by_id, perf_counter, get_password_valid_days")
CheckAndGetSession = namedtuple("CheckAndGetSession", "auto_refresh, request_ip, token, exception_code, "
                                                      "find_session_by_user_id, check_hash_password")
GetAllInfo = namedtuple("GetAllInfo", "param_dict, judge_token, status")


class TestDefaultConfig:
    use_cases = {
        "test_judge_token": {
            "timeout": (
                "1.2.3.4", "xxx", "true",
                Session(user_id=1, reset_time="1"),
                EdgeConfig(token_timeout=1),
                User(),
                float(6000),
                0,
            ),
            "password_invalid": (
                "1.2.3.4", "xxx", "false",
                Session(user_id=1, reset_time="1"),
                EdgeConfig(token_timeout=1),
                User(),
                float(0),
                0,
            ),
        },
        "test_check_and_get_session": {
            "invalid_auto_refresh": (
                False, None, None, error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_WRONG.code, None, None,
            ),
            "invalid_token": (
                "false", None, None, error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND.code, None, None,
            ),
            "invalid_user_session": (
                "false", None, "jfakfd", error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND.code, None, None,
            ),
            "check_hash_password_failed": (
                "false", None, "jfakfd", error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND.code,
                Session(token=""), False,
            ),
            "invalid_request_ip": (
                "false", None, "jfakfd", error_codes.UserManageErrorCodes.ERROR_REQUEST_IP_ADDR.code,
                Session(token="", request_ip="xx"), True,
            ),
        },
        "test_get_all_info": {
            "exception": (
                {
                    "token": "",
                    "auto_refresh": "",
                    "request_ip": "",
                },
                Exception,
                400,
            ),
            "success": (
                {
                    "token": "",
                    "auto_refresh": "",
                    "request_ip": "",
                },
                ("user_name", "user_id", "account_insecure_prompt"),
                200,
            )
        },
    }

    @staticmethod
    def test_judge_token(mocker: MockerFixture, model: JudgeToken):
        with pytest.raises(Exception) as exception_info:
            mocker.patch.object(TokenVerification, "check_and_get_session", return_value=model.check_and_get_session)
            mocker.patch.object(EdgeConfigManage, "find_edge_config", return_value=model.find_edge_config)
            mocker.patch.object(UserManager, "find_user_by_id", return_value=model.find_user_by_id)
            mocker.patch.object(time, "perf_counter", return_value=model.perf_counter)
            mocker.patch.object(SessionManager, "delete_session_by_user_id", return_value=True)
            mocker.patch.object(SessionManager, "update_session", return_value=True)
            mocker.patch.object(UserUtils, "get_password_valid_days", return_value=model.get_password_valid_days)
            mocker.patch.object(UserManager, "update_user_specify_column", return_value=True)
            time_variance = model.perf_counter - float(model.check_and_get_session.reset_time)
            token_timeout = model.find_edge_config.token_timeout
            TokenVerification.judge_token(model.request_ip, model.token, model.auto_refresh)
            if time_variance > token_timeout:
                assert time_variance in str(exception_info.value)
            if model.get_password_valid_days != "--" and int(model.get_password_valid_days) <= 0:
                assert model.get_password_valid_days in str(exception_info.value)

    @staticmethod
    def test_check_and_get_session(mocker: MockerFixture, model: CheckAndGetSession):
        with pytest.raises(Exception) as exception_info:
            mocker.patch.object(SessionManager, "find_session_by_user_id", return_value=model.find_session_by_user_id)
            mocker.patch.object(UserUtils, "check_hash_password", return_value=model.check_hash_password)
            TokenVerification.check_and_get_session(model.auto_refresh, model.request_ip, model.token)
            assert model.exception_code in str(exception_info.value)

    @staticmethod
    def test_get_all_info(mocker: MockerFixture, model: GetAllInfo):
        mocker.patch.object(TokenVerification, "judge_token", return_value=model.judge_token)
        ret = TokenVerification().get_all_info(model.param_dict)
        assert ret["status"] == model.status
