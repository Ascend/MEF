from unittest import mock

from common.file_utils import FileCheck
from common.utils.result_base import Result
from lib.Linux.systems.security_service.security_service import SecurityService


class TestSecurityService:
    FAILED = 400
    OK = 200

    def test_check_parameter_should_failed_when_parameter_is_none(self):
        ret = SecurityService.cheack_parameter(None)
        assert ret[0] == self.FAILED

    def test_check_parameter_should_failed_when_filename_not_match_regex(self):
        ret = SecurityService.cheack_parameter("~123")
        assert ret[0] == self.FAILED and ret[1] == "Incorrect parameter name"

    def test_check_parameter_should_failed_when_filename_not_end_with_right_suffix(self):
        ret = SecurityService.cheack_parameter("123.p23")
        assert ret[0] == self.FAILED and ret[1] == "Incorrect parameter type"

    def test_check_parameter_should_ok_when_parameter_is_ok(self):
        ret = SecurityService.cheack_parameter("123.crt")
        assert ret[0] == self.OK

    @mock.patch.object(SecurityService, "cheack_parameter")
    @mock.patch.object(FileCheck, "check_path_is_exist_and_valid")
    def test_post_request_should_failed_when_path_failed(self,
                                                         mock_check_path_is_exist_and_valid,
                                                         mock_cheack_parameter):
        mock_check_path_is_exist_and_valid.return_value = Result(False, err_msg="test failed")
        mock_cheack_parameter.return_value = [200, ""]
        request_data = {
            "FileName": "123.p23"
        }
        ret = SecurityService().post_request(request_data)
        assert ret[0] == 400 and ret[1] == "Importing a custom certificate failed"
