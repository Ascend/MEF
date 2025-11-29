from collections import namedtuple

from common.constants.error_codes import CommonErrorCodes
from common.exception.biz_exception import Exceptions
from conftest import TestBase


class TestExceptions(TestBase):
    CheckListCase = namedtuple("CheckListCase", "expected, list_checked")
    BizExceptionCase = namedtuple("CheckListCase", "expected, error_code, arg")

    use_cases = {
        "test_check_list_is_null": {
            "param_null": (True, tuple()),
            "param_not_null": (False, (1, None,)),
            "param_item_null": (True, (None,))
        },
        "test_biz_exception": {
            "args_is_not_null": ("Error argument [(123,)] empty",
                                 CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST, (123,)),
            "args_is_null": ("Error argument empty", CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST, None),
        },
    }

    def test_check_list_is_null(self, model: CheckListCase):
        ret = Exceptions.check_list_is_null(model.list_checked)
        assert model.expected == ret

    def test_biz_exception(self, model: BizExceptionCase):
        ret = Exceptions.biz_exception(model.error_code, model.arg)
        assert model.expected == ret.args[0].messageKey
