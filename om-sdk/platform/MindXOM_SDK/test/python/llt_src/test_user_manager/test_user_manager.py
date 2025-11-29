# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import mock
import pytest
from pytest_mock import MockerFixture

from common.constants import error_codes
from common.constants.base_constants import UserManagerConstants
from common.utils.app_common_method import AppCommonMethod
from common.utils.system_utils import SystemUtils
from lib_restful_adapter import LibRESTfulAdapter
from user_manager.user_manager import SessionManager, UserManager, HisPwdManage, UserUtils, EdgeConfigManage
from user_manager.models import User, Session, EdgeConfig
from test_mqtt_api.get_log_info import GetLogInfo
from test_mqtt_api.get_log_info import GetOperationLog

getLog = GetLogInfo()
getOplog = GetOperationLog()


class TestUserManager:
    @staticmethod
    def test_find_user_by_empty_token():
        token = ""
        with pytest.raises(RuntimeError) as exception_info:
            UserManager.find_user_by_token(token)
            assert "token" in str(exception_info.value)

    @staticmethod
    def test_find_user_by_valid_token(mocker: MockerFixture):
        token = "123"
        session = Session()
        mocker.patch.object(SessionManager, "find_session_by_token", return_value=session)
        mocker.patch.object(UserManager, "find_user_by_id", return_value=User())
        user = UserManager.find_user_by_token(token)
        assert session.user_id == user.id

    @staticmethod
    def test_find_user_id_list(mocker: MockerFixture):
        user = User(id=10)
        mocker.patch.object(UserManager, "find_user_list", return_value=[user])
        ret = UserManager.find_user_id_list()
        assert str(user.id) in ret

    @staticmethod
    def test_get_account_service(mocker: MockerFixture):
        edge_config = EdgeConfig(id=10)
        mocker.patch.object(EdgeConfigManage, "find_edge_config", return_value=edge_config)
        ret = UserManager.get_account_service()
        assert edge_config.default_expiration_days == ret

    @staticmethod
    def test_modify_account_service_with_invalid_expiration_day():
        expiration_day = 366
        user_id = "321"
        password = "312"
        with pytest.raises(RuntimeError) as exception_info:
            UserManager.modify_account_service(expiration_day, user_id, password)
            assert "expiration_day" in str(exception_info.value)

    @staticmethod
    def test_modify_account_service_with_invalid_user_id():
        expiration_day = 366
        user_id = ""
        password = "312"
        with pytest.raises(RuntimeError) as exception_info:
            UserManager.modify_account_service(expiration_day, user_id, password)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_modify_account_service_with_invalid_password():
        expiration_day = 366
        user_id = "fdsa"
        password = ""
        with pytest.raises(RuntimeError) as exception_info:
            UserManager.modify_account_service(expiration_day, user_id, password)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_modify_expiration_days(mocker: MockerFixture):
        expiration_day = 10
        edge_config = EdgeConfig(default_expiration_days=expiration_day)
        mocker.patch.object(EdgeConfigManage, "update_expiration_days", return_value=True)
        mocker.patch.object(EdgeConfigManage, "find_edge_config", return_value=edge_config)
        assert expiration_day == UserManager.modify_expiration_days(expiration_day)

    @staticmethod
    def test_get_user_information(mocker: MockerFixture):
        user_id = 1
        user = User()
        password_valid_days = 3
        last_login_ip = "1.12.12.1"
        mocker.patch.object(UserManager, "find_user_by_id", return_value=user)
        mocker.patch.object(UserUtils, "get_password_valid_days", return_value=password_valid_days)
        mocker.patch.object(SessionManager, "get_login_ip_from_db", return_value=last_login_ip)
        user_info = UserManager.get_user_information(user_id)
        assert user_info["Id"] == str(user.id) and \
               user_info["LastLoginSuccessTime"] == user.last_login_success_time and \
               user_info["LastLoginFailureTime"] == user.last_login_failure_time and \
               user_info["AccountInsecurePrompt"] == user.account_insecure_prompt and \
               user_info["ConfigNavigatorPrompt"] == user.config_navigator_prompt and \
               user_info["PasswordValidDays"] == password_valid_days

    @staticmethod
    def test_modify_username_password_with_different_new_password(mocker: MockerFixture):
        user_id = 1
        new_username = "administration"
        old_password = "123"
        new_password = "new password"
        new_password_second = ""
        mocker.patch.object(UserManager, "check_modify_user_password", return_value=user_id)
        mocker.patch.object(UserManager, "find_user_by_id", return_value=User())
        with pytest.raises(RuntimeError) as exception_info:
            UserManager.modify_username_password(user_id, new_username, old_password, new_password, new_password_second)
            assert error_codes.UserManageErrorCodes.ERROR_PASSWORD_TWO_INCONSISTENT.code in str(exception_info.value)

    @staticmethod
    def test_modify_username_password_with_full_space(mocker: MockerFixture):
        user_id = 1
        new_username = "administration"
        old_password = "123"
        new_password = "new password"
        new_password_second = "new password"
        mocker.patch.object(UserManager, "check_modify_user_password", return_value=user_id)
        mocker.patch.object(UserManager, "find_user_by_id", return_value=User())
        mocker.patch.object(SystemUtils, "get_available_size", return_value=0)
        with pytest.raises(RuntimeError) as exception_info:
            UserManager.modify_username_password(user_id, new_username, old_password, new_password, new_password_second)
            assert error_codes.CommonErrorCodes.ERROR_DIR_SPACE_LOW.code in str(exception_info.value)

    @staticmethod
    def test_modify_username_password_with_different_username(mocker: MockerFixture):
        user_id = 1
        new_username = "administration"
        old_password = "123"
        new_password = "new password"
        new_password_second = "new password"
        mocker.patch.object(UserManager, "check_modify_user_password", return_value=user_id)
        mocker.patch.object(UserManager, "find_user_by_id", return_value=User(username_db="admin"))
        mocker.patch.object(SystemUtils, "get_available_size", return_value=1000)
        with pytest.raises(RuntimeError) as exception_info:
            UserManager.modify_username_password(user_id, new_username, old_password, new_password, new_password_second)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_check_modify_user_password_with_empty_new_username():
        user_id = 1
        new_username = ""
        old_password = "123"
        with pytest.raises(RuntimeError) as exception_info:
            UserManager.check_modify_user_password(new_username, old_password, user_id)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_check_modify_user_password_with_empty_old_password():
        user_id = 1
        new_username = "admin"
        old_password = ""
        with pytest.raises(RuntimeError) as exception_info:
            UserManager.check_modify_user_password(new_username, old_password, user_id)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_check_modify_user_password(mocker: MockerFixture):
        user_id = 1
        new_username = "admin"
        old_password = "admin"
        mocker.patch.object(UserUtils, "check_username", return_value=True)
        mocker.patch.object(UserUtils, "unlock_user_locked", return_value=True)
        mocker.patch.object(UserUtils, "verify_pword_and_locked", return_value=True)
        assert user_id == UserManager.check_modify_user_password(new_username, old_password, user_id)

    @staticmethod
    def test_check_user_password_with_empty_password():
        user_id = 1
        password = ""
        with pytest.raises(RuntimeError) as exception_info:
            UserManager.check_user_password(user_id, password)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_fd_modify_passwd_with_wrong_param_dict():
        param_dict = {"UserName": "admin", "Password": "password123", "oper_type": "modify_password"}
        with pytest.raises(RuntimeError) as exception_info:
            UserManager().fd_modify_passwd(param_dict)
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG.code in str(exception_info.value)

    @staticmethod
    def test_get_all_info_by_token(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.USER_INFO_BY_TOKEN, "token": "token"}
        user_id = 1
        username_db = "admin"
        mocker.patch.object(UserManager, "find_user_by_token", return_value=User(id=user_id, username_db=username_db))
        user_manager = UserManager()
        user_manager.get_all_info(param_dict)
        assert user_manager.result["user_id"] == user_id and user_manager.result["user_name"] == username_db

    @staticmethod
    def test_get_all_info_get_account_expiration_day(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_ACCOUNT_EXPIRATION_DAY}
        expiration_day = 100
        mocker.patch.object(UserManager, "get_account_service", return_value=expiration_day)
        user_manager = UserManager()
        user_manager.get_all_info(param_dict)
        assert user_manager.PasswordExpirationDays == expiration_day

    @staticmethod
    def test_get_all_info_get_user_id_list(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_USER_ID_LIST}
        id_list = ["100"]
        mocker.patch.object(UserManager, "find_user_id_list", return_value=id_list)
        user_manager = UserManager()
        user_manager.get_all_info(param_dict)
        assert user_manager.result == id_list

    @staticmethod
    def test_get_all_info_get_user_info(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_USER_INFO, "user_id": 1}
        mocker.patch.object(UserManager, "get_user_information", return_value={"Id": 1})
        user_manager = UserManager()
        user_manager.get_all_info(param_dict)
        assert user_manager.result["Id"] == 1

    @staticmethod
    def test_get_all_info_get_user_list(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_USER_LIST}
        users = ["admin"]
        mocker.patch.object(UserManager, "find_users", return_value=users)
        user_manager = UserManager()
        user_manager.get_all_info(param_dict)
        assert user_manager.result == users

    @staticmethod
    def test_get_all_info_check_password_with_wrong_param_dict(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_CHECK_PASSWORD}
        mocker.patch.object(UserManager, "check_user_password", return_value=False)
        assert UserManager().get_all_info(param_dict)["status"] == AppCommonMethod.ERROR

    @staticmethod
    def test_get_all_info_check_password_pass(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_CHECK_PASSWORD, "password": "password", "user_id": 1}
        mocker.patch.object(UserManager, "check_user_password", return_value=True)
        assert UserManager().get_all_info(param_dict)

    @staticmethod
    def test_get_all_info_check_password_failed(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_CHECK_PASSWORD, "password": "password", "user_id": 1}
        mocker.patch.object(UserManager, "check_user_password", return_value=False)
        assert UserManager().get_all_info(param_dict)["status"] == AppCommonMethod.ERROR

    @staticmethod
    def test_modify_password_expiration_day_from_fd_with_wrong_param_dict():
        with pytest.raises(RuntimeError) as exception_info:
            UserManager().modify_password_expiration_day_from_fd(dict())
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG.code in str(exception_info.value)

    @staticmethod
    def test_modify_password_expiration_day_from_fd_with_invalid_min():
        param_dict = {"PasswordExpirationDays": UserManagerConstants.MIN_PASSWORD_EXPIRATION_DAY - 1}
        with pytest.raises(RuntimeError) as exception_info:
            UserManager().modify_password_expiration_day_from_fd(param_dict)
            assert error_codes.UserManageErrorCodes.ERROR_PARAM_RANGE.code in str(exception_info.value)

    @staticmethod
    def test_modify_password_expiration_day_from_fd_with_invalid_max():
        param_dict = {"PasswordExpirationDays": UserManagerConstants.MAX_PASSWORD_EXPIRATION_DAY + 1}
        with pytest.raises(RuntimeError) as exception_info:
            UserManager().modify_password_expiration_day_from_fd(param_dict)
            assert error_codes.UserManageErrorCodes.ERROR_PARAM_RANGE.code in str(exception_info.value)

    @staticmethod
    def test_modify_user_name_password_with_wrong_param_dict():
        with pytest.raises(RuntimeError) as exception_info:
            UserManager().modify_user_name_password(dict())
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG.code in str(exception_info.value)

    @staticmethod
    def test_modify_user_name_password(mocker: MockerFixture):
        param_dict = {
            "UserName": "admin",
            "Password": "password123",
            "new_password_second": "password123",
            "user_id": 1,
            "old_password": "old_password",
        }
        mocker.patch.object(UserManager, "modify_username_password", return_value=True)
        mocker.patch.object(UserManager, "get_user_information", return_value={"Id": 1})
        mocker.patch.object(SessionManager, "delete_session_by_user_id", return_value=1)
        user_manager = UserManager()
        user_manager.modify_user_name_password(param_dict)
        assert user_manager.result == {"Id": 1}

    @staticmethod
    def test_modify_password_expiration_day_with_wrong_param_dict():
        with pytest.raises(RuntimeError) as exception_info:
            UserManager().modify_password_expiration_day(dict())
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG.code in str(exception_info.value)

    @staticmethod
    def test_modify_password_expiration_day(mocker: MockerFixture):
        param_dict = {"PasswordExpirationDays": "PasswordExpirationDays", "Password": "password123", "user_id": 1}
        mocker.patch.object(UserManager, "modify_account_service", return_value=1)
        user_manager = UserManager()
        user_manager.modify_password_expiration_day(param_dict)
        assert user_manager.PasswordExpirationDays == 1

    @staticmethod
    def test_patch_request_with_wrong_param_dict():
        assert UserManager().patch_request(dict())["status"] == AppCommonMethod.ERROR

    @staticmethod
    def test_patch_request_modify_password_expiration_day(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_MODIFY_ACCOUNT_EXPIRATION_DAY}
        mocker.patch.object(UserManager, "modify_password_expiration_day", return_value=True)
        assert UserManager().patch_request(param_dict)["status"] == AppCommonMethod.OK

    @staticmethod
    def test_patch_request_modify_user_name_password(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_MODIFY_USER_INFO}
        mocker.patch.object(UserManager, "modify_user_name_password", return_value=True)
        assert UserManager().patch_request(param_dict)["status"] == AppCommonMethod.OK

    @staticmethod
    def test_patch_request_fd_modify_passwd(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_MODIFY_PASSWORD}
        mocker.patch.object(UserManager, "fd_modify_passwd", return_value=True)
        assert UserManager().patch_request(param_dict)["status"] == AppCommonMethod.OK

    @staticmethod
    def test_patch_request_modify_password_expiration_day_from_fd(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_FD_MODIFY_ACCOUNT_EXPIRATION_DAY}
        mocker.patch.object(UserManager, "modify_password_expiration_day_from_fd", return_value=True)
        assert UserManager().patch_request(param_dict)["status"] == AppCommonMethod.OK


