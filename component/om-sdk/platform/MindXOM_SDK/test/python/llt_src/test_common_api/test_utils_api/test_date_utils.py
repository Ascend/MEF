from common.utils.date_utils import DateUtils
from conftest import TestBase


class TestDateUtils(TestBase):

    def test_get_format_time(self):
        ret = DateUtils.get_format_time(1000000000.0)
        assert ret

    def test_get_time(self):
        ret = DateUtils.get_time("2022-07-11 15:15:15")
        assert ret
