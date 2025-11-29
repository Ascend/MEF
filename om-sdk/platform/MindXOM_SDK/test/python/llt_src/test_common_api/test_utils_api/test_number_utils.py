from collections import namedtuple

import pytest

from common.exception.biz_exception import BizException
from common.utils.number_utils import NumberUtils
from conftest import TestBase


class TestNumberUtils(TestBase):
    InRangeCase = namedtuple("InRangeCase", "excepted, target_param, start_param, end_param, prompt")
    IsIntCase = namedtuple("IsIntCase", "excepted, parma")
    IsIntExceptionCase = namedtuple("IsIntCase", "parma")

    use_cases = {
        "test_in_range": {
            "in": (True, 10, 1, 20, ""),
            "out": (False, 10, 1, 8, ""),
        },
        "test_is_int": {
            "int": (10, 10),
            "not_int": (10, "10"),
        },
        "test_is_exception": {
            "null": (None,),
            "not_isdigit": ("10a",),
            "len_wrong": ("10" * 123,),
        }
    }

    def test_is_int_maxsize_normal(self):
        ret = NumberUtils.is_int_maxsize(10)
        assert ret == 10

    def test_is_int_maxsize_exception(self):
        with pytest.raises(BizException):
            NumberUtils.is_int_maxsize("10" * 10)

    def test_in_range(self, model: InRangeCase):
        ret = NumberUtils.in_range(model.target_param, model.start_param, model.end_param, model.prompt)
        assert model.excepted == ret

    def test_is_int(self, model: IsIntCase):
        ret = NumberUtils.is_int_maxsize(model.parma)
        assert model.excepted == ret

    def test_is_exception(self, model: IsIntExceptionCase):
        with pytest.raises(BizException):
            NumberUtils.is_int_maxsize(model.parma)