class TestSessionManager:
    @staticmethod
    def test_get_session_service(mocker: MockerFixture):
        edge_config = EdgeConfig()
        mocker.patch.object(EdgeConfigManage, "find_edge_config", return_value=edge_config)
        mocker.patch.object(edge_config, "token_timeout", return_value=150)
        assert isinstance(SessionManager.get_session_service(), int)

    @staticmethod
    def test_find_session_by_dialog_id_with_wrong_param():
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().find_session_by_dialog_id("")
            assert error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND.code in str(exception_info.value)

    @staticmethod
    def test_find_session_by_dialog_id_with_wrong_max_length():
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().find_session_by_dialog_id("12345678901234567890123456789012345678901234567890")
            assert error_codes.UserManageErrorCodes.ERROR_PARAM_RANGE.code in str(exception_info.value)

    @staticmethod
    def test_create_redfish_session_with_wrong_username():
        username = ""
        password = ""
        real_ip = ""
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().create_redfish_session(username, password, real_ip)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_create_redfish_session_with_wrong_ip(mocker: MockerFixture):
        username = "admin"
        password = ""
        real_ip = ""
        mocker.patch.object(UserUtils, "check_username", check_username=True)
        mocker.patch.object(UserManager, "find_user_by_username", check_username=User(id=1))
        mocker.patch.object(UserUtils, "unlock_user_locked", check_username=True)
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().create_redfish_session(username, password, real_ip)
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_WRONG.code in str(exception_info.value)

    @staticmethod
    def test_get_all_info_with_wrong_param_dict():
        param_dict = dict()
        assert SessionManager().get_all_info(param_dict)["status"] == AppCommonMethod.ERROR

    @staticmethod
    def test_get_all_info_by_token(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_USER_TOKEN}
        mocker.patch.object(SessionManager, "create_session", return_value=True)
        assert SessionManager().get_all_info(param_dict)["status"] == AppCommonMethod.OK

    @staticmethod
    def test_get_all_info_delete_token(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_DELETE_USER_TOKEN}
        mocker.patch.object(SessionManager, "del_session", return_value=True)
        assert SessionManager().get_all_info(param_dict)["status"] == AppCommonMethod.OK

    @staticmethod
    def test_get_all_info_get_session_timeout(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_SESSION_TIMEOUT}
        mocker.patch.object(SessionManager, "get_session_service", return_value=10)
        session_manager = SessionManager()
        assert session_manager.get_all_info(param_dict)["status"] == AppCommonMethod.OK and \
               session_manager.SessionTimeout == 10

    @staticmethod
    def test_create_session_with_wrong_param_dict():
        param_dict = dict()
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().create_session(param_dict)
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG.code in str(exception_info.value)

    @staticmethod
    def test_create_session_with_wrong_ip():
        param_dict = {
            "UserName": "admin",
            "Password": "admin123",
            "real_ip": "",
        }
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().create_session(param_dict)
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_WRONG.code in str(exception_info.value)

    @staticmethod
    def test_create_session_with_socket_path_not_exist(mocker: MockerFixture):
        param_dict = {
            "UserName": "admin",
            "Password": "admin123",
            "real_ip": "1.1.1.1",
        }
        mock_rest_ret = {
            "status": AppCommonMethod.ERROR,
            "message": "Socket path is not exist.",
        }
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=mock_rest_ret)
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().create_session(param_dict)
            assert error_codes.CommonErrorCodes.ERROR_SERVICE_IS_STARTING.code in str(exception_info.value)

    @staticmethod
    def test_create_session_with_service_startup_failed(mocker: MockerFixture):
        param_dict = {
            "UserName": "admin",
            "Password": "admin123",
            "real_ip": "1.1.1.1",
        }
        mock_rest_ret = {
            "status": AppCommonMethod.ERROR,
            "message": "Send message failed.",
        }
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=mock_rest_ret)
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().create_session(param_dict)
            assert error_codes.CommonErrorCodes.ERROR_SERVICE_STARTUP_FAILED.code in str(exception_info.value)

    @staticmethod
    def test_create_session_with_internal_error(mocker: MockerFixture):
        param_dict = {
            "UserName": "admin",
            "Password": "admin123",
            "real_ip": "1.1.1.1",
        }
        mock_rest_ret = {
            "status": AppCommonMethod.ERROR,
            "message": "internal error.",
        }
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=mock_rest_ret)
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().create_session(param_dict)
            assert error_codes.UserManageErrorCodes.ERROR_SECURITY_CFG_NOT_MEET.code in str(exception_info.value)

    @staticmethod
    def test_create_session_ok(mocker: MockerFixture):
        param_dict = {
            "UserName": "admin",
            "Password": "admin123",
            "real_ip": "1.1.1.1",
        }
        mock_rest_ret = {
            "status": AppCommonMethod.OK,
        }
        user_token = "mock token"
        one_user = User(id=1, username_db="admin", account_insecure_prompt=True)
        dialog_id = "1234567890123"
        one_session = Session(dialog_id=dialog_id)
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=mock_rest_ret)
        mocker.patch.object(SessionManager, "create_redfish_session", return_value=user_token)
        mocker.patch.object(UserManager, "find_user_by_username", return_value=one_user)
        mocker.patch.object(SessionManager, "find_session_by_user_id", return_value=one_session)
        session_manager = SessionManager()
        session_manager.create_session(param_dict)
        assert session_manager.result["Id"] == dialog_id and \
               session_manager.result["UserName"] == "admin" and \
               session_manager.result["AccountInsecurePrompt"]

    @staticmethod
    def test_del_session_with_wrong_param_dict():
        param_dict = dict()
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().del_session(param_dict)
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG.code in str(exception_info.value)

    @staticmethod
    def test_del_session_with_wrong_dialog_id():
        param_dict = {"dialog_id": "123465"}
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().del_session(param_dict)
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG.code in str(exception_info.value)

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(SessionManager, "delete_dialog_id", mock.Mock(return_value=True))
    def test_del_session_succeed():
        param_dict = {"dialog_id": "123465123465123465123465123465123465123465123465"}
        SessionManager().del_session(param_dict)
        assert getLog.get_log() is not None

    @staticmethod
    def test_modify_session_service_with_empty_session_timeout():
        session_timeout = None
        user_id = 1
        password = "123"
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().modify_session_service(session_timeout, user_id, password)
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST.code in str(exception_info.value)

    @staticmethod
    def test_modify_session_service_with_wrong_session_timeout():
        session_timeout = "12"
        user_id = 1
        password = "123"
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().modify_session_service(session_timeout, user_id, password)
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_TYPE.code in str(exception_info.value)

    @staticmethod
    def test_modify_session_service_with_wrong_min_session_timeout():
        session_timeout = UserManagerConstants.MIN_SESSION_TIMEOUT - 1
        user_id = 1
        password = "123"
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().modify_session_service(session_timeout, user_id, password)
            assert error_codes.UserManageErrorCodes.ERROR_PARAM_RANGE.code in str(exception_info.value)

    @staticmethod
    def test_modify_session_service_with_wrong_password():
        session_timeout = UserManagerConstants.MIN_SESSION_TIMEOUT - 1
        user_id = 1
        password = ""
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().modify_session_service(session_timeout, user_id, password)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_modify_session_service_succeed(mocker: MockerFixture):
        session_timeout = 10
        user_id = 1
        password = "123"
        mocker.patch.object(UserManager, "check_user_password", return_value=True)
        mocker.patch.object(EdgeConfigManage, "update_token_timeouts", return_value=True)
        mocker.patch.object(EdgeConfigManage, "find_edge_config",
                            return_value=EdgeConfig(token_timeout=session_timeout * 60))
        session_manager = SessionManager()
        session_manager.modify_session_service(session_timeout, user_id, password)
        assert session_manager.SessionTimeout == session_timeout

    @staticmethod
    def test_modify_session_timeout_with_wrong_param_dict():
        param_dict = dict()
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().modify_session_timeout(param_dict)
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG.code in str(exception_info.value)

    @staticmethod
    def test_update_session_timeout_succeed(mocker: MockerFixture):
        session_timeout = 10
        user_id = 1
        password = "123"
        mocker.patch.object(EdgeConfigManage, "update_token_timeouts", return_value=True)
        mocker.patch.object(EdgeConfigManage, "find_edge_config",
                            return_value=EdgeConfig(token_timeout=session_timeout * 60))
        session_manager = SessionManager()
        session_manager.update_session_timeout(session_timeout)
        assert session_manager.SessionTimeout == session_timeout

    @staticmethod
    def test_fd_modify_session_timeout_with_wrong_param_dict():
        param_dict = dict()
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().fd_modify_session_timeout(param_dict)
            assert error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG.code in str(exception_info.value)

    @staticmethod
    def test_fd_modify_session_timeout_with_wrong_min_session_timeout():
        param_dict = {"SessionTimeout": UserManagerConstants.MIN_SESSION_TIMEOUT - 1}
        with pytest.raises(RuntimeError) as exception_info:
            SessionManager().fd_modify_session_timeout(param_dict)
            assert error_codes.UserManageErrorCodes.ERROR_PARAM_RANGE.code in str(exception_info.value)

    @staticmethod
    def test_patch_request_with_wrong_param_dict():
        param_dict = dict()
        assert SessionManager().patch_request(param_dict)["status"] == AppCommonMethod.ERROR

    @staticmethod
    def test_patch_request_with_modify_session_timeout(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_MODIFY_SESSION_TIMEOUT}
        mocker.patch.object(SessionManager, "modify_session_timeout", return_value=True)
        assert SessionManager().patch_request(param_dict)["status"] == AppCommonMethod.OK

    @staticmethod
    def test_patch_request_with_fd_modify_session_timeout(mocker: MockerFixture):
        param_dict = {"oper_type": UserManagerConstants.OPER_TYPE_FD_MODIFY_SESSION_TIMEOUT}
        mocker.patch.object(SessionManager, "fd_modify_session_timeout", return_value=True)
        assert SessionManager().patch_request(param_dict)["status"] == AppCommonMethod.OK

    @staticmethod
    def test_patch_request_with_wrong_oper_type(mocker: MockerFixture):
        param_dict = {"oper_type": "wrong operate type"}
        assert SessionManager().patch_request(param_dict)["status"] == AppCommonMethod.ERROR


