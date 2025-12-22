from collections import namedtuple


from common.checkers.fd_param_checker import string_compare
from common.checkers.param_checker import pword_compare
from conftest import TestBase


class TestParamCompare(TestBase):
    CompareStrCase = namedtuple("CompareStrCase", "excepted, str1, str2")

    use_cases = {
        "test_string_compare": {
            "string_compare_same": (False, "1", "1"),
            "string_compare_not_same": (True, "1", "2")
        },
        "test_pword_compare": {
            "pword_compare_same": (True, "1", "1"),
            "pword_compare_not_same": (False, "1", "2")
        },
    }

    def test_string_compare(self, model: CompareStrCase):
        ret = string_compare(model.str1, model.str2)
        assert model.excepted == ret

    def test_pword_compare(self, model: CompareStrCase):
        ret = pword_compare(model.str1, model.str2)
        assert model.excepted == ret
