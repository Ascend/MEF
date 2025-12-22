from common.constants.error_code_constants import ErrorCode
from conftest import TestBase


class TestErrorCode(TestBase):

    def test_midware_generate_err_msg(self):
        ret = ErrorCode.midware_generate_err_msg(640, "failed")
        assert ret == "ERR.640, failed"
