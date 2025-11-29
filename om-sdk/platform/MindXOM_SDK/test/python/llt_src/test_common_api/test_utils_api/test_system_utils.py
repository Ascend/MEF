from collections import namedtuple

import pytest
from pytest_mock import MockerFixture


from conftest import TestBase
from common.exception.biz_exception import BizException
from common.utils.system_utils import SystemUtils
from common.utils.app_common_method import AppCommonMethod


class TestSystemUtils(TestBase):

    use_cases = {
        "test_get_available_size_exception": {
            "exec_failed": ([-1, ""],),
            "size_type_wrong": ([-1, "a"],)
        }
    }

    def test_get_available_size(self):
        ret = SystemUtils.get_available_size("/usr")
        assert ret > 0

    def test_get_available_size_exception(self, mocker: MockerFixture):
        mocker.patch("os.statvfs", side_effect=BizException(500))
        with pytest.raises(BizException):
            SystemUtils.get_available_size("/usr")
