from collections import namedtuple

from pytest_mock import MockerFixture

from conftest import TestBase
from lib.Linux.systems.Alarm.alarm import Alarm
from ut_utils.mock_utils import mock_check_input_path_valid
from ut_utils.mock_utils import mock_check_path
from ut_utils.mock_utils import mock_path_exists
from ut_utils.mock_utils import mock_read_data

AlarmCase = namedtuple("AlarmCase", "expect, exists, path_valid, read_data")
MidwareAalarmCase = namedtuple("MidwareAalarmCase", "expect, msg_ret, json_data")
ALARM_FILE = "/run/all_active_alarm"


class TestAlarm(TestBase):
    use_cases = {
        "test_get_all_alarm_dev": {
            "normal": (["0"], True, True, "0@0@0@0@0@aabb"),
            "invalid_path": ([], True, False, ""),
            "invalid_alarm": ([], True, True, "0@0@0@0@aabb")},
        "test_get_file_info": {
            "normal": (
                [{"AlarmId": "0", "AlarmName": "0", "AlarmInstance": "0", "Timestamp": "0", "PerceivedSeverity": "0"}],
                True, True, "0@0@0@0@0@aabb"),
            "invalid_file": ([], True, False, ""),
            "null_file_data": ([], True, True, None)
        }
    }

    def test_get_all_alarm_dev(self, mocker: MockerFixture, model: AlarmCase):
        mock_read_data(mocker, model.read_data)
        mock_check_path(mocker, model.path_valid)
        assert model.expect == Alarm.getallalarmdev()

    def test_get_file_info(self, mocker: MockerFixture, model: AlarmCase):
        mocker.patch.object(Alarm, "get_shield_flag_by_inner_id", return_value=False)
        mock_read_data(mocker, model.read_data)
        mock_check_input_path_valid(mocker, model.path_valid)
        mock_path_exists(mocker, model.exists)
        alarm = Alarm()
        alarm.get_file_info(ALARM_FILE)
        assert model.expect == alarm.AlarMessages
