from collections import namedtuple

from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from lib.Linux.systems.security_service.cert_alarm_time import CertAlarmTime
from common.common_methods import CommonMethods
from ut_utils.mock_utils import mock_write_file_with_os_open

PatchRequest = namedtuple("PatchRequest", "expect, request, locked, link")


class TestCertAlarmTime:
    use_cases = {
        "test_patch_request": {
            "null request": ([1, "Request parameter is empty or not dict"], None, False, None),
            "invalid CertAlarmTime": (
                [CommonMethods.ERROR, "Parameter CertAlarmTime error"], {"CertAlarmTime": None}, False, None
            ),
            "locked": (
                [CommonMethods.ERROR, "CertAlarm modify is busy"], {"CertAlarmTime": 100}, True, None
            ),
            "set failed": (
                [CommonMethods.ERROR, "Setting CertAlarmTime Failed"], {"CertAlarmTime": 100}, False, False
            ),
            "ok": (
                [CommonMethods.OK, "Setting CertAlarmTime success"], {"CertAlarmTime": 100}, False, True
            )
        }
    }

    @staticmethod
    def test_patch_request(mocker: MockerFixture, model: PatchRequest):
        mocker.patch.object(CertAlarmTime, "CERT_ALARM_TIME_MANAGE_LOCK").locked.return_value = model.locked
        mocker.patch.object(FileCheck, "check_is_link", return_value=model.link)
        mock_write_file_with_os_open(mocker)
        assert CertAlarmTime().patch_request(model.request) == model.expect
