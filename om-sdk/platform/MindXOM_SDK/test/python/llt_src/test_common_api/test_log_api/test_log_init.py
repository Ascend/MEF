from collections import namedtuple

from pytest_mock import MockerFixture

from common.file_utils import FileCreate
from common.log.log_init import init_om_logger, OMLogger
from conftest import TestBase

CreateCase = namedtuple("CreateCase", "expected, create_dir, exists, realpath, op_type")


class TestLogInit(TestBase):
    use_cases = {
        "test_init_logger": {
            "create_dir_false": (False, False, False, "/var/plog/om_log", "run"),
            "type_wrong": (False, True, False, "/var/plog/om_log", "test"),
        },
    }

    def test_init_om_logger(self):
        ret = init_om_logger("monitor", "/home/log", True)
        assert ret is None

    def test_init_logger(self, mocker: MockerFixture, model: CreateCase):
        # 执行成功用例待补充
        mocker.patch("os.path.realpath", side_effect=model.realpath)
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch.object(FileCreate, "create_file")
        ret = OMLogger().init_logger("monitor")
        assert model.expected == ret
