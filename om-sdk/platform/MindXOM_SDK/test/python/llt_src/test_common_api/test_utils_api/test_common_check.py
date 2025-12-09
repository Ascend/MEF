from collections import namedtuple
from unittest.mock import mock_open

from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.utils.result_base import Result

from conftest import TestBase
from common.utils.common_check import CommonCheck, get_error_resource, TokenCheck

CheckParamCase = namedtuple("CheckPasswordCase", "excepted, param")
CheckPatternCase = namedtuple("CheckPatternCase", "excepted, str_buffer, pattern")
GetErrorResourceCase = namedtuple("GetErrorResourceCase", "excepted, check_path, read, path")
TokenCheckCase = namedtuple("TokenCheckCase", "excepted, message")
UpdateJsonOfErrorInfoCase = namedtuple("UpdateJsonOfErrorInfoCase", "excepted, ret_dict")

index_file_path_info, _ = get_error_resource(TokenCheck.index_file_path)
index_file_path_resource_info, _ = get_error_resource(TokenCheck.index_file_path_resource)


class TestCommonCheck(TestBase):
    use_cases = {
        "test_check_all_param_not_empty": {
            "parm_list_item_null": (False, [None]),
            "parm_list_dict_item_null": (False, {"a": None}),
            "parm_list_success": (True, {"a": 1})
        },
        "test_check_check_code": {
            "match": (True, "abcd" * 16),
            "not_match": (False, "Edge123456")
        },
        "test_get_error_resource": {
            "check_path_failed": ('{"err_message": "Illegal request"}',
                                  Result(result=False, err_msg="path is not exists"), "", "/home/a.log"),
            "normal": ({'err_message': 'param valid'}, Result(result=True), '{"err_message": "param valid"}',
                       "/home/a.log"),
        },
        "test_token_check": {
            "message_not_list": ([400, "PropertyMissing"], ""),
            "message_not_int": ([400, "PropertyMissing"], ["a", ""]),
            "ERROR_SESSION_NOT_FOUND": ([400, "AccountForSessionNoLongerExists"], [110201, ""]),
            "ERROR_REQUEST_IP_ADDR": ([400, "AccountForSessionNoLongerExists"], [110205, ""]),
            "ERROR_SESSION_TIME_LIMIT": ([400, "SessionTimeLimitExceeded"], [110202, ""]),
            "ERROR_PASSWORD_VALID_DAY": ([400, "PasswordValidDaysError"], [110204, ""]),
            "other": ([400, "X-Auth-Token"], [110206, ""]),
        },
        "test_update_json_of_error_info": {
            "ret_dict_null": (500, None),
            "ret_dict_not_dict": (500, ["a", ""]),
            "ret_dict_without_status": (500, {"a": "1"}),
            "ret_dict_without_message": (500, {"status": "1"}),
            "ret_dict_status_200": (500, {"status", 200}),
            "normal": (400, {"status": 400, "message": "AccountForSessionNoLongerExists", "errorKey": "123",
                             "errorValue": "test", "ParamTypes": "test", "NumberOfArgs": 1}, ),
        }
    }

    def test_check_all_param_not_empty(self, model: CheckParamCase):
        ret = CommonCheck.check_all_param_not_empty(model.param)
        assert model.excepted == ret

    def test_check_sub_list(self):
        ret = CommonCheck.check_sub_list(["1", "2", "3"], ["1", "2"])
        assert ret

    def test_check_check_code(self, model: CheckParamCase):
        ret = CommonCheck.check_check_code(model.param)
        assert model.excepted == ret

    def test_get_error_resource(self, mocker: MockerFixture, model: GetErrorResourceCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch("builtins.open", mock_open(read_data=model.read))
        ret = get_error_resource(model.path)
        assert model.excepted == ret[0]

    def test_token_check(self, model: TokenCheckCase):
        ret = TokenCheck.token_check(model.message)
        assert model.excepted[0] == ret[0]
        assert model.excepted[1] in ret[1]

    def test_update_json_of_error_info(self, model: UpdateJsonOfErrorInfoCase):
        ret = TokenCheck.update_json_of_error_info(index_file_path_info, index_file_path_resource_info,
                                                   model.ret_dict, "AccountForSessionNoLongerExists")
        assert model.excepted == ret[0]


