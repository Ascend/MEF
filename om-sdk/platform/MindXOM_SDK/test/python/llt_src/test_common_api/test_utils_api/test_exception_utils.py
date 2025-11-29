from collections import namedtuple

from common.constants.error_codes import CommonErrorCodes
from common.exception.biz_exception import BizException
from common.utils.exception_utils import ExceptionUtils
from conftest import TestBase


class TestExceptionUtils(TestBase):
    ExceptionCase = namedtuple("ExceptionCase", "excepted_code, exception")

    use_cases = {
        "test_exception_process": {
            "biz": (100000, BizException(CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST)),
            "not_biz": (100011, IOError())
        }
    }

    def test_exception_process(self, model: ExceptionCase):
        ret = ExceptionUtils.exception_process(model.exception)
        assert model.excepted_code == ret[0]