class TestUserUtils:
    @staticmethod
    def test_check_username_pattern_with_invalid_username():
        assert not UserUtils.check_username_pattern("123")

    @staticmethod
    def test_check_username_pattern_with_valid_username():
        assert UserUtils.check_username_pattern("1234abc")

    @staticmethod
    def test_check_username_with_wrong_param():
        username = None
        with pytest.raises(RuntimeError) as exception_info:
            UserUtils.check_username(username)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_check_username_failed(mocker: MockerFixture):
        username = None
        mocker.patch.object(UserUtils, "check_username_pattern", return_value=False)
        with pytest.raises(RuntimeError) as exception_info:
            UserUtils.check_username(username)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_check_password_valid_with_invalid_password():
        assert not UserUtils.check_password_valid(b"1234abcd")

    @staticmethod
    def test_check_password_valid_with_valid_password():
        assert UserUtils.check_password_valid("1234abcd")

    @staticmethod
    def test_verify_pword_and_locked_failed(mocker: MockerFixture):
        user_id = 1
        password = "AnyStr"
        mocker.patch.object(UserManager, "find_user_by_id", return_value=User())
        mocker.patch.object(UserUtils, "check_password_valid", return_value=True)
        mocker.patch.object(UserUtils, "check_hash_password", return_value=False)
        mocker.patch.object(UserUtils, "modify_wrong_times", return_value=True)
        with pytest.raises(RuntimeError) as exception_info:
            UserUtils.verify_pword_and_locked(user_id, password)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_modify_wrong_times_failed(mocker: MockerFixture):
        user_id = 1
        mocker.patch.object(UserManager, "find_user_by_id", return_value=User(pword_wrong_times=0))
        mocker.patch.object(EdgeConfigManage, "find_edge_config", return_value=EdgeConfig(default_lock_times=0))
        mocker.patch.object(UserManager, "update_user_specify_column", return_value=True)
        mocker.patch.object(SessionManager, "delete_session_by_user_id", return_value=True)
        with pytest.raises(RuntimeError) as exception_info:
            UserUtils.modify_wrong_times(user_id)
            assert error_codes.UserManageErrorCodes.ERROR_USER_LOCK_STATE.code in str(exception_info.value)

    @staticmethod
    def test_check_password_with_low_complexity():
        with pytest.raises(RuntimeError) as exception_info:
            UserUtils.check_password("1234567", "Password")
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    @getLog.clear_common_log
    @mock.patch.object(UserUtils, "hash_pword", mock.Mock(return_value=True))
    @mock.patch.object(UserManager, "update_user_specify_column", mock.Mock(return_value=True))
    @mock.patch.object(HisPwdManage, "save_his_pwd", mock.Mock(return_value=True))
    def test_modify_password_with_account_prompt_true():
        user_id = 1
        new_password = "123456"
        account_prompt = True
        UserUtils.modify_password(user_id, new_password, account_prompt)
        assert getLog.get_log() is not None

    @staticmethod
    def test_check_hash_password_with_wrong_param():
        pword_hash = None
        password = None
        with pytest.raises(RuntimeError) as exception_info:
            UserUtils.check_hash_password(pword_hash, password)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_judge_password_with_different_new_passwords():
        user_id = 1
        new_username = "admin"
        old_username = "admin"
        new_password = "1234567"
        new_password_second = "01234567"
        with pytest.raises(RuntimeError) as exception_info:
            UserUtils.judge_password(user_id, new_username, old_username, new_password, new_password_second)
            assert error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD.code in str(exception_info.value)

    @staticmethod
    def test_judge_password_with_new_password_same_with_username():
        user_id = 1
        new_username = "admin"
        old_username = "admin"
        new_password = "admin"
        new_password_second = "admin"
        with pytest.raises(RuntimeError) as exception_info:
            UserUtils.judge_password(user_id, new_username, old_username, new_password, new_password_second)
            assert error_codes.UserManageErrorCodes.ERROR_PASSWORD_COMPARED_USERNAME_REVERSAL.code in \
                   str(exception_info.value)

    @staticmethod
    def test_get_password_valid_days_with_account_insecure_prompt_is_true(mocker: MockerFixture):
        account_insecure_prompt = True
        modify_time = ""
        mocker.patch.object(EdgeConfigManage, "find_edge_config", return_value=EdgeConfig())
        assert UserUtils.get_password_valid_days(account_insecure_prompt, modify_time) == "--"

    @staticmethod
    def test_get_password_valid_days_with_invalid_time(mocker: MockerFixture):
        account_insecure_prompt = False
        modify_time = "2070-01-01 00:00:00"
        edge_config = EdgeConfig()
        mocker.patch.object(EdgeConfigManage, "find_edge_config", return_value=edge_config)
        mocker.patch.object(edge_config, "default_expiration_days", return_value=10)
        assert UserUtils.get_password_valid_days(account_insecure_prompt, modify_time) == "-1"

