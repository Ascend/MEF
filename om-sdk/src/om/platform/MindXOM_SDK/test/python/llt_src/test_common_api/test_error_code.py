from common.error_code import ErrorCode


class TestErrorCode:

    def test_generate_err_msg(self):
        ret = ErrorCode.generate_err_msg("166")
        assert ret == "ERR.166, Upgrade script parameter error."

    def test_generate_err_msg_add_body(self):
        ret = ErrorCode.generate_err_msg_add_body("166", "test")
        assert ret == "ERR.166, Upgrade script parameter error.test"
