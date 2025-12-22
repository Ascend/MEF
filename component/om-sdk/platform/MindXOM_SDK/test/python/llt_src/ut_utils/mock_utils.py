from functools import partial
from typing import Any, List
from unittest.mock import patch, Mock, mock_open

from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.utils.result_base import Result
from ut_utils.models import FakeCDLL

mock_cdll = partial(patch, "ctypes.CDLL", Mock(return_value=FakeCDLL()))
mock_npu_smi_board_type = partial(patch, "bin.monitor_config.SystemSetting.get_board_type")
mock_npu_smi_npu_version = partial(patch, "bin.monitor_config.SystemSetting.get_npu_version")


def mock_check_path(mocker: MockerFixture, path_valid: bool):
    mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", Mock(return_value=Result(path_valid)))


def mock_check_input_path_valid(mocker: MockerFixture, return_value: bool):
    mocker.patch.object(FileCheck, "check_input_path_valid", return_value=Result(return_value))


def mock_check_is_link(mocker: MockerFixture, return_value: bool):
    mocker.patch.object(FileCheck, "check_is_link", return_value=return_value)


def mock_read_data(mocker: MockerFixture, read_data: Any):
    mocker.patch("builtins.open", mock_open(read_data=read_data))


def mock_json_loads(mocker: MockerFixture, json_data: Any):
    mocker.patch("json.loads", return_value=json_data)


def mock_path_exists(mocker: MockerFixture, return_value: bool = None, side_effect: List[bool] = None):
    side_effect = side_effect or [return_value, ]
    mocker.patch("os.path.exists", side_effect=side_effect)


def mock_write_file_with_os_open(mocker: MockerFixture):
    mocker.patch("os.open")
    mocker.patch("os.fdopen")


def mock_os_sync(mocker: MockerFixture):
    mocker.patch("os.sync")


def mock_time_sleep(mocker: MockerFixture):
    mocker.patch("time.sleep")
